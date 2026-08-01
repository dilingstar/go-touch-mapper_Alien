package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/kenshaw/evdev"
)

// Version: V3.5.0
type TouchHandler struct {
	events             chan *event_pack            //接收事件的channel
	touch_control_func touch_control_func          //发送触屏控制信号的channel
	u_input            chan *u_input_control_pack  //发送u_input控制信号的channel
	map_on             bool                        //映射模式开关
	view_id            int32                       //视角的触摸ID
	wheel_id           int32                       //左摇杆的触摸ID
	allocated_id       []bool                      //10个触摸点分配情况
	config             *simplejson.Json            //映射配置文件
	joystickInfo       map[string]*simplejson.Json //所有摇杆配置文件 dev_name 为key

	screen_x       int32
	screen_y       int32
	rel_screen_x   int32
	rel_screen_y   int32
	view_init_x    int32
	view_init_y    int32
	
	// [V3.4.6] 动态视角锚点，用于防越界半径(甜甜圈)的动态跟随
	view_anchor_x  int32
	view_anchor_y  int32
	
	view_current_x int32
	view_current_y int32
	view_speed_x   int32
	view_speed_y   int32

	rs_speed_x float64
	rs_speed_y float64

	wheel_init_x int32
	wheel_init_y int32
	wheel_range  int32
	wheel_wasd   []string

	view_lock          sync.Mutex
	wheel_lock         sync.Mutex
	touch_control_lock sync.Mutex

	id_alloc_lock             sync.Mutex
	auto_release_view_counter int32
	abs_last                  sync.Map
	using_joystick_name       string
	ls_wheel_released         bool
	ls_force_release_signal   bool 
	wasd_wheel_released       bool
	wasd_wheel_last_x         int32
	wasd_wheel_last_y         int32
	wasd_up_down_statues      []bool
	key_action_state_save     sync.Map
	BTN_SELECT_UP_DOWN        int32
	KEYBOARD_SWITCH_KEY_NAME_S map[string]bool

	POINTER_SWITCH_KEY_NAME_S map[string]bool
	pointer_switch_key_status map[string]bool
	
	// 核心状态：临时切出指针（半映射模式）标志
	pointer_is_out_temp       bool

	view_range_limited         bool
	map_switch_signal          chan bool
	measure_sensitivity_mode   bool
	total_move_x               int32
	total_move_y               int32
	wheel_shift_enable         bool
	wheel_shift_range          int32

	key_jitter_enable    bool
	key_jitter_amount_px int32

	wheel_step_speed           float64

	wheel_bezier_enable        bool
	wheel_bezier_speed         float64
	wheel_bezier_curve_amount  float64
	wheel_bezier_dynamic_curve float64

	wheel_random_point_enable        bool
	wheel_random_start_radius_px     int32
	wheel_random_end_radius_px       int32
	wheel_random_shift_end_radius_px int32

	wheel_easing_enable bool
	wheel_easing_in     float64
	wheel_easing_out    float64

	wheel_noise_enable    bool
	wheel_noise_intensity float64
	wheel_noise_counter   float64
	wheel_noise_fade      float64

	wheel_planet_enable          bool
	wheel_planet_radius_px       int32
	wheel_planet_speed           float64
	planet_angle                 float64
	wheel_planet_noise_intensity float64

	wheel_star_x int32
	wheel_star_y int32

	shift_press_toggle   bool
	shift_release_toggle bool
	shift_state          bool

	wheel_walk_mode_enable        bool
	wheel_temp_penetration_enable bool
	wheel_penetrating             bool

	planet_dynamic_speed_enable  bool
	planet_dynamic_speed_min     float64
	planet_dynamic_speed_freq    float64
	planet_dynamic_speed_counter float64

	planet_curve_enable    bool
	planet_curve_amount_px int32
	planet_curve_freq      float64

	view_auto_release_enable       bool
	view_auto_release_ms           int
	view_reset_radius_enable       bool
	view_reset_radius_px           int32
	view_reset_radius_thickness_px int32
	view_random_reset_enable       bool
	view_random_reset_radius_px    int32
	view_delay_reset_enable        bool
	view_delay_reset_ms            int
	view_delay_reset_random_enable bool
	view_delay_reset_min_ms        int
	view_resetting_lock            bool

	scroll_slider_enable                bool
	scroll_slider_init_x                int32
	scroll_slider_init_y                int32
	scroll_slider_bound_up              int32
	scroll_slider_bound_down            int32
	scroll_slider_timeout_duration      time.Duration
	scroll_slider_speed_px              int32
	scroll_slider_release_delay_ms      int
	scroll_slider_noise_enable          bool
	scroll_slider_noise_intensity       float64
	scroll_slider_random_enable         bool
	scroll_slider_random_radius_px      int32
	scroll_slider_curve_enable          bool
	scroll_slider_curve_amount_px       int32
	scroll_slider_curve_freq            float64
	scroll_slider_delay_reset_ms        int
	scroll_slider_delay_random_enable   bool
	scroll_slider_delay_reset_min_ms    int

	scroll_slider_id              int32
	scroll_slider_current_y       int32
	scroll_slider_last_scroll_time time.Time
	scroll_slider_last_reset_time  time.Time
	scroll_slider_lock            sync.Mutex

	wheel_delay_reset_duration time.Duration
	wheel_last_input_time      time.Time

	real_key_down_state sync.Map

	recoil_active        bool
	current_recoil_speed float64
	base_recoil_speed    float64

	recoil_trigger_status map[string]bool
	recoil_scope_status   map[string]bool
	recoil_scope_mode     bool

	pm *PluginManager

	combo_handler *ComboHandler
	macro_handler *MacroHandler 
	
	end_exit_enable bool
	end_exit_keys   []string
}

const (
	TouchActionRequire int8 = 0
	TouchActionRelease int8 = 1
	TouchActionMove    int8 = 2
)

const (
	TouchActionResetResolution int8 = 3
)

const (
	UInput_mouse_move  int8 = 0
	UInput_mouse_btn   int8 = 1
	UInput_mouse_wheel int8 = 2
	UInput_key_event   int8 = 3
)

const (
	DOWN int32 = 1
	UP   int32 = 0
)

var UDF map[int32](string) = map[int32](string){
	DOWN: "🟢",
	UP:   "🔴",
}

const (
	Wheel_action_move    int8 = 1
	Wheel_action_release int8 = 0
)

const (
	MAIN_LOOP_HZ        = 250.0
	MAIN_LOOP_NS_DOUBLE = float64(time.Second) / MAIN_LOOP_HZ
	MAIN_LOOP_NS_INT    = int64(MAIN_LOOP_NS_DOUBLE)
)

var HAT_D_U map[string]([]int32) = map[string]([]int32){
	"0.5_1.0": []int32{1, DOWN},
	"0.5_0.0": []int32{0, DOWN},
	"1.0_0.5": []int32{1, UP},
	"0.0_0.5": []int32{0, UP},
}

var HAT0_KEY_NAME map[string][]string = map[string][]string{
	"HAT0X": {"BTN_DPAD_LEFT", "BTN_DPAD_RIGHT"},
	"HAT0Y": {"BTN_DPAD_UP", "BTN_DPAD_DOWN"},
}

// 辅助函数
func get_jitter_offset(amount int32) int32 {
	if amount <= 0 {
		return 0
	}
	return (rand.Int31n(2*amount + 1)) - amount
}

func (self *TouchHandler) get_random_offset(radius_px int32) (int32, int32) {
	if radius_px <= 0 {
		return 0, 0
	}
	rand_angle := rand.Float64() * 2 * math.Pi
	rand_radius := rand.Float64() * float64(radius_px)
	offset_x := int32(rand_radius * math.Cos(rand_angle))
	offset_y := int32(rand_radius * math.Sin(rand_angle))
	return offset_x, offset_y
}

