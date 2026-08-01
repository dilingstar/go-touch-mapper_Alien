package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/akamensky/argparse"
	"github.com/kenshaw/evdev"
	"go.bug.st/serial"
)

var go_build_version string = "star fork v3.5.2 Q群： 1067729485"

var uinput_keyboard_mouse_dev_name = ""

type event_pack struct {
	dev_name string
	dev_type dev_type
	events   []*evdev.Event
}

type touch_control_pack struct {
	action   int8
	id       int32
	x        int32
	y        int32
	screen_x int32
	screen_y int32
}

type u_input_control_pack struct {
	action int8
	arg1   int32
	arg2   int32
}

type touch_control_func func(data touch_control_pack)

func dev_reader(event_reader chan *event_pack, index int) {
	fd, err := os.OpenFile(fmt.Sprintf("/dev/input/event%d", index), os.O_RDONLY, 0)
	if err != nil {
		logger.Errorf("读取设备失败 : %v", err)
		return
	}
	d := evdev.Open(fd)
	defer d.Close()
	event_ch := d.Poll(context.Background())
	events := make([]*evdev.Event, 0)
	dev_name := d.Name()
	dev_type := check_dev_type(d, fd)

	if dev_type == type_motion_sensors {
		global_motion_sensors_range[dev_name] = map[uint16][]int32{}
		for k, v := range d.AbsoluteTypes() {
			global_motion_sensors_range[dev_name][uint16(k)] = []int32{v.Min, v.Max}
		}
		defer delete(global_motion_sensors_range, dev_name)
	}

	logger.Infof("开始读取设备 : %s", dev_name)
	d.Lock()
	defer d.Unlock()
	for {
		select {
		case <-global_close_signal:
			logger.Infof("释放设备 : %s", dev_name)
			return
		case event := <-event_ch:
			if event == nil {
				logger.Warnf("移除设备 : %s", dev_name)
				return
			} else if event.Type == evdev.SyncReport {
				pack := &event_pack{
					dev_name: dev_name,
					dev_type: dev_type,
					events:   events,
				}
				event_reader <- pack
				events = make([]*evdev.Event, 0)
			} else {
				events = append(events, &event.Event)
			}
		}
	}
}

func touch_dev_reader(event_reader chan *event_pack, index int) {
	fd, err := os.OpenFile(fmt.Sprintf("/dev/input/event%d", index), os.O_RDONLY, 0)
	if err != nil {
		logger.Errorf("读取设备失败 : %v", err)
		return
	}
	d := evdev.Open(fd)
	defer d.Close()
	event_ch := d.Poll(context.Background())
	events := make([]*evdev.Event, 0)
	dev_name := d.Name()
	abs_max_x := d.AbsoluteTypes()[evdev.AbsoluteMTPositionX].Max
	abs_max_y := d.AbsoluteTypes()[evdev.AbsoluteMTPositionY].Max

	logger.Infof("开始读取设备 : %s", dev_name)
	d.Lock()
	defer d.Unlock()
	for {
		select {
		case <-global_close_signal:
			logger.Infof("释放设备 : %s", dev_name)
			return
		case event := <-event_ch:
			if event == nil {
				logger.Warnf("移除设备 : %s", dev_name)
				return
			} else if event.Type == evdev.SyncReport {
				pack := &event_pack{
					dev_name: dev_name,
					dev_type: type_touch_screen,
					events:   events,
				}
				event_reader <- pack
				events = make([]*evdev.Event, 0)
			} else {
				if event.Event.Code == absMtPositionX {
					event.Event.Value = int32(int64(event.Event.Value) * 0x7ffffffe / int64(abs_max_x))
				}
				if event.Event.Code == absMtPositionY {
					event.Event.Value = int32(int64(event.Event.Value) * 0x7ffffffe / int64(abs_max_y))
				}
				events = append(events, &event.Event)
			}
		}
	}
}

