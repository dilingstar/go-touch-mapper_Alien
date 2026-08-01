package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

func uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

func readUint32(reader *bufio.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := reader.Read(buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

func writeString(writer *bufio.Writer, s string) error {
	lenBytes := uint32ToBytes(uint32(len(s)))
	writer.Write(lenBytes)
	writer.WriteString(s)
	return writer.Flush()
}

func readString(reader *bufio.Reader) (string, error) {
	len, err := readUint32(reader)
	if err != nil {
		return "", err
	}
	buf := make([]byte, len)
	if _, err := reader.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeInt32List(writer *bufio.Writer, list []int32) error {
	buffer := make([]byte, 4*len(list)+4)
	binary.LittleEndian.PutUint32(buffer[0:], uint32(len(list)))
	for i, v := range list {
		binary.LittleEndian.PutUint32(buffer[i*4+4:], uint32(v))
	}
	writer.Write(buffer)
	return writer.Flush()
}

func readInt32List(reader *bufio.Reader) ([]int32, error) {
	len, err := readUint32(reader)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4*len)
	if _, err := reader.Read(buf); err != nil {
		return nil, err
	}
	list := make([]int32, len)
	for i := range list {
		list[i] = int32(binary.LittleEndian.Uint32(buf[i*4:]))
	}
	return list, nil
}

type PluginManager struct {
	id                    string
	config_template       string
	user_config           map[string]interface{}
	user_config_file_path string
	writer                *bufio.Writer
	reader                *bufio.Reader
}

func InitPluginManager() (*PluginManager, error) {
	// 链接插件进程
	// 链接Unix Domain Socket
	// 等待插件进程连接
	// 插件进程连接后，会主动推送插件配置模版，版本信息等

	path, _ := exec.LookPath(os.Args[0])
	abspath, _ := filepath.Abs(path)
	workingDir, _ := filepath.Split(abspath)
	pluginBinPath := filepath.Join(workingDir, "plugin.bin")
	if _, err := os.Stat(pluginBinPath); os.IsNotExist(err) {
		logger.Debugf("未加载插件进程 : %s", pluginBinPath)
		return nil, err
	}
	pluginConfigDir := filepath.Join(workingDir, "pluginconfig")
	if _, err := os.Stat(pluginConfigDir); os.IsNotExist(err) {
		os.Mkdir(pluginConfigDir, os.ModePerm)
	}
	unixAddr, err := net.ResolveUnixAddr("unix", "@uds_go_touch_mapper_plugin")
	if err != nil {
		logger.Errorf("创建Unix Domain Socket失败 : %s", err.Error())
		os.Exit(3)
	}
	unixListener, _ := net.ListenUnix("unix", unixAddr)
	logger.Info("waiting for plugin process to connect")
	go func() {
		cmd := exec.Command(pluginBinPath)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				logger.Debugf("PLUGIN.STDOUT: %s", scanner.Text())
			}
		}()
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				logger.Debugf("PLUGIN.STDERR: %s", scanner.Text())
			}
		}()
		if err := cmd.Run(); err != nil {
			os.Exit(1) // 子进程异常，主进程退出
		}
		os.Exit(0) // 子进程正常结束，主进程退出
	}()
	unixConn, _ := unixListener.AcceptUnix()
	logger.Info("plugin connected")
	writer := bufio.NewWriter(unixConn)
	reader := bufio.NewReader(unixConn)
	go func() {
		<-global_close_signal
		unixConn.Close()
	}()
	// 链接之后 会主动推动插件ID，插件配置模版
	// 然后等待 接收插件用户配置
	plugin_id, err := readString(reader)
	if err != nil {
		logger.Debugf("读取插件ID失败 : %s", err.Error())
		os.Exit(3)
	} else {
		logger.Infof("插件ID: %s", plugin_id)
	}
	config_template, err := readString(reader)
	if err != nil {
		logger.Errorf("读取插件配置模版失败 : %s", err.Error())
		os.Exit(3)
	} else {
		logger.Infof("插件配置模版: %s", config_template)
	}
	user_config_file_path := filepath.Join(pluginConfigDir, fmt.Sprintf("%s.json", plugin_id))
	var user_config map[string]interface{}
	if !fileExists(user_config_file_path) {
		user_config = make(map[string]interface{})
		var config_template_json map[string]interface{}
		err := json.Unmarshal([]byte(config_template), &config_template_json)
		if err != nil {
			logger.Debugf("插件默认配置解析失败: %v", err)
			return nil, err
		} else {
			logger.Infof("插件默认配置项: %v", config_template_json)
			for k, v := range config_template_json {
				logger.Infof("插件默认配置项: %s = %v", k, v.(map[string]interface{}))
				user_config[k] = (v.(map[string]interface{}))["default"]
			}
		}
		cb, err := json.Marshal(user_config)
		logger.Infof("已创建插件默认配置文件:%v", user_config_file_path)
		os.WriteFile(user_config_file_path, cb, 0644)
	} else {
		raw, _ := os.ReadFile(user_config_file_path)
		json.Unmarshal(raw, &user_config)
	}
	userConfigStringBytes, _ := json.Marshal(user_config)
	userConfigString := string(userConfigStringBytes)
	logger.Infof("插件用户配置: %s", userConfigString)
	writeString(writer, userConfigString)
	return &PluginManager{
		id:                    plugin_id,
		config_template:       config_template,
		user_config:           user_config,
		user_config_file_path: user_config_file_path,
		writer:                writer,
		reader:                reader,
	}, nil
}

func (pm *PluginManager) update_user_config(config map[string]interface{}) {
	config_bytes, _ := json.Marshal(config)
	config_string := string(config_bytes)
	pm.writer.WriteByte(0xf1)
	writeString(pm.writer, config_string)
	cb, err := json.Marshal(config)
	if err != nil {
		logger.Errorf("插件用户配置序列化失败: %v", err)
		return
	}
	os.WriteFile(pm.user_config_file_path, cb, 0644)
	logger.Infof("已更新插件配置文件:%v", pm.user_config_file_path)
	logger.Infof("插件用户配置更新: %v", config)
}

func (pm *PluginManager) update_wheel_config(wheel_radius int32, shift_wheel_radius int32, center_x int32, center_y int32, screen_x int32, screen_y int32) {
	pm.writer.WriteByte(0xf2)
	writeInt32List(pm.writer, []int32{wheel_radius, shift_wheel_radius, center_x, center_y, screen_x, screen_y})
	logger.Infof("参数已设置:  wheel_radius=%d, shift_wheel_radius=%d, center_x=%d, center_y=%d, screen_x=%d, screen_y=%d", wheel_radius, shift_wheel_radius, center_x, center_y, screen_x, screen_y)
}

func (pm *PluginManager) get_wheel_move_target(wheel_pos_x, wheel_pos_y, wheel_axis_x, wheel_asix_y, shift_pressed int32,
) (int32, int32) {
	pm.writer.WriteByte(0xf3)
	writeInt32List(pm.writer, []int32{wheel_pos_x, wheel_pos_y, wheel_axis_x, wheel_asix_y, shift_pressed})
	outBytes := make([]byte, 9)
	if n, err := pm.reader.Read(outBytes); err != nil || n != 9 {
		logger.Errorf("读取插件返回值失败 : %s, 读取到 %d 字节", err.Error(), n)
		return 0, 0
	}
	x := int32(binary.LittleEndian.Uint32(outBytes[1:5]))
	y := int32(binary.LittleEndian.Uint32(outBytes[5:9]))
	return x, y
}