func (self *TouchHandler) get_delay_ms(base_ms int, min_ms int, random_enable bool) time.Duration {
	if !random_enable {
		return time.Duration(base_ms) * time.Millisecond
	}
	if min_ms >= base_ms {
		return time.Duration(base_ms) * time.Millisecond
	}
	delay := rand.Intn(base_ms-min_ms+1) + min_ms
	return time.Duration(delay) * time.Millisecond
}

func (self *TouchHandler) get_random_duration(max_ms int, min_ms int) time.Duration {
	if min_ms >= max_ms {
		return time.Duration(max_ms) * time.Millisecond
	}
	dur := rand.Intn(max_ms-min_ms+1) + min_ms
	return time.Duration(dur) * time.Millisecond
}

func (self *TouchHandler) get_dynamic_speed(
	max_speed float64,
	min_speed float64,
	freq float64,
	counter *float64,
) float64 {
	if min_speed >= max_speed {
		return max_speed
	}
	rad_per_frame := (freq * 2 * math.Pi) / MAIN_LOOP_HZ
	*counter += rad_per_frame
	if *counter > (2 * math.Pi) {
		*counter -= (2 * math.Pi)
	}
	sin_wave := (math.Sin(*counter) + 1) / 2.0
	speed_range := max_speed - min_speed
	current_speed := (sin_wave * speed_range) + min_speed
	return current_speed
}

func (self *TouchHandler) apply_key_jitter(x int32, y int32) (int32, int32) {
	if !self.key_jitter_enable || self.key_jitter_amount_px == 0 {
		return x, y
	}
	offset_x, offset_y := self.get_random_offset(self.key_jitter_amount_px)
	return x + offset_x, y + offset_y
}

func (self *TouchHandler) getStick(stick_name string) (float64, float64) {
	if jsconfig, ok := self.joystickInfo[self.using_joystick_name]; ok {
		_x, _ := self.abs_last.Load(stick_name + "_X")
		_y, _ := self.abs_last.Load(stick_name + "_Y")
		if _x == nil || _y == nil {
			return 0.5, 0.5
		}
		x, y := _x.(float64), _y.(float64)
		deadZone_left := jsconfig.Get("DEADZONE").Get(stick_name).GetIndex(0).MustFloat64()
		deadZone_right := jsconfig.Get("DEADZONE").Get(stick_name).GetIndex(1).MustFloat64()
		if deadZone_left < x && x < deadZone_right && deadZone_left < y && y < deadZone_right {
			return 0.5, 0.5
		} else {
			return x, y
		}
	} else {
		return 0.5, 0.5
	}
}

func (self *TouchHandler) get_scaled_pos(x int32, y int32) (int32, int32) {
	return int32(int64(x) * 0x7ffffffe / int64(self.rel_screen_x)), int32(int64(y) * 0x7ffffffe / int64(self.rel_screen_y))
}

func (self *TouchHandler) handel_key_events(events []*evdev.Event, dev_type dev_type, dev_name string) {
	if jsconfig, ok := self.joystickInfo[dev_name]; ok && dev_type == type_joystick {
		for _, event := range events {
			if key_name, ok := jsconfig.Get("BTN").CheckGet(strconv.Itoa(int(event.Code))); ok {
				self.handel_key_up_down(key_name.MustString(), event.Value, dev_name)
			} else {
				logger.Debugf("joyStick[%s]\t%d\t未知键码", dev_name, event.Code)
			}
		}
	} else {
		for _, event := range events {
			self.handel_key_up_down(GetKeyName(event.Code), event.Value, dev_name)
		}
	}
}

func (self *TouchHandler) handel_abs_events(events []*evdev.Event, dev_type dev_type, dev_name string) {
	LS_MOVED := false
	var ls_target_x int32
	var ls_target_y int32
	for _, event := range events {
		if dev_type == type_joystick {
			if jsconfig, ok := self.joystickInfo[dev_name]; ok {
				abs_info := jsconfig.Get("ABS").Get(strconv.Itoa(int(event.Code)))
				name := abs_info.Get("name").MustString("")
				abs_mini := int32(abs_info.Get("range").GetIndex(0).MustInt())
				abs_max := int32(abs_info.Get("range").GetIndex(1).MustInt())
				formatted_value := float64(event.Value-abs_mini) / float64(abs_max-abs_mini)
				_last_value, _ := self.abs_last.Load(name)
				var last_value float64
				if _last_value != nil {
					last_value = _last_value.(float64)
				} else {
					last_value = 0.5 // Default center
				}

				if name == "HAT0X" || name == "HAT0Y" {
					down_up_key := fmt.Sprintf("%s_%s", strconv.FormatFloat(last_value, 'f', 1, 64), strconv.FormatFloat(formatted_value, 'f', 1, 64))
					self.abs_last.Store(name, formatted_value)
					if instruction, ok := HAT_D_U[down_up_key]; ok {
						direction := instruction[0]
						up_down := instruction[1]
						translated_name := HAT0_KEY_NAME[name][direction]
						self.handel_key_up_down(translated_name, up_down, dev_name)
					}
				} else if name == "LT" || name == "RT" {
					for i := 0; i < 6; i++ {
						if last_value < float64(i)/5 && formatted_value >= float64(i)/5 {
							translated_name := fmt.Sprintf("%s_%d", name, i)
							self.handel_key_up_down("BTN_"+translated_name, DOWN, dev_name)
							if i == 1 {
								self.handel_key_up_down("BTN_"+name, DOWN, dev_name)
							}
						} else if last_value >= float64(i)/5 && formatted_value < float64(i)/5 {
							translated_name := fmt.Sprintf("%s_%d", name, i)
							self.handel_key_up_down("BTN_"+translated_name, UP, dev_name)
							if i == 1 {
								self.handel_key_up_down("BTN_"+name, UP, dev_name)
							}
						}
					}
					self.abs_last.Store(name, formatted_value)
				} else { 
					if self.using_joystick_name != dev_name {
						self.using_joystick_name = dev_name
					}
					self.abs_last.Store(name, formatted_value)

					if (name == "LS_X" || name == "LS_Y") && (self.map_on || self.wheel_penetrating) {
						ls_x, ls_y := self.getStick("LS")
						if ls_x == 0.5 && ls_y == 0.5 {
							if self.ls_wheel_released == false {
								self.ls_wheel_released = true
								self.ls_force_release_signal = true
							}
						} else {
							self.ls_wheel_released = false
							self.ls_force_release_signal = false 
							wheel_range := self.wheel_range
							if self.wheel_shift_enable {
								wheel_range = self.wheel_shift_range
							}
							ls_target_x = self.wheel_init_x + int32(float64(wheel_range)*2*(ls_x-0.5))
							ls_target_y = self.wheel_init_y + int32(float64(wheel_range)*2*(ls_y-0.5))
							LS_MOVED = true
						}
					}
				}
			} else {
				logger.Warnf("%v config not found", dev_name)
			}
		} else if dev_type == type_motion_sensors {
			self.handel_gyro_events(events, dev_name)
		}
	}
	if LS_MOVED {
		self.handel_wheel_action(Wheel_action_move, ls_target_x, ls_target_y, true)
	}
}