func udp_event_injector(ch chan *event_pack, port int) {
	time.Sleep(time.Duration(200) * time.Millisecond)
	conn, err := CreateUDPReader(fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		logger.Errorf("ERROR:%v", err)
		return
	}
	recv_ch := conn.ReadChan()
	logger.Infof("已准备从以下地址接收远程事件: ")
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ipv4 := ipNet.IP.To4()
			if ipv4 == nil {
				continue // 跳过非 IPv4 地址
			}
			if !ipv4.IsLoopback() {
				logger.Infof("UDP://%s:%v", ipv4, port)
			}
		}
	}
	for {
		select {
		case <-global_close_signal:
			return
		case pack := <-recv_ch: //数据包格式：<event_count:1byte><event1:8byte><event2:8byte>...<eventN:8byte><dev_type:1byte><dev_name:N byte>
			event_count := int(pack[0])
			events := make([]*evdev.Event, 0)
			for i := 0; i < event_count; i++ {
				event := &evdev.Event{
					Type:  evdev.EventType(uint16(binary.LittleEndian.Uint16(pack[8*i+1 : 8*i+3]))),
					Code:  uint16(binary.LittleEndian.Uint16(pack[8*i+3 : 8*i+5])),
					Value: int32(binary.LittleEndian.Uint32(pack[8*i+5 : 8*i+9])),
				}
				events = append(events, event)
			}
			e_pack := &event_pack{
				dev_name: string(pack[event_count*8+2:]),
				dev_type: dev_type(pack[event_count*8+1]),
				events:   events,
			}
			ch <- e_pack
		}
	}
}

func uds_event_injector(ch chan *event_pack, address string) {
	time.Sleep(time.Duration(200) * time.Millisecond)
	conn, err := CreateUDSReader(address)
	if err != nil {
		logger.Errorf("ERROR:%v", err)
		return
	}
	recv_ch := conn.ReadChan()
	logger.Infof("已准备从 UDS:[%s] 接收远程事件", address)
	for {
		select {
		case <-global_close_signal:
			return
		case pack := <-recv_ch:
			event_count := int(pack[0])
			events := make([]*evdev.Event, 0)
			for i := 0; i < event_count; i++ {
				event := &evdev.Event{
					Type:  evdev.EventType(uint16(binary.LittleEndian.Uint16(pack[8*i+1 : 8*i+3]))),
					Code:  uint16(binary.LittleEndian.Uint16(pack[8*i+3 : 8*i+5])),
					Value: int32(binary.LittleEndian.Uint32(pack[8*i+5 : 8*i+9])),
				}
				events = append(events, event)
			}
			e_pack := &event_pack{
				dev_name: string(pack[event_count*8+2:]),
				dev_type: dev_type(pack[event_count*8+1]),
				events:   events,
			}
			ch <- e_pack
		}
	}
}

var global_close_signal = make(chan bool)
var global_is_wordking_remote bool = false
var global_device_orientation int32 = 0
var global_screen_x int32 = 1000
var global_screen_y int32 = 1000