func InitTouchHandler(
	mapperFilePath string,
	events chan *event_pack,
	touch_control_func touch_control_func,
	u_input chan *u_input_control_pack,
	view_range_limited bool,
	map_switch_signal chan bool,
	measure_sensitivity_mode bool,
	pm *PluginManager,
) *TouchHandler {
	rand.Seed(time.Now().UnixNano())

	if _, err := os.Stat(mapperFilePath); os.IsNotExist(err) {
		logger.Errorf("没有找到映射配置文件 : %s ", mapperFilePath)
		os.Exit(1)
	} else {
		logger.Infof("使用映射配置文件 : %s ", mapperFilePath)
	}

	content, _ := ioutil.ReadFile(mapperFilePath)
	config_json, _ := simplejson.NewJson(content)

	joystickInfo := make(map[string]*simplejson.Json)
	rjsJson := []byte(`{
    "DEADZONE": { "LS": [0.05, 0.05], "RS": [0.05, 0.05] },
    "ABS": {
        "7": {"name": "HAT0Y", "range": [-1, 1], "reverse": false},
        "6": {"name": "HAT0X", "range": [-1, 1], "reverse": false},
        "0": {"name": "LS_X", "range": [-32767, 32767], "reverse": false},
        "1": {"name": "LS_Y", "range": [-32767, 32767], "reverse": false},
		"2": {"name": "RS_X", "range": [-32767, 32767], "reverse": false},
        "3": {"name": "RS_Y", "range": [-32767, 32767], "reverse": false},
        "4": {"name": "LT", "range": [-1023, 1023], "reverse": false},
        "5": {"name": "RT", "range": [-1023, 1023], "reverse": false}
    },
    "BTN": {
        "0": "BTN_A", "1": "BTN_B", "2": "BTN_X", "3": "BTN_Y", "8": "BTN_LS", "9": "BTN_RS",
        "4": "BTN_LB", "5": "BTN_RB", "6": "BTN_SELECT", "7": "BTN_START", "10": "BTN_HOME"
    },
    "MAP_KEYBOARD": {
        "BTN_LT": "BTN_RIGHT", "BTN_RT": "BTN_LEFT", "BTN_DPAD_UP": "KEY_UP",
        "BTN_DPAD_LEFT": "KEY_LEFT", "BTN_DPAD_RIGHT": "KEY_RIGHT", "BTN_DPAD_DOWN": "KEY_DOWN",
        "BTN_A": "KEY_ENTER", "BTN_B": "KEY_BACK", "BTN_SELECT": "KEY_COMPOSE", "BTN_THUMBL": "KEY_HOME"
    }
}`)
	rjsJsonObj, err := simplejson.NewJson(rjsJson)
	if err != nil {
		logger.Errorf("Failed to parse rjs joystick config: %v", err)
		os.Exit(1)
	}
	joystickInfo["rjs"] = rjsJsonObj
	path, _ := exec.LookPath(os.Args[0])
	abs, _ := filepath.Abs(path)
	workingDir, _ := filepath.Split(abs)
	joystickInfosDir := filepath.Join(workingDir, "joystickInfos")
	if _, err := os.Stat(joystickInfosDir); os.IsNotExist(err) {
		logger.Warnf("%s 文件夹不存在,没有载入任何手柄配置文件", joystickInfosDir)
	} else {
		files, _ := ioutil.ReadDir(joystickInfosDir)
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if file.Name()[len(file.Name())-5:] != ".json" {
				continue
			}
			content, _ := ioutil.ReadFile(filepath.Join(joystickInfosDir, file.Name()))
			info, _ := simplejson.NewJson(content)
			joystickInfo[file.Name()[:len(file.Name())-5]] = info
			logger.Infof("手柄配置文件已载入 : %s", file.Name())
		}
	}

	abs_last_map := sync.Map{}
	abs_last_map.Store("HAT0X", 0.5)
	abs_last_map.Store("HAT0Y", 0.5)
	abs_last_map.Store("LT", 0.0)
	abs_last_map.Store("RT", 0.0)
	abs_last_map.Store("LS_X", 0.5)
	abs_last_map.Store("LS_Y", 0.5)
	abs_last_map.Store("RS_X", 0.5)
	abs_last_map.Store("RS_Y", 0.5)

	screenSizeX := config_json.Get("SCREEN").Get("SIZE").GetIndex(0).MustInt(3200)
	screenSizeY := config_json.Get("SCREEN").Get("SIZE").GetIndex(1).MustInt(1440)
	screenSizeX_f := float64(screenSizeX)
	screenSizeY_f := float64(screenSizeY)
	global_screen_x = int32(screenSizeX)
	global_screen_y = int32(screenSizeY)

	KEYBOARD_SWITCH_KEY_NAME_S := make(map[string]bool)
	for _, key := range config_json.Get("MOUSE").Get("SWITCH_KEYS").MustStringArray() {
		if key != "" {
			KEYBOARD_SWITCH_KEY_NAME_S[key] = true
		} else {
			logger.Warnf("映射配置文件中有空的键盘切换按键,请检查配置文件")
		}
	}

	POINTER_SWITCH_KEY_NAME_S := make(map[string]bool)
	if pointerKeys, ok := config_json.Get("MOUSE").CheckGet("POINTER_SWITCH_KEYS"); ok {
		for _, key := range pointerKeys.MustStringArray() {
			if key != "" {
				POINTER_SWITCH_KEY_NAME_S[key] = true
			}
		}
	}

	var key_jitter_enable_val bool
	var key_jitter_amount_px_val int32
	if keyJitterJSON, ok := config_json.CheckGet("KEY_JITTER"); ok {
		key_jitter_enable_val = keyJitterJSON.Get("ENABLE").MustBool(true)
		key_jitter_amount_val := keyJitterJSON.Get("AMOUNT").MustFloat64(0.003)
		key_jitter_amount_px_val = int32(key_jitter_amount_val * screenSizeX_f)
	} else {
		key_jitter_enable_val = true
		key_jitter_amount_px_val = int32(0.003 * screenSizeX_f)
	}

	wheel_cfg := config_json.Get("WHEEL")
	wheel_delay_reset_ms_val := wheel_cfg.Get("DELAY_RESET_MS").MustInt(50)
	wheel_walk_mode_enable_val := wheel_cfg.Get("WALK_MODE_ENABLE").MustBool(false)

	wheel_temp_penetration_enable := wheel_cfg.Get("TEMP_PENETRATION_ENABLE").MustBool(false)
	wheel_penetrating := false

	wheel_step_speed_val := wheel_cfg.Get("STEP_SPEED").MustFloat64(60.0)

	wheel_bezier_enable_val := wheel_cfg.Get("BEZIER_ENABLE").MustBool(false)
	wheel_bezier_speed_val := wheel_cfg.Get("BEZIER_SPEED").MustFloat64(60.0)
	wheel_bezier_curve_amount_val := wheel_cfg.Get("BEZIER_CURVE_AMOUNT").MustFloat64(0.5)
	wheel_bezier_dynamic_curve_val := wheel_cfg.Get("BEZIER_DYNAMIC_CURVE").MustFloat64(0.0)

	wheel_random_point_enable_val := wheel_cfg.Get("RANDOM_POINT_ENABLE").MustBool(false)
	wheel_random_start_radius_px_val := int32(wheel_cfg.Get("RANDOM_START_RADIUS").MustFloat64(0.01) * screenSizeX_f)
	wheel_random_end_radius_px_val := int32(wheel_cfg.Get("RANDOM_END_RADIUS").MustFloat64(0.01) * screenSizeX_f)
	wheel_random_shift_end_radius_px_val := int32(wheel_cfg.Get("RANDOM_SHIFT_END_RADIUS").MustFloat64(0.015) * screenSizeX_f)

	wheel_easing_enable_val := wheel_cfg.Get("EASING_ENABLE").MustBool(false)
	wheel_easing_in_val := wheel_cfg.Get("EASING_IN").MustFloat64(0.2)
	wheel_easing_out_val := wheel_cfg.Get("EASING_OUT").MustFloat64(0.2)

	wheel_noise_enable_val := wheel_cfg.Get("NOISE_ENABLE").MustBool(false)
	wheel_noise_intensity_val := wheel_cfg.Get("NOISE_INTENSITY").MustFloat64(0.002)
	wheel_noise_counter := 0.0
	wheel_noise_fade := 0.0

	wheel_planet_enable_val := wheel_cfg.Get("WHEEL_PLANET").Get("ENABLE").MustBool(false)
	wheel_planet_radius_px_val := int32(wheel_cfg.Get("WHEEL_PLANET").Get("RADIUS").MustFloat64(0.015) * screenSizeX_f)
	wheel_planet_speed_val := wheel_cfg.Get("WHEEL_PLANET").Get("SPEED").MustFloat64(1.5) / MAIN_LOOP_HZ
	wheel_planet_noise_intensity_val := wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_NOISE_INTENSITY").MustFloat64(0.002)

	planet_dynamic_speed_enable_val := wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_DYNAMIC_SPEED").Get("ENABLE").MustBool(false)
	planet_dynamic_speed_min_val := wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_DYNAMIC_SPEED").Get("MIN_SPEED").MustFloat64(0.5)
	planet_dynamic_speed_freq_val := wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_DYNAMIC_SPEED").Get("FREQUENCY").MustFloat64(1.0)

	planet_curve_enable_val := wheel_cfg.Get("PLANET_CURVE").Get("ENABLE").MustBool(false)
	planet_curve_amount_val := wheel_cfg.Get("PLANET_CURVE").Get("CURVE_AMOUNT").MustFloat64(0.005)
	planet_curve_amount_px_val := int32(planet_curve_amount_val * screenSizeX_f)
	planet_curve_freq_val := wheel_cfg.Get("PLANET_CURVE").Get("CURVE_FREQUENCY").MustFloat64(1.0)

	mouse_cfg := config_json.Get("MOUSE")
	view_auto_release_enable_val := mouse_cfg.Get("VIEW_AUTO_RELEASE_ENABLE").MustBool(false)
	view_auto_release_ms_val := mouse_cfg.Get("VIEW_AUTO_RELEASE_MS").MustInt(200)
	view_reset_radius_enable_val := mouse_cfg.Get("VIEW_RESET_RADIUS_ENABLE").MustBool(false)
	view_reset_radius_px_val := int32(mouse_cfg.Get("VIEW_RESET_RADIUS").MustFloat64(0.1) * screenSizeX_f)
	view_reset_radius_thickness_px_val := int32(mouse_cfg.Get("VIEW_RESET_RADIUS_THICKNESS").MustFloat64(0.005) * screenSizeX_f)
	view_random_reset_enable_val := mouse_cfg.Get("VIEW_RANDOM_RESET_ENABLE").MustBool(false)
	view_random_reset_radius_px_val := int32(mouse_cfg.Get("VIEW_RANDOM_RESET_RADIUS").MustFloat64(0.01) * screenSizeX_f)
	view_delay_reset_enable_val := mouse_cfg.Get("VIEW_DELAY_RESET_ENABLE").MustBool(false)
	view_delay_reset_ms_val := mouse_cfg.Get("VIEW_DELAY_RESET_MS").MustInt(20)
	view_delay_reset_random_enable_val := mouse_cfg.Get("VIEW_DELAY_RESET_RANDOM_ENABLE").MustBool(false)
	view_delay_reset_min_ms_val := mouse_cfg.Get("VIEW_DELAY_RESET_MIN_MS").MustInt(10)
	view_resetting_lock := false

	scroll_cfg := config_json.Get("SCROLL_SLIDER")
	scroll_slider_enable_val := scroll_cfg.Get("ENABLE").MustBool(false)
	scroll_slider_init_x_val := int32(scroll_cfg.Get("POS").GetIndex(0).MustFloat64(0.9) * screenSizeX_f)
	scroll_slider_init_y_val := int32(scroll_cfg.Get("POS").GetIndex(1).MustFloat64(0.5) * screenSizeY_f)
	scroll_slider_bound_up_val := scroll_slider_init_y_val - int32(scroll_cfg.Get("LENGTH_UP").MustFloat64(0.2)*screenSizeY_f)
	scroll_slider_bound_down_val := scroll_slider_init_y_val + int32(scroll_cfg.Get("LENGTH_DOWN").MustFloat64(0.2)*screenSizeY_f)
	scroll_slider_timeout_duration_val := time.Duration(scroll_cfg.Get("TIMEOUT_S").MustFloat64(3.0) * float64(time.Second))
	scroll_slider_speed_val := scroll_cfg.Get("SPEED").MustFloat64(1.0)
	scroll_slider_speed_px_val := int32(scroll_slider_speed_val * 0.01 * screenSizeY_f)

	scroll_slider_release_delay_ms_val := scroll_cfg.Get("RELEASE_DELAY_MS").MustInt(50)
	scroll_slider_noise_enable_val := scroll_cfg.Get("NOISE_ENABLE").MustBool(false)
	scroll_slider_noise_intensity_val := scroll_cfg.Get("NOISE_INTENSITY").MustFloat64(0.002)

	scroll_slider_random_enable_val := scroll_cfg.Get("RANDOM_START_ENABLE").MustBool(false)
	scroll_slider_random_radius_px_val := int32(scroll_cfg.Get("RANDOM_START_RADIUS").MustFloat64(0.005) * screenSizeX_f)
	scroll_slider_curve_enable_val := scroll_cfg.Get("CURVE_ENABLE").MustBool(false)
	scroll_slider_curve_amount_px_val := int32(scroll_cfg.Get("CURVE_AMOUNT").MustFloat64(0.005) * screenSizeX_f)
	scroll_slider_curve_freq_val := scroll_cfg.Get("CURVE_FREQUENCY").MustFloat64(1.0)
	scroll_slider_delay_reset_ms_val := scroll_cfg.Get("DELAY_RESET_MS").MustInt(20)
	scroll_slider_delay_random_enable_val := scroll_cfg.Get("DELAY_RANDOM_ENABLE").MustBool(false)
	scroll_slider_delay_reset_min_ms_val := scroll_cfg.Get("DELAY_RESET_MIN_MS").MustInt(10)

	scroll_slider_id := int32(-1)
	scroll_slider_current_y := scroll_slider_init_y_val
	scroll_slider_last_scroll_time := time.Now()
	scroll_slider_last_reset_time := time.Now()

	recoil_cfg := config_json.Get("GLOBAL_RECOIL")
	base_recoil_speed_val := recoil_cfg.Get("BASE_SPEED").MustFloat64(0)
	current_recoil_speed_val := base_recoil_speed_val
	recoil_scope_mode_val := recoil_cfg.Get("SCOPE_MODE").MustBool(false)
	recoil_active_val := false
	
	// [V3.5.0] 解析退出快捷键配置
	exit_cfg := config_json.Get("END_EXIT")
	end_exit_enable_val := exit_cfg.Get("EXIT_ENABLE").MustBool(false)
	var end_exit_keys_val []string
	if keys, err := exit_cfg.Get("EXIT_KEYS").StringArray(); err == nil {
		end_exit_keys_val = keys
	} else {
		end_exit_keys_val = []string{}
	}

	if pm != nil {
		var shift_range int32 = int32(config_json.Get("WHEEL").Get("RANGE").MustFloat64() * float64(screenSizeX))
		if config_json.Get("WHEEL").Get("SHIFT_RANGE_ENABLE").MustBool() == true {
			shift_range = int32(config_json.Get("WHEEL").Get("SHIFT_RANGE").MustFloat64() * float64(screenSizeX))
		}
		pm.update_wheel_config(
			int32(config_json.Get("WHEEL").Get("RANGE").MustFloat64()*float64(screenSizeX)),
			shift_range,
			int32(config_json.Get("WHEEL").Get("POS").GetIndex(0).MustFloat64()*float64(screenSizeX)),
			int32(config_json.Get("WHEEL").Get("POS").GetIndex(1).MustFloat64()*float64(screenSizeY)),
			int32(screenSizeX),
			int32(screenSizeY),
		)
	}

	handler := &TouchHandler{
		events:             events,
		touch_control_func: touch_control_func,
		u_input:            u_input,
		map_on:             false,
		view_id:            -1,
		wheel_id:           -1,
		allocated_id:       make([]bool, 12),
		config:             config_json,
		joystickInfo:       joystickInfo,

		screen_x:       0x7ffffffe,
		screen_y:       0x7ffffffe,
		rel_screen_x:   int32(screenSizeX),
		rel_screen_y:   int32(screenSizeY),
		view_init_x:    int32(config_json.Get("MOUSE").Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe),
		view_init_y:    int32(config_json.Get("MOUSE").Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe),
		
		// [V3.4.6] 初始化动态锚点
		view_anchor_x:  int32(config_json.Get("MOUSE").Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe),
		view_anchor_y:  int32(config_json.Get("MOUSE").Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe),
		
		view_current_x: int32(config_json.Get("MOUSE").Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe),
		view_current_y: int32(config_json.Get("MOUSE").Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe),
		view_speed_x:   int32(config_json.Get("MOUSE").Get("SPEED").GetIndex(0).MustFloat64() * 0x7ffffffe / screenSizeX_f),
		view_speed_y:   int32(config_json.Get("MOUSE").Get("SPEED").GetIndex(1).MustFloat64() * 0x7ffffffe / screenSizeX_f),

		rs_speed_x: 32,
		rs_speed_y: 32,

		wheel_init_x: int32(config_json.Get("WHEEL").Get("POS").GetIndex(0).MustFloat64() * screenSizeX_f),
		wheel_init_y: int32(config_json.Get("WHEEL").Get("POS").GetIndex(1).MustFloat64() * screenSizeY_f),
		wheel_range:  int32(config_json.Get("WHEEL").Get("RANGE").MustFloat64() * screenSizeX_f),

		wheel_wasd: []string{
			config_json.Get("WHEEL").Get("WASD").GetIndex(0).MustString(),
			config_json.Get("WHEEL").Get("WASD").GetIndex(1).MustString(),
			config_json.Get("WHEEL").Get("WASD").GetIndex(2).MustString(),
			config_json.Get("WHEEL").Get("WASD").GetIndex(3).MustString(),
		},
		view_lock:                  sync.Mutex{},
		wheel_lock:                 sync.Mutex{},
		touch_control_lock:         sync.Mutex{},
		id_alloc_lock:              sync.Mutex{},
		auto_release_view_counter:  0,
		abs_last:                   abs_last_map,
		using_joystick_name:        "",
		ls_wheel_released:          true,
		ls_force_release_signal:    false, 
		wasd_wheel_released:        true,
		wasd_wheel_last_x:          int32(config_json.Get("WHEEL").Get("POS").GetIndex(0).MustFloat64() * screenSizeX_f),
		wasd_wheel_last_y:          int32(config_json.Get("WHEEL").Get("POS").GetIndex(1).MustFloat64() * screenSizeY_f),
		wasd_up_down_statues:       make([]bool, 5),
		key_action_state_save:      sync.Map{},
		BTN_SELECT_UP_DOWN:         0,
		KEYBOARD_SWITCH_KEY_NAME_S: KEYBOARD_SWITCH_KEY_NAME_S,
		POINTER_SWITCH_KEY_NAME_S:  POINTER_SWITCH_KEY_NAME_S,
		pointer_switch_key_status:  make(map[string]bool),
		
		pointer_is_out_temp:        false, // 初始默认为 false

		view_range_limited:       view_range_limited,
		map_switch_signal:        map_switch_signal,
		measure_sensitivity_mode: measure_sensitivity_mode,
		wheel_shift_enable:       config_json.Get("WHEEL").Get("SHIFT_RANGE_ENABLE").MustBool(false),
		wheel_shift_range:        int32(config_json.Get("WHEEL").Get("SHIFT_RANGE").MustFloat64() * screenSizeX_f),

		key_jitter_enable:    key_jitter_enable_val,
		key_jitter_amount_px: key_jitter_amount_px_val,

		wheel_step_speed: wheel_step_speed_val,

		wheel_bezier_enable:              wheel_bezier_enable_val,
		wheel_bezier_speed:               wheel_bezier_speed_val,
		wheel_bezier_curve_amount:        wheel_bezier_curve_amount_val,
		wheel_bezier_dynamic_curve:       wheel_bezier_dynamic_curve_val,
		wheel_random_point_enable:        wheel_random_point_enable_val,
		wheel_random_start_radius_px:     wheel_random_start_radius_px_val,
		wheel_random_end_radius_px:       wheel_random_end_radius_px_val,
		wheel_random_shift_end_radius_px: wheel_random_shift_end_radius_px_val,
		wheel_easing_enable:              wheel_easing_enable_val,
		wheel_easing_in:                  wheel_easing_in_val,
		wheel_easing_out:                 wheel_easing_out_val,
		wheel_noise_enable:               wheel_noise_enable_val,
		wheel_noise_intensity:            wheel_noise_intensity_val,
		wheel_noise_counter:              wheel_noise_counter,
		wheel_noise_fade:                 wheel_noise_fade,

		wheel_delay_reset_duration:    time.Duration(wheel_delay_reset_ms_val) * time.Millisecond,
		wheel_last_input_time:         time.Now(),
		wheel_walk_mode_enable:        wheel_walk_mode_enable_val,
		wheel_temp_penetration_enable: wheel_temp_penetration_enable,
		wheel_penetrating:             wheel_penetrating,

		wheel_planet_enable:          wheel_planet_enable_val,
		wheel_planet_radius_px:       wheel_planet_radius_px_val,
		wheel_planet_speed:           wheel_planet_speed_val / MAIN_LOOP_HZ,
		planet_angle:                 0,
		wheel_planet_noise_intensity: wheel_planet_noise_intensity_val,

		planet_dynamic_speed_enable:  planet_dynamic_speed_enable_val,
		planet_dynamic_speed_min:     planet_dynamic_speed_min_val,
		planet_dynamic_speed_freq:    planet_dynamic_speed_freq_val,
		planet_dynamic_speed_counter: 0,

		planet_curve_enable:    planet_curve_enable_val,
		planet_curve_amount_px: planet_curve_amount_px_val,
		planet_curve_freq:      planet_curve_freq_val,

		shift_press_toggle:   config_json.Get("WHEEL").Get("SHIFT_PRESS_TOGGLE").MustBool(false),
		shift_release_toggle: config_json.Get("WHEEL").Get("SHIFT_RELEASE_TOGGLE").MustBool(false),
		shift_state:          false,

		wheel_star_x: 0,
		wheel_star_y: 0,

		view_auto_release_enable:       view_auto_release_enable_val,
		view_auto_release_ms:           view_auto_release_ms_val,
		view_reset_radius_enable:       view_reset_radius_enable_val,
		view_reset_radius_px:           view_reset_radius_px_val,
		view_reset_radius_thickness_px: view_reset_radius_thickness_px_val,
		view_random_reset_enable:       view_random_reset_enable_val,
		view_random_reset_radius_px:    view_random_reset_radius_px_val,
		view_delay_reset_enable:        view_delay_reset_enable_val,
		view_delay_reset_ms:            view_delay_reset_ms_val,
		view_delay_reset_random_enable: view_delay_reset_random_enable_val,
		view_delay_reset_min_ms:        view_delay_reset_min_ms_val,
		view_resetting_lock:            view_resetting_lock,

		scroll_slider_enable:              scroll_slider_enable_val,
		scroll_slider_init_x:              scroll_slider_init_x_val,
		scroll_slider_init_y:              scroll_slider_init_y_val,
		scroll_slider_bound_up:            scroll_slider_bound_up_val,
		scroll_slider_bound_down:          scroll_slider_bound_down_val,
		scroll_slider_timeout_duration:    scroll_slider_timeout_duration_val,
		scroll_slider_speed_px:            scroll_slider_speed_px_val,
		scroll_slider_release_delay_ms:    scroll_slider_release_delay_ms_val,
		scroll_slider_noise_enable:        scroll_slider_noise_enable_val,
		scroll_slider_noise_intensity:     scroll_slider_noise_intensity_val,
		scroll_slider_random_enable:       scroll_slider_random_enable_val,
		scroll_slider_random_radius_px:    scroll_slider_random_radius_px_val,
		scroll_slider_curve_enable:        scroll_slider_curve_enable_val,
		scroll_slider_curve_amount_px:     scroll_slider_curve_amount_px_val,
		scroll_slider_curve_freq:          scroll_slider_curve_freq_val,
		scroll_slider_delay_reset_ms:      scroll_slider_delay_reset_ms_val,
		scroll_slider_delay_random_enable: scroll_slider_delay_random_enable_val,
		scroll_slider_delay_reset_min_ms:  scroll_slider_delay_reset_min_ms_val,

		scroll_slider_id:               scroll_slider_id,
		scroll_slider_current_y:        scroll_slider_current_y,
		scroll_slider_last_scroll_time: scroll_slider_last_scroll_time,
		scroll_slider_last_reset_time:  scroll_slider_last_reset_time,
		scroll_slider_lock:             sync.Mutex{},

		real_key_down_state: sync.Map{},

		recoil_active:         recoil_active_val,
		base_recoil_speed:     base_recoil_speed_val,
		current_recoil_speed:  current_recoil_speed_val,
		recoil_trigger_status: make(map[string]bool),
		recoil_scope_status:   make(map[string]bool),
		recoil_scope_mode:     recoil_scope_mode_val,
		
		// [V3.5.0] 初始化退出快捷键
		end_exit_enable: end_exit_enable_val,
		end_exit_keys:   end_exit_keys_val,

		pm: pm,
	}

	handler.combo_handler = InitComboHandler(handler)
	handler.combo_handler.LoadCombos()

	handler.macro_handler = InitMacroHandler(handler)
	handler.macro_handler.LoadMacros()

	go handler.loop_handel_recoil()

	return handler
}

func (self *TouchHandler) reloadConfigure(mapperFilePath string) {
	if self.map_on {
		self.switch_map_mode()
	}
	if _, err := os.Stat(mapperFilePath); os.IsNotExist(err) {
		logger.Errorf("没有找到映射配置文件 : %s ", mapperFilePath)
		os.Exit(1)
	} else {
		logger.Infof("使用映射配置文件 : %s ", mapperFilePath)
	}
	content, _ := ioutil.ReadFile(mapperFilePath)
	config_json, _ := simplejson.NewJson(content)
	screenSizeX := config_json.Get("SCREEN").Get("SIZE").GetIndex(0).MustInt(3200)
	screenSizeY := config_json.Get("SCREEN").Get("SIZE").GetIndex(1).MustInt(1440)
	screenSizeX_f := float64(screenSizeX)
	screenSizeY_f := float64(screenSizeY)
	global_screen_x = int32(screenSizeX)
	global_screen_y = int32(screenSizeY)

	self.config = config_json

	self.screen_x = 0x7ffffffe
	self.screen_y = 0x7ffffffe
	self.rel_screen_x = int32(screenSizeX)
	self.rel_screen_y = int32(screenSizeY)
	self.view_init_x = int32(config_json.Get("MOUSE").Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe)
	self.view_init_y = int32(config_json.Get("MOUSE").Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe)
	
	// [V3.4.6] 重载配置时同步更新锚点
	self.view_anchor_x = self.view_init_x
	self.view_anchor_y = self.view_init_y
	
	self.view_current_x = int32(config_json.Get("MOUSE").Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe)
	self.view_current_y = int32(config_json.Get("MOUSE").Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe)
	self.view_speed_x = int32(config_json.Get("MOUSE").Get("SPEED").GetIndex(0).MustFloat64() * 0x7ffffffe / screenSizeX_f)
	self.view_speed_y = int32(config_json.Get("MOUSE").Get("SPEED").GetIndex(1).MustFloat64() * 0x7ffffffe / screenSizeX_f)

	self.wheel_init_x = int32(config_json.Get("WHEEL").Get("POS").GetIndex(0).MustFloat64() * screenSizeX_f)
	self.wheel_init_y = int32(config_json.Get("WHEEL").Get("POS").GetIndex(1).MustFloat64() * screenSizeY_f)
	self.wheel_range = int32(config_json.Get("WHEEL").Get("RANGE").MustFloat64() * screenSizeX_f)

	self.wheel_wasd = []string{
		config_json.Get("WHEEL").Get("WASD").GetIndex(0).MustString(),
		config_json.Get("WHEEL").Get("WASD").GetIndex(1).MustString(),
		config_json.Get("WHEEL").Get("WASD").GetIndex(2).MustString(),
		config_json.Get("WHEEL").Get("WASD").GetIndex(3).MustString(),
	}
	self.wasd_wheel_last_x = int32(config_json.Get("WHEEL").Get("POS").GetIndex(0).MustFloat64() * screenSizeX_f)
	self.wasd_wheel_last_y = int32(config_json.Get("WHEEL").Get("POS").GetIndex(1).MustFloat64() * screenSizeY_f)
	self.KEYBOARD_SWITCH_KEY_NAME_S = make(map[string]bool)
	for _, key := range config_json.Get("MOUSE").Get("SWITCH_KEYS").MustStringArray() {
		if key != "" {
			self.KEYBOARD_SWITCH_KEY_NAME_S[key] = true
		} else {
			logger.Warnf("映射配置文件中有空的键盘切换按键,请检查配置文件")
		}
	}

	self.POINTER_SWITCH_KEY_NAME_S = make(map[string]bool)
	if pointerKeys, ok := config_json.Get("MOUSE").CheckGet("POINTER_SWITCH_KEYS"); ok {
		for _, key := range pointerKeys.MustStringArray() {
			if key != "" {
				self.POINTER_SWITCH_KEY_NAME_S[key] = true
			}
		}
	}

	// [V3.4.5] 重载时清空指针标志位
	self.pointer_is_out_temp = false

	self.wheel_shift_enable = config_json.Get("WHEEL").Get("SHIFT_RANGE_ENABLE").MustBool(false)
	self.wheel_shift_range = int32(config_json.Get("WHEEL").Get("SHIFT_RANGE").MustFloat64() * screenSizeX_f)

	if keyJitterJSON, ok := config_json.CheckGet("KEY_JITTER"); ok {
		self.key_jitter_enable = keyJitterJSON.Get("ENABLE").MustBool(true)
		self.key_jitter_amount_px = int32(keyJitterJSON.Get("AMOUNT").MustFloat64(0.003) * screenSizeX_f)
	} else {
		self.key_jitter_enable = true
		self.key_jitter_amount_px = int32(0.003 * screenSizeX_f)
	}

	wheel_cfg := config_json.Get("WHEEL")
	self.wheel_delay_reset_duration = time.Duration(wheel_cfg.Get("DELAY_RESET_MS").MustInt(50)) * time.Millisecond
	self.wheel_last_input_time = time.Now()
	self.wheel_walk_mode_enable = wheel_cfg.Get("WALK_MODE_ENABLE").MustBool(false)

	self.wheel_temp_penetration_enable = wheel_cfg.Get("TEMP_PENETRATION_ENABLE").MustBool(false)
	self.wheel_penetrating = false

	self.wheel_step_speed = wheel_cfg.Get("STEP_SPEED").MustFloat64(60.0)

	self.wheel_bezier_enable = wheel_cfg.Get("BEZIER_ENABLE").MustBool(false)
	self.wheel_bezier_speed = wheel_cfg.Get("BEZIER_SPEED").MustFloat64(60.0)
	self.wheel_bezier_curve_amount = wheel_cfg.Get("BEZIER_CURVE_AMOUNT").MustFloat64(0.5)
	self.wheel_bezier_dynamic_curve = wheel_cfg.Get("BEZIER_DYNAMIC_CURVE").MustFloat64(0.0)

	self.wheel_random_point_enable = wheel_cfg.Get("RANDOM_POINT_ENABLE").MustBool(false)
	self.wheel_random_start_radius_px = int32(wheel_cfg.Get("RANDOM_START_RADIUS").MustFloat64(0.01) * screenSizeX_f)
	self.wheel_random_end_radius_px = int32(wheel_cfg.Get("RANDOM_END_RADIUS").MustFloat64(0.01) * screenSizeX_f)
	self.wheel_random_shift_end_radius_px = int32(wheel_cfg.Get("RANDOM_SHIFT_END_RADIUS").MustFloat64(0.015) * screenSizeX_f)

	self.wheel_easing_enable = wheel_cfg.Get("EASING_ENABLE").MustBool(false)
	self.wheel_easing_in = wheel_cfg.Get("EASING_IN").MustFloat64(0.2)
	self.wheel_easing_out = wheel_cfg.Get("EASING_OUT").MustFloat64(0.2)

	self.wheel_noise_enable = wheel_cfg.Get("NOISE_ENABLE").MustBool(false)
	self.wheel_noise_intensity = wheel_cfg.Get("NOISE_INTENSITY").MustFloat64(0.002)
	self.wheel_noise_counter = 0
	self.wheel_noise_fade = 0.0

	self.wheel_planet_enable = wheel_cfg.Get("WHEEL_PLANET").Get("ENABLE").MustBool(false)
	self.wheel_planet_radius_px = int32(wheel_cfg.Get("WHEEL_PLANET").Get("RADIUS").MustFloat64(0.015) * screenSizeX_f)
	self.wheel_planet_speed = wheel_cfg.Get("WHEEL_PLANET").Get("SPEED").MustFloat64(1.5) / MAIN_LOOP_HZ
	self.wheel_planet_noise_intensity = wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_NOISE_INTENSITY").MustFloat64(0.002)

	self.planet_dynamic_speed_enable = wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_DYNAMIC_SPEED").Get("ENABLE").MustBool(false)
	self.planet_dynamic_speed_min = wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_DYNAMIC_SPEED").Get("MIN_SPEED").MustFloat64(0.5)
	self.planet_dynamic_speed_freq = wheel_cfg.Get("WHEEL_PLANET").Get("PLANET_DYNAMIC_SPEED").Get("FREQUENCY").MustFloat64(1.0)

	self.planet_curve_enable = wheel_cfg.Get("PLANET_CURVE").Get("ENABLE").MustBool(false)
	self.planet_curve_amount_px = int32(wheel_cfg.Get("PLANET_CURVE").Get("CURVE_AMOUNT").MustFloat64(0.005) * screenSizeX_f)
	self.planet_curve_freq = wheel_cfg.Get("PLANET_CURVE").Get("CURVE_FREQUENCY").MustFloat64(1.0)

	mouse_cfg := config_json.Get("MOUSE")
	self.view_auto_release_enable = mouse_cfg.Get("VIEW_AUTO_RELEASE_ENABLE").MustBool(false)
	self.view_auto_release_ms = mouse_cfg.Get("VIEW_AUTO_RELEASE_MS").MustInt(200)
	self.view_reset_radius_enable = mouse_cfg.Get("VIEW_RESET_RADIUS_ENABLE").MustBool(false)
	self.view_reset_radius_px = int32(mouse_cfg.Get("VIEW_RESET_RADIUS").MustFloat64(0.1) * screenSizeX_f)
	self.view_reset_radius_thickness_px = int32(mouse_cfg.Get("VIEW_RESET_RADIUS_THICKNESS").MustFloat64(0.005) * screenSizeX_f)
	self.view_random_reset_enable = mouse_cfg.Get("VIEW_RANDOM_RESET_ENABLE").MustBool(false)
	self.view_random_reset_radius_px = int32(mouse_cfg.Get("VIEW_RANDOM_RESET_RADIUS").MustFloat64(0.01) * screenSizeX_f)
	self.view_delay_reset_enable = mouse_cfg.Get("VIEW_DELAY_RESET_ENABLE").MustBool(false)
	self.view_delay_reset_ms = mouse_cfg.Get("VIEW_DELAY_RESET_MS").MustInt(20)
	self.view_delay_reset_random_enable = mouse_cfg.Get("VIEW_DELAY_RESET_RANDOM_ENABLE").MustBool(false)
	self.view_delay_reset_min_ms = mouse_cfg.Get("VIEW_DELAY_RESET_MIN_MS").MustInt(10)
	self.view_resetting_lock = false

	scroll_cfg := config_json.Get("SCROLL_SLIDER")
	self.scroll_slider_enable = scroll_cfg.Get("ENABLE").MustBool(false)
	self.scroll_slider_init_x = int32(scroll_cfg.Get("POS").GetIndex(0).MustFloat64(0.9) * screenSizeX_f)
	self.scroll_slider_init_y = int32(scroll_cfg.Get("POS").GetIndex(1).MustFloat64(0.5) * screenSizeY_f)
	self.scroll_slider_bound_up = self.scroll_slider_init_y - int32(scroll_cfg.Get("LENGTH_UP").MustFloat64(0.2)*screenSizeY_f)
	self.scroll_slider_bound_down = self.scroll_slider_init_y + int32(scroll_cfg.Get("LENGTH_DOWN").MustFloat64(0.2)*screenSizeY_f)
	self.scroll_slider_timeout_duration = time.Duration(scroll_cfg.Get("TIMEOUT_S").MustFloat64(3.0) * float64(time.Second))
	scroll_slider_speed_val := scroll_cfg.Get("SPEED").MustFloat64(1.0)
	self.scroll_slider_speed_px = int32(scroll_slider_speed_val * 0.01 * screenSizeY_f)

	self.scroll_slider_release_delay_ms = scroll_cfg.Get("RELEASE_DELAY_MS").MustInt(50)
	self.scroll_slider_noise_enable = scroll_cfg.Get("NOISE_ENABLE").MustBool(false)
	self.scroll_slider_noise_intensity = scroll_cfg.Get("NOISE_INTENSITY").MustFloat64(0.002)

	self.scroll_slider_random_enable = scroll_cfg.Get("RANDOM_START_ENABLE").MustBool(false)
	self.scroll_slider_random_radius_px = int32(scroll_cfg.Get("RANDOM_START_RADIUS").MustFloat64(0.005) * screenSizeX_f)
	self.scroll_slider_curve_enable = scroll_cfg.Get("CURVE_ENABLE").MustBool(false)
	self.scroll_slider_curve_amount_px = int32(scroll_cfg.Get("CURVE_AMOUNT").MustFloat64(0.005) * screenSizeX_f)
	self.scroll_slider_curve_freq = scroll_cfg.Get("CURVE_FREQUENCY").MustFloat64(1.0)
	self.scroll_slider_delay_reset_ms = scroll_cfg.Get("DELAY_RESET_MS").MustInt(20)
	self.scroll_slider_delay_random_enable = scroll_cfg.Get("DELAY_RANDOM_ENABLE").MustBool(false)
	self.scroll_slider_delay_reset_min_ms = scroll_cfg.Get("DELAY_RESET_MIN_MS").MustInt(10)

	self.scroll_slider_id = -1
	self.scroll_slider_current_y = self.scroll_slider_init_y
	self.scroll_slider_last_scroll_time = time.Now()
	self.scroll_slider_last_reset_time = time.Now()

	recoil_cfg := config_json.Get("GLOBAL_RECOIL")
	self.base_recoil_speed = recoil_cfg.Get("BASE_SPEED").MustFloat64(0)
	self.current_recoil_speed = self.base_recoil_speed
	self.recoil_scope_mode = recoil_cfg.Get("SCOPE_MODE").MustBool(false)
	self.recoil_active = false

	self.recoil_trigger_status = make(map[string]bool)
	self.recoil_scope_status = make(map[string]bool)
	
	// [V3.5.0] 重载退出快捷键配置
	exit_cfg := config_json.Get("END_EXIT")
	self.end_exit_enable = exit_cfg.Get("EXIT_ENABLE").MustBool(false)
	if keys, err := exit_cfg.Get("EXIT_KEYS").StringArray(); err == nil {
		self.end_exit_keys = keys
	} else {
		self.end_exit_keys = []string{}
	}

	if self.pm != nil {
		var shift_range int32 = int32(config_json.Get("WHEEL").Get("RANGE").MustFloat64() * float64(screenSizeX))
		if config_json.Get("WHEEL").Get("SHIFT_RANGE_ENABLE").MustBool() == true {
			shift_range = int32(config_json.Get("WHEEL").Get("SHIFT_RANGE").MustFloat64() * float64(screenSizeX))
		}
		self.pm.update_wheel_config(
			int32(config_json.Get("WHEEL").Get("RANGE").MustFloat64()*float64(screenSizeX)),
			shift_range,
			int32(config_json.Get("WHEEL").Get("POS").GetIndex(0).MustFloat64()*float64(screenSizeX)),
			int32(config_json.Get("WHEEL").Get("POS").GetIndex(1).MustFloat64()*float64(screenSizeY)),
			int32(screenSizeX),
			int32(screenSizeY),
		)
	}

	self.shift_press_toggle = config_json.Get("WHEEL").Get("SHIFT_PRESS_TOGGLE").MustBool(false)
	self.shift_release_toggle = config_json.Get("WHEEL").Get("SHIFT_RELEASE_TOGGLE").MustBool(false)
	self.shift_state = false

	if self.combo_handler != nil {
		self.combo_handler.LoadCombos()
	}

	if self.macro_handler != nil {
		self.macro_handler.LoadMacros()
	}
}