func get_device_orientation() int32 {
	output, err := exec.Command("sh", "-c", "dumpsys input").Output()
	if err != nil {
		return 0
	}
	re := regexp.MustCompile(`orientation=(\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		orientation := matches[1]
		result, err := strconv.Atoi(string(orientation))
		if err != nil {
			return 0
		} else {
			return int32(result)
		}
	} else {
		return 0
	}
}

func listen_device_orientation() {
	for {
		select {
		case <-global_close_signal:
			return
		default:
			var now_orientation int32 = get_device_orientation()
			if global_device_orientation != now_orientation {
				global_device_orientation = now_orientation
				logger.Debugf("设备方向改变\t[%d]", now_orientation)
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

func rotateAbsoluteXY(x, y int32) (int32, int32) {
	switch global_device_orientation {
	case 0:
		return x, y
	case 1:
		return 0x7ffffffe - y, x
	case 2:
		return 0x7ffffffe - x, 0x7ffffffe - y
	case 3:
		return y, 0x7ffffffe - x
	default:
		return x, y
	}
}

type dev_type uint8

const (
	type_mouse          = dev_type(0)
	type_keyboard       = dev_type(1)
	type_joystick       = dev_type(2)
	type_touch_screen   = dev_type(3)
	type_touch_pad      = dev_type(4)
	type_motion_sensors = dev_type(5)
	type_unknown        = dev_type(6)
)

const (
	_INPUT_PROP_POINTER = 0
	_INPUT_PROP_DIRECT  = 1
)

func check_dev_type(dev *evdev.Evdev, fd *os.File) dev_type {
	abs := dev.AbsoluteTypes()
	key := dev.KeyTypes()
	rel := dev.RelativeTypes()
	_, MTPositionX := abs[evdev.AbsoluteMTPositionX]
	_, MTPositionY := abs[evdev.AbsoluteMTPositionY]
	_, MTSlot := abs[evdev.AbsoluteMTSlot]
	_, MTTrackingID := abs[evdev.AbsoluteMTTrackingID]

	// [V3.4.5] 同步原版 V5.0.5 位运算精确识别触控设备
	if MTPositionX && MTPositionY && MTSlot && MTTrackingID {
		var bits int32
		ioctl(fd.Fd(), EVIOCGPROP(), uintptr(unsafe.Pointer(&bits)))
		if bits&(1<<_INPUT_PROP_DIRECT) != 0 {
			return type_touch_screen
		}
		if bits&(1<<_INPUT_PROP_POINTER) != 0 {
			return type_touch_pad
		}
		return type_unknown
	}
	_, RelX := rel[evdev.RelativeX]
	_, RelY := rel[evdev.RelativeY]
	_, Wheel := rel[evdev.RelativeWheel]
	_, MouseLeft := key[evdev.BtnLeft]
	_, MouseRight := key[evdev.BtnRight]
	if RelX && RelY && Wheel && MouseLeft && MouseRight {
		return type_mouse //鼠标 检测XY 滚轮 左右键
	}
	keyboard_keys := true
	for i := evdev.KeyEscape; i <= evdev.KeyScrollLock; i++ {
		_, ok := key[i]
		keyboard_keys = keyboard_keys && ok
	}
	if keyboard_keys {
		return type_keyboard
	}

	axis_count := len(abs)
	btn_count := len(key)
	if axis_count >= 4 {
		if btn_count == 0 {
			return type_motion_sensors
		} else if btn_count > 8 {
			return type_joystick
		}
	}
	return type_unknown
}

type byEventNum []os.FileInfo

func (s byEventNum) Len() int {
	return len(s)
}

func (s byEventNum) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byEventNum) Less(i, j int) bool {
	numI := mustNum(s[i].Name())
	numJ := mustNum(s[j].Name())
	return numI < numJ
}

func mustNum(name string) int {
	base := strings.TrimPrefix(name, "event")
	n, _ := strconv.Atoi(base)
	return n
}

func get_possible_device_indexes(skipList map[int]bool) map[int]dev_type {
	files, _ := ioutil.ReadDir("/dev/input")
	var events []os.FileInfo
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "event") {
			events = append(events, f)
		}
	}

	sort.Sort(byEventNum(events))
	result := make(map[int]dev_type)
	for _, file := range events {
		if file.IsDir() {
			continue
		}
		if len(file.Name()) <= 5 {
			continue
		}
		if file.Name()[:5] != "event" {
			continue
		}

		index, _ := strconv.Atoi(file.Name()[5:])
		reading, exist := skipList[index]
		if exist && reading {
			continue
		} else {
			fd, err := os.OpenFile(fmt.Sprintf("/dev/input/%s", file.Name()), os.O_RDONLY, 0)
			if err != nil {
				logger.Errorf("读取设备/dev/input/%s失败 : %v ", file.Name(), err)
				continue
			}
			d := evdev.Open(fd)
			devType := check_dev_type(d, fd)
			d.Close()
			if devType != type_unknown {
				result[index] = devType
			}
		}
	}
	return result
}

func get_dev_name_by_index(index int) string {
	fd, err := os.OpenFile(fmt.Sprintf("/dev/input/event%d", index), os.O_RDONLY, 0)
	if err != nil {
		return "读取设备名称失败"
	}
	d := evdev.Open(fd)
	defer d.Close()
	return d.Name()
}

func execute_view_move(handelerInstance *TouchHandler, x, stepValue, sleepMS int) {
	handelerInstance.handel_view_move(0, 0)
	time.Sleep(time.Millisecond * time.Duration(sleepMS))
	if x > 0 {
		steps := x / stepValue
		for i := 0; i < steps; i++ {
			if !handelerInstance.map_on {
				break
			}
			handelerInstance.handel_view_move(int32(stepValue), 0)
			time.Sleep(time.Millisecond * time.Duration(sleepMS))
		}
		handelerInstance.handel_view_move(int32(x%stepValue), 0)
	} else {
		steps := -x / stepValue
		for i := 0; i < steps; i++ {
			if !handelerInstance.map_on {
				break
			}
			handelerInstance.handel_view_move(-int32(stepValue), 0)
			time.Sleep(time.Millisecond * time.Duration(sleepMS))
		}
		handelerInstance.handel_view_move(-int32(x%stepValue), 0)
	}
}

func stdin_control_view_move(handelerInstance *TouchHandler) {
	logger.Info("输入数值以精确控制view移动 [ x | x 步长 间隔 ]")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		args := strings.Split(scanner.Text(), " ")
		var x, stepValue, sleepMS int
		var err error
		if len(args) == 1 {
			x, err = strconv.Atoi(args[0])
			if err != nil {
				logger.Errorf("输入错误: %s", err)
				continue
			} else {
				stepValue = 24
				sleepMS = 16
			}
		} else if len(args) == 3 {
			x, err = strconv.Atoi(args[0])
			if err != nil {
				logger.Errorf("输入错误: %s", err)
				continue
			}
			stepValue, err = strconv.Atoi(args[1])
			if err != nil {
				logger.Errorf("输入错误: %s", err)
				continue
			}
			sleepMS, err = strconv.Atoi(args[2])
			if err != nil {
				logger.Errorf("输入错误: %s", err)
				continue
			}
		} else {
			logger.Error("参数错误 usage:x | x stepValue sleepMS")
			continue
		}
		logger.Infof("x: %d stepValue: %d sleepMS: %d", x, stepValue, sleepMS)
		if handelerInstance.map_on {
			execute_view_move(handelerInstance, x, stepValue, sleepMS)
		} else {
			logger.Info("等待映射开关打开中...")
			for {
				if !handelerInstance.map_on {
					time.Sleep(time.Duration(100) * time.Millisecond)
				} else {
					execute_view_move(handelerInstance, x, stepValue, sleepMS)
					break
				}
			}
		}
	}
}

func get_MT_size(indexes map[int]bool) (int32, int32) {
	for index, _ := range indexes {
		fd, err := os.OpenFile(fmt.Sprintf("/dev/input/event%d", index), os.O_RDONLY, 0)
		if err != nil {
			logger.Errorf("get_MT_size error:%v", err)
		}
		d := evdev.Open(fd)
		defer d.Close()
		abs := d.AbsoluteTypes()
		MTPositionX, _ := abs[evdev.AbsoluteMTPositionX]
		MTPositionY, _ := abs[evdev.AbsoluteMTPositionY]
		return MTPositionX.Max, MTPositionY.Max
	}
	return int32(1), int32(1)
}

func OpenSerialWritePipe(portName string, baudRate int) (serial.Port, error) {
	mode := &serial.Mode{BaudRate: baudRate}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	return port, nil
}

func auto_detect_and_read(event_chan chan *event_pack, patern string) {
	devices := make(map[int]bool)
	for {
		select {
		case <-global_close_signal:
			return
		default:
			auto_detect_result := get_possible_device_indexes(devices)
			devTypeFriendlyName := map[dev_type]string{
				type_mouse:          "鼠标",
				type_keyboard:       "键盘",
				type_joystick:       "手柄",
				type_touch_screen:   "触屏",
				type_touch_pad:      "触摸板",
				type_motion_sensors: "运动传感器",
				type_unknown:        "未知",
			}
			for index, devType := range auto_detect_result {
				devName := get_dev_name_by_index(index)
				if devName == uinput_keyboard_mouse_dev_name {
					continue
				}
				re := regexp.MustCompile(patern)
				if !re.MatchString(devName) {
					continue
				}

				if devType == type_mouse || devType == type_keyboard || devType == type_joystick || devType == type_motion_sensors {
					logger.Infof("检测到设备 %s(/dev/input/event%d) : %s", devName, index, devTypeFriendlyName[devType])
					localIndex := index
					go func() {
						devices[localIndex] = true
						dev_reader(event_chan, localIndex)
						devices[localIndex] = false
					}()
				}
			}
			time.Sleep(time.Duration(400) * time.Millisecond)
		}
	}
}

func parseSenderAddress(s string, defaultPort int) (string, int, error) {
	parts := strings.Split(s, ":")
	if len(parts) == 1 {
		return parts[0], defaultPort, nil
	}
	if len(parts) == 2 {
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %v", err)
		}
		return parts[0], port, nil
	}
	return "", 0, fmt.Errorf("invalid address format")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

//go:embed configs
var configs embed.FS

func main() {
	parser := argparse.NewParser("go-touch-mapper", " ")

	var configPath *string = parser.String("c", "config", &argparse.Options{
		Required: false,
		Help:     "指定配置文件路径，如果不存在则会使用默认模板创建",
	})
	var create_js_info *bool = parser.Flag("", "create-js-info", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "创建手柄配置文件模式",
	})
	var patern *string = parser.String("", "pattern", &argparse.Options{
		Required: false,
		Default:  ".*",
		Help:     "用于筛选设备名称的正则",
	})
	var as_remote_control *string = parser.String("s", "sender", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "发送本地事件到远程",
	})
	var control_mode *string = parser.String("m", "mode", &argparse.Options{
		Required: false,
		Default:  "uinput",
		Help:     "触摸方案，可用控制模式:    \tuinput:\t\t使用uinput创建虚拟触屏  \tinputmanager:\t通过UDS控制安卓inputManager \thid:\t\t通过串口控制单片机模拟usb触屏  \totg:\t\t本机配置LinuxUSBgadget模拟usb触屏,设备文件为/dev/hidg0 \tdirect:\t\t直接写入设备真实触屏,需要root权限或者低版本安卓",
	})
	var mixTouchDisabled *bool = parser.Flag("t", "disable-mix", &argparse.Options{
		Required: false,
		Help:     "关闭触屏混合,仅在uinput与inputmanager模式生效",
		Default:  false,
	})
	
	var inputTouchEventIndex *int = parser.Int("i", "input-touch-index", &argparse.Options{
		Required: false,
		Default:  -1,
		Help:     "真实触屏设备index，在direct模式与触屏混合模式下指定真实触屏设备index,默认-1不指定",
	})
	var uinputMouseKeyboardDisabled *bool = parser.Flag("u", "disable-uinput", &argparse.Options{
		Required: false,
		Help:     "不再创建uinput鼠标键盘设备,仅在uinput、inputmanager与direct模式生效",
		Default:  false,
	})
	var usingInputManagerDisplayID *int = parser.Int("", "display-id", &argparse.Options{
		Required: false,
		Default:  0,
		Help:     "显示器ID,仅inputmanager模式生效",
	})
	var usingHIDTouchTtyPath *string = parser.String("", "tty-path", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "串口设备路径,hid模式下必须指定",
	})
	var usingDeviceRotation *int = parser.Int("", "rotation", &argparse.Options{
		Required: false,
		Default:  1,
		Help:     "手动指定屏幕方向,仅在hid与otg模式下需要",
	})
	var usingAppBridge *bool = parser.Flag("", "app", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "在hid模式下启用本机 App 桥接接口进行发送",
	})
	var usingAppBridgeMode *string = parser.String("", "app-mode", &argparse.Options{
		Required: false,
		Default:  "udp",
		Help:     "App 桥接协议模式，可选: tcp, udp。默认 udp",
	})
	var usingAppBridgeAddr *string = parser.String("", "app-addr", &argparse.Options{
		Required: false,
		Default:  "127.0.0.1:61071",
		Help:     "接口地址 默认 127.0.0.1:61071",
	})
	var using_remote_control *bool = parser.Flag("r", "remote-control", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "是否从UDP/uds接收远程事件",
	})
	var port *int = parser.Int("p", "port", &argparse.Options{
		Required: false,
		Help:     "指定监听远程事件的UDP端口号(默认61069)",
		Default:  61069,
	})
	var uds_address *string = parser.String("", "uds", &argparse.Options{
		Required: false,
		Default:  "@uds_mouse_keyboard_event_socket",
		Help:     "监听的uds地址",
	})
	var using_v_mouse *bool = parser.Flag("v", "v-mouse", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "用触摸操作模拟鼠标",
	})
	var using_remote_v_mouse *string = parser.String("", "v-mouse-addr", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "模拟光标显示程序地址",
	})
	var measure_sensitivity_mode *bool = parser.Flag("", "measure-mode", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "显示视角移动像素计数用于测试灵敏度",
	})
	var debug_mode *bool = parser.Flag("d", "debug-mode", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "打印debug信息",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	uinput_keyboard_mouse_dev_name = fmt.Sprintf("EVO80_Keyboard_%s", go_build_version)
	logger.Infof("当前构建版本: %v", go_build_version)

	if *debug_mode {
		logger.WithDebug()
		logger.Debug("debug on")
	}

	if *create_js_info {
		auto_detect_result := get_possible_device_indexes(make(map[int]bool))
		js_events := make([]int, 0)
		for index, devType := range auto_detect_result {
			if devType == type_joystick {
				js_events = append(js_events, index)
			}
		}
		if len(js_events) == 1 {
			create_js_info_file(js_events[0])
		} else {
			if len(js_events) == 0 {
				logger.Error("未检测到手柄")
			} else {
				logger.Infof("检测到多个手柄")
				for index, EventIndex := range js_events {
					devName := get_dev_name_by_index(EventIndex)
					logger.Infof("[%d] %s (/dev/input/event%d)", index, devName, EventIndex)
				}
				logger.Infof("请输入要创建配置文件的手柄索引")
				var js_index int
				fmt.Scanln(&js_index)
				if js_index < 0 || js_index >= len(js_events) {
					logger.Errorf("输入的手柄索引 %d 无效", js_index)
					return
				}
				create_js_info_file(js_events[js_index])
			}
		}
		return
	} else if *as_remote_control != "" {
		ip, remotePort, err := parseSenderAddress(*as_remote_control, 61069)
		if err != nil {
			logger.Errorf("解析发送地址失败: %v", err)
			return
		}
		logger.Infof("启动远程事件发送器 目标地址 %s:%d", ip, remotePort)
		
		// [V3.5.0] 提示并实现安全退出机制
		logger.Infof("已锁定本地设备，发送数据中……按下快捷组合键 ESC + F1 可结束程序并解除锁定")
		
		events_ch := make(chan *event_pack)
		go auto_detect_and_read(events_ch, *patern)
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(ip), Port: remotePort})
		if err != nil {
			fmt.Println("Dial UDP failed, err:", err)
			return
		}
		defer conn.Close()

		pack_count := 0
		pack_size := 0
		go (func() {
			ticker := time.NewTicker(time.Second * 1)
			for {
				select {
				case <-global_close_signal:
					return
				case <-ticker.C:
					logger.Debugf("发送频率: %d Pack/s 发送数据量: %d B/s\r", pack_count, pack_size)
					pack_count = 0
					pack_size = 0
				}
			}
		})()

		data := make([]byte, 4096)
		esc_pressed := false
		f1_pressed := false
		
		for {
			select {
			case <-global_close_signal:
				return
			case pack := <-events_ch:
				// [V3.5.0] 监听退出组合键
				trigger_exit := false
				for _, event := range pack.events {
					if event.Type == evdev.EventKey {
						if event.Code == 1 { // ESC
							esc_pressed = (event.Value != UP)
						} else if event.Code == 59 { // F1
							f1_pressed = (event.Value != UP)
						}
					}
				}

				if esc_pressed && f1_pressed {
					logger.Warnf("检测到 ESC + F1，准备解除锁定并退出程序...")
					// 强制发送 ESC 和 F1 的抬起事件，防止远端按键粘连
					pack.events = []*evdev.Event{
						{Type: evdev.EventKey, Code: 1, Value: UP},
						{Type: evdev.EventKey, Code: 59, Value: UP},
						{Type: evdev.EventSync, Code: 0, Value: 0},
					}
					trigger_exit = true
				}

				event_count := len(pack.events)
				data[0] = byte(event_count)
				for i, event := range pack.events {
					offset := 1 + i*8
					binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(event.Type))
					binary.LittleEndian.PutUint16(data[offset+2:offset+4], event.Code)
					binary.LittleEndian.PutUint32(data[offset+4:offset+8], uint32(event.Value))
				}
				name_offset := 1 + event_count*8
				data[name_offset] = byte(pack.dev_type)
				if pack.dev_type == type_keyboard || pack.dev_type == type_mouse {
					pack.dev_name = "rkm"
				}
				copy(data[name_offset+1:], []byte(pack.dev_name))
				data_length := name_offset + 1 + len(pack.dev_name)
				_, err = conn.Write(data[:data_length])
				pack_count++
				pack_size += data_length
				if err != nil {
					fmt.Println("Send data failed, err:", err)
					return
				}

				if trigger_exit {
					time.Sleep(100 * time.Millisecond) // 确保最后的抬起包发送完成
					os.Exit(0)
				}
			}
		}
	} else {
		switch *control_mode {
		case "uinput":
		case "inputmanager":
		case "hid":
			if !*usingAppBridge && *usingHIDTouchTtyPath == "" {
				logger.Error("物理串口模式需要使用--tty-path参数指定设备，或使用 --app 参数走本机桥接模式")
				os.Exit(1)
			}
		case "otg":
		case "direct":
		default:
			logger.Errorf("未知模式%s", *control_mode)
			os.Exit(1)
		}

		if *configPath == "" {
			logger.Warn("未指定配置文件，使用默认配置文件")
			exePath, _ := os.Executable()
			exeDir := filepath.Dir(exePath)
			*configPath = filepath.Join(exeDir, "default.json")
		}
		if !fileExists(*configPath) {
			bytes, _ := configs.ReadFile("configs/EXAMPLE.JSON")
			logger.Infof("已从模板创建配置文件:%v", *configPath)
			os.WriteFile(*configPath, bytes, 0644)
		}
		logger.Infof("使用配置文件: %v", *configPath)

		pm, _ := InitPluginManager()

		main_events_ch := make(chan *event_pack)
		mix_touch_event_ch := make(chan *event_pack)
		u_input_control_ch := make(chan *u_input_control_pack)
		fileted_u_input_control_ch := make(chan *u_input_control_pack)

		go auto_detect_and_read(main_events_ch, *patern)

		// [V3.4.5] 同步原版 V5.0.5 真实触屏处理逻辑
		if !*mixTouchDisabled && (*control_mode == "uinput" || *control_mode == "inputmanager") {
			if *inputTouchEventIndex >= 0 {
				logger.Warnf("指定了真实触屏设备 /dev/input/event%d,将忽略其他触屏设备", *inputTouchEventIndex)
				go touch_dev_reader(mix_touch_event_ch, *inputTouchEventIndex)
			} else {
				for index, devType := range get_possible_device_indexes(make(map[int]bool)) {
					if devType == type_touch_screen {
						logger.Infof("启用触屏混合 %s(/dev/input/event%d)", get_dev_name_by_index(index), index)
						go touch_dev_reader(mix_touch_event_ch, index)
					}
				}
			}
		}

		var touch_control_func touch_control_func

		switch *control_mode {
		case "uinput":
			logger.Info("触屏控制将使用uinput在本机处理")
			if *uinputMouseKeyboardDisabled {
				go (func() {
					for {
						select {
						case <-global_close_signal:
							return
						case <-fileted_u_input_control_ch:
						}
					}
				})()
			} else {
				go handel_u_input_mouse_keyboard(fileted_u_input_control_ch)
			}
			touch_control_func = handel_touch_using_uinput_touch()
		case "inputmanager":
			logger.Info("触屏控制将使用inputManager在本机处理")
			if *uinputMouseKeyboardDisabled {
				go (func() {
					for {
						select {
						case <-global_close_signal:
							return
						case <-fileted_u_input_control_ch:
						}
					}
				})()
			} else {
				go handel_u_input_mouse_keyboard(fileted_u_input_control_ch)
			}
			touch_control_func = handel_touch_using_input_manager(*usingInputManagerDisplayID)
		case "hid":
			global_is_wordking_remote = true
			if *usingDeviceRotation < 0 || *usingDeviceRotation > 3 {
				logger.Error("旋转参数错误 可用选值有 0(竖屏) 1(横屏) 2(反向竖屏) 3(反向横屏)")
				return
			}
			
			go (func() {
				for {
					select {
					case <-global_close_signal:
						return
					case <-fileted_u_input_control_ch:
					}
				}
			})()

			if *usingAppBridge {
				logger.Info("触屏控制将通过 Android App 桥接接口 发送至 ESP32")
				logger.Infof("桥接通信模式：%s", *usingAppBridgeMode)
				logger.Infof("桥接接口地址：%s", *usingAppBridgeAddr)
				logger.Infof("触屏方向：%d", *usingDeviceRotation)
				touch_control_func = handel_touch_using_app_manager(*usingAppBridgeMode, *usingAppBridgeAddr)
			} else {
				if *usingHIDTouchTtyPath == "" {
					logger.Error("物理串口模式需要使用--tty-path参数指定设备，或使用 --app 参数走本机桥接模式")
					os.Exit(1)
				}
				logger.Info("触屏控制将使用串口控制外接的HID设备发送至主机")
				logger.Infof("串口路径：%s", *usingHIDTouchTtyPath)
				logger.Infof("触屏方向：%d", *usingDeviceRotation)
				port, err := OpenSerialWritePipe(*usingHIDTouchTtyPath, 115200)
				if err != nil {
					logger.Errorf("无法打开串口: %v", err)
					os.Exit(1)
				}
				touch_control_func = handel_touch_using_hid_manager(port)
			}
		case "otg":
			global_is_wordking_remote = true
			logger.Info("触屏控制将使用本机模拟为HID设备发送至主机")
			logger.Infof("触屏方向：%d", *usingDeviceRotation)
			global_device_orientation = int32(*usingDeviceRotation)
			touch_control_func = handel_touch_using_otg_manager()
			go (func() {
				for {
					select {
					case <-global_close_signal:
						return
					case <-fileted_u_input_control_ch:
					}
				}
			})()
		case "direct":
			if *uinputMouseKeyboardDisabled {
				go (func() {
					for {
						select {
						case <-global_close_signal:
							return
						case <-fileted_u_input_control_ch:
						}
					}
				})()
			} else {
				go handel_u_input_mouse_keyboard(fileted_u_input_control_ch)
			}

			var direct_touch_index int
			if *inputTouchEventIndex >= 0 {
				logger.Infof("指定了真实触屏设备 /dev/input/event%d,将直接写入该设备", *inputTouchEventIndex)
				direct_touch_index = *inputTouchEventIndex
			} else {
				for index, devType := range get_possible_device_indexes(make(map[int]bool)) {
					if devType == type_touch_screen {
						logger.Infof("将会直接写入触屏 %s(/dev/input/event%d)", get_dev_name_by_index(index), index)
						direct_touch_index = index
						break
					}
				}
			}
			touch_control_func = handel_touch_using_direct_touch(direct_touch_index)
		}

		map_switch_signal := make(chan bool)
		view_range_limited := true

		touchHandler := InitTouchHandler(
			*configPath,
			main_events_ch,
			touch_control_func,
			u_input_control_ch,
			view_range_limited,
			map_switch_signal,
			*measure_sensitivity_mode,
			pm,
		)

		if !global_is_wordking_remote {
			go touchHandler.mix_touch(mix_touch_event_ch)
			if !*using_v_mouse {
				go listen_device_orientation()
			}
		}

		go touchHandler.auto_handel_view_release()
		go touchHandler.loop_handel_wasd_wheel()
		go touchHandler.loop_handel_rs_move()
		go touchHandler.loop_handel_wheel_planet()
		go touchHandler.loop_auto_release_scroll_slider()

		go touchHandler.handel_event()

		if *using_v_mouse {
			ip := net.IPv4(0, 0, 0, 0)
			port := 6533
			if *using_remote_v_mouse != "" {
				s_ip, s_port, err := parseSenderAddress(*using_remote_v_mouse, 6533)
				if err != nil {
					logger.Errorf("解析模拟光标显示程序地址失败: %v", err)
					return
				}
				ip = net.ParseIP(s_ip)
				port = s_port
			}
			v_mouse := init_v_mouse_controller(touchHandler, u_input_control_ch, fileted_u_input_control_ch, map_switch_signal, net.UDPAddr{IP: ip, Port: port})
			go v_mouse.main_loop()
			go v_mouse.loop_handel_v_mouse_wheel_move()
		} else {
			go (func() {
				for {
					select {
					case tmp := <-u_input_control_ch:
						fileted_u_input_control_ch <- tmp
					case <-map_switch_signal:
					}
				}
			})()
		}

		if *using_remote_control {
			logger.Infof("使用远程控制中。。。。")
			go udp_event_injector(main_events_ch, *port)
			go uds_event_injector(main_events_ch, *uds_address)
		}

		if *measure_sensitivity_mode {
			go stdin_control_view_move(touchHandler)
		}

		go serve(*configPath, touchHandler.reloadConfigure, pm)

		exitChan := make(chan os.Signal, 1)
		signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
		<-exitChan
		close(global_close_signal)
		logger.Info("已停止")
		time.Sleep(time.Millisecond * 40)
	}
}