// [V3.4.5] touch_require 与 touch_move 同步原版的 TOUCH_MAJOR 压力等扩展支持
// 在底层的 input_manager 和 uinput 接口中会使用到这些数据
func (self *TouchHandler) touch_require(x int32, y int32, scale bool) int32 {
	self.id_alloc_lock.Lock()
	defer self.id_alloc_lock.Unlock()

	var final_x, final_y int32
	if scale {
		final_x, final_y = self.get_scaled_pos(x, y)
	} else {
		final_x, final_y = x, y
	}

	for i, v := range self.allocated_id {
		if !v {
			self.allocated_id[i] = true
			self.send_touch_control_pack(TouchActionRequire, int32(i), final_x, final_y)
			logger.Debugf("touch require (%v,%v) => [%v]", x, y, i)
			return int32(i)
		}
	}
	return -1
}

func (self *TouchHandler) touch_release(id int32) int32 {
	logger.Debugf("touch release [%v]", id)
	if id != -1 {
		self.id_alloc_lock.Lock()
		if id >= 0 && id < int32(len(self.allocated_id)) && self.allocated_id[int(id)] {
			self.allocated_id[int(id)] = false
			self.send_touch_control_pack(TouchActionRelease, id, -1, -1)
		}
		self.id_alloc_lock.Unlock()
	}
	return -1
}

func (self *TouchHandler) touch_move(id int32, x int32, y int32, scale bool) {
	logger.Debugf("touch move to (%v,%v) [%v]", x, y, id)
	if id != -1 {
		var final_x, final_y int32
		if scale {
			final_x, final_y = self.get_scaled_pos(x, y)
		} else {
			final_x, final_y = x, y
		}
		self.send_touch_control_pack(TouchActionMove, id, final_x, final_y)
	}
}

func (self *TouchHandler) u_input_control(action int8, arg1 int32, arg2 int32) {
	self.u_input <- &u_input_control_pack{
		action: action,
		arg1:   arg1,
		arg2:   arg2,
	}
}

// [V3.4.5] 同步原版，传入 screen_x 和 screen_y，底层支持压力与接触面积换算
func (self *TouchHandler) send_touch_control_pack(action int8, id int32, x int32, y int32) {
	self.touch_control_lock.Lock()
	defer self.touch_control_lock.Unlock()
	self.touch_control_func(touch_control_pack{
		action:   action,
		id:       id,
		x:        x,
		y:        y,
		screen_x: self.rel_screen_x,
		screen_y: self.rel_screen_y,
	})
}

func (self *TouchHandler) mix_touch(touch_events chan *event_pack) {
	id_2_vid := make([]int32, 10)
	var last_id int32 = 0
	pos_s := make([][]int32, 10)
	for i := 0; i < 10; i++ {
		pos_s[i] = make([]int32, 2)
	}
	id_statuses := make([]bool, 10)
	for i := 0; i < 10; i++ {
		id_statuses[i] = false
	}

	translate_xy := func(x, y int32) (int32, int32) {
		switch global_device_orientation {
		case 0:
			return x, y
		case 1:
			return y, 0x7ffffffe - x
		case 2:
			return 0x7ffffffe - x, 0x7ffffffe - y
		case 3:
			return 0x7ffffffe - y, x
		default:
			return x, y
		}
	}

	for {
		copy_pos_s := make([][]int32, 10)
		copy(copy_pos_s, pos_s)
		copy_id_statuses := make([]bool, 10)
		copy(copy_id_statuses, id_statuses)
		select {
		case <-global_close_signal:
			return
		case event_pack := <-touch_events:
			for _, event := range event_pack.events {
				switch event.Code {
				case ABS_MT_POSITION_X:
					pos_s[last_id] = []int32{event.Value, pos_s[last_id][1]}
				case ABS_MT_POSITION_Y:
					pos_s[last_id] = []int32{pos_s[last_id][0], event.Value}
				case ABS_MT_TRACKING_ID:
					if event.Value == -1 {
						id_statuses[last_id] = false
					} else {
						id_statuses[last_id] = true
					}
				case ABS_MT_SLOT:
					last_id = event.Value
				}
			}
			for i := 0; i < 10; i++ {
				if copy_id_statuses[i] != id_statuses[i] {
					if id_statuses[i] {
						x, y := translate_xy(pos_s[i][0], pos_s[i][1])
						id_2_vid[i] = self.touch_require(x, y, false)
						logger.Debugf("mixTouch\trequire\t[%d] translate_xy(%d,%d) => (%d,%d)", i, pos_s[i][0], pos_s[i][1], x, y)
					} else {
						self.touch_release(id_2_vid[i])
						logger.Debugf("mixTouch\trelease\t[%d] ", i)
					}
				} else {
					if pos_s[i][0] != copy_pos_s[i][0] || pos_s[i][1] != copy_pos_s[i][1] {
						x, y := translate_xy(pos_s[i][0], pos_s[i][1])
						self.touch_move(id_2_vid[i], x, y, false)
						logger.Debugf("mixTouch\tmove\t[%d] translate_xy(%d,%d) => (%d,%d)", i, pos_s[i][0], pos_s[i][1], x, y)
					}
				}
			}
		}
	}
}

func (self *TouchHandler) handel_event() {
	for {
		key_events := make([]*evdev.Event, 0)
		abs_events := make([]*evdev.Event, 0)
		var x int32 = 0
		var y int32 = 0
		var HWhell int32 = 0
		var Wheel int32 = 0
		select {
		case <-global_close_signal:
			return
		case event_pack := <-self.events:
			for _, event := range event_pack.events {
				switch event.Type {
				case evdev.EventKey:
					key_events = append(key_events, event)
				case evdev.EventAbsolute:
					abs_events = append(abs_events, event)
				case evdev.EventRelative:
					switch event.Code {
					case uint16(evdev.RelativeX):
						x = event.Value
					case uint16(evdev.RelativeY):
						y = event.Value
					case uint16(evdev.RelativeHWheel):
						HWhell = event.Value
					case uint16(evdev.RelativeWheel):
						Wheel = event.Value
					}
				}
			}

			if x != 0 || y != 0 || HWhell != 0 || Wheel != 0 {
				self.handel_rel_event(x, y, HWhell, Wheel)
			}
			if len(key_events) != 0 {
				self.handel_key_events(key_events, event_pack.dev_type, event_pack.dev_name)
			}
			if len(abs_events) != 0 {
				self.handel_abs_events(abs_events, event_pack.dev_type, event_pack.dev_name)
			}
		}
	}
}