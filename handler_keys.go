package main

import (
	"time"

	"github.com/bitly/go-simplejson"
)

// Version: V3.5.0

// switch_map_mode 切换映射模式 (开/关)
func (self *TouchHandler) switch_map_mode(force_state ...bool) {
	var target_state bool
	if len(force_state) > 0 {
		target_state = force_state[0]
		if self.map_on == target_state {
			return
		}
	} else {
		target_state = !self.map_on
	}

	self.total_move_x = 0
	self.total_move_y = 0

	self.pointer_is_out_temp = false
	self.pointer_switch_key_status = make(map[string]bool)

	if !target_state {
		self.view_lock.Lock()
		if self.view_id >= 0 {
			self.view_id = self.touch_release(self.view_id)
		}
		self.view_id = -1
		self.view_resetting_lock = false
		
		// [V3.4.6] 彻底关闭映射时，强制重置视角坐标和锚点，防止下次开启映射时发生偏移乱点
		self.view_anchor_x = self.view_init_x
		self.view_anchor_y = self.view_init_y
		self.reset_view_position(self.view_init_x, self.view_init_y)
		
		self.view_lock.Unlock()
	}

	self.real_key_down_state.Range(func(key, value interface{}) bool {
		if code, ok := friendly_name_2_keycode[key.(string)]; ok {
			self.u_input_control(UInput_key_event, int32(code), UP)
		}
		self.real_key_down_state.Delete(key)
		return true
	})

	should_reset_wasd := true

	if !target_state {
		if self.wheel_temp_penetration_enable {
			self.wheel_lock.Lock()
			if self.wheel_id != -1 {
				self.wheel_penetrating = true
				should_reset_wasd = false
				logger.Info("进入轮盘临时穿透模式")
			}
			self.wheel_lock.Unlock()
		}
	} else {
		if self.wheel_penetrating {
			self.wheel_penetrating = false
			should_reset_wasd = false
			logger.Info("轮盘临时穿透结束 (切回映射)")
		}
	}

	if should_reset_wasd {
		for i := range self.wasd_up_down_statues {
			self.wasd_up_down_statues[i] = false
		}
		self.shift_state = false
		self.wheel_penetrating = false
	}

	if should_reset_wasd {
		self.wheel_lock.Lock()
		if self.wheel_id != -1 {
			self.wheel_id = self.touch_release(self.wheel_id)
		}
		self.wheel_star_x = self.wheel_init_x
		self.wheel_star_y = self.wheel_init_y
		self.wheel_id = -1
		self.wheel_lock.Unlock()

		self.wasd_wheel_released = true
		self.ls_wheel_released = true
	}

	self.key_action_state_save.Range(func(key, value interface{}) bool {
		if ch, ok := value.(chan bool); ok {
			select {
			case ch <- true:
			default:
			}
		}
		if tid, ok := value.(int32); ok {
			self.touch_release(tid)
		}
		self.key_action_state_save.Delete(key)
		return true
	})

	if self.combo_handler != nil {
		self.combo_handler.ResetState()
	}

	self.map_on = target_state
	self.map_switch_signal <- self.map_on

	if self.map_on {
		logger.Info("映射[on]")
	} else {
		logger.Info("映射[off]")
	}
}

// [V3.4.6] 强力垃圾回收：切出指针前，强制清除鼠标键与等价键在映射表中遗留的按下状态与触点
func (self *TouchHandler) release_mouse_and_equiv_keys_state() {
	keys_to_check := []string{"BTN_LEFT", "BTN_RIGHT", "BTN_MIDDLE", "BTN_SIDE", "BTN_EXTRA"}

	if vmouse_cfg, ok := self.config.CheckGet("V_MOUSE_SETTINGS"); ok {
		if arr, err := vmouse_cfg.Get("LEFT_CLICK_KEYS").StringArray(); err == nil {
			keys_to_check = append(keys_to_check, arr...)
		}
		if arr, err := vmouse_cfg.Get("RIGHT_CLICK_KEYS").StringArray(); err == nil {
			keys_to_check = append(keys_to_check, arr...)
		}
	}

	for _, k := range keys_to_check {
		if state, ok := self.key_action_state_save.Load(k); ok {
			if tid, ok2 := state.(int32); ok2 {
				self.touch_release(tid)
			} else if ch, ok2 := state.(chan bool); ok2 {
				select {
				case ch <- true:
				default:
				}
			}
			self.key_action_state_save.Delete(k)
		}
	}
}

func (self *TouchHandler) handel_key_up_down(key_name string, up_down int32, dev_name string) {
	if key_name == "" {
		return
	}

	if dev_name != "VirtualCombo" && self.combo_handler != nil {
		self.combo_handler.HandlePhysicalInput(key_name, up_down)
	}
	
	// [V3.5.0] 快捷退出拦截，优先级最高 (收到按下信号直接触发自杀)
	if self.end_exit_enable && up_down == DOWN {
		for _, exitKey := range self.end_exit_keys {
			if key_name == exitKey {
				go TriggerSafeExit()
				return
			}
		}
	}

	if self.pointer_is_out_temp {
		is_mouse_button := (key_name == "BTN_LEFT" || key_name == "BTN_RIGHT" || key_name == "BTN_MIDDLE" || key_name == "BTN_SIDE" || key_name == "BTN_EXTRA")

		is_equ_left := false
		is_equ_right := false

		vmouse_cfg := self.config.Get("V_MOUSE_SETTINGS")
		for _, k := range vmouse_cfg.Get("LEFT_CLICK_KEYS").MustStringArray() {
			if k == key_name {
				is_equ_left = true
				break
			}
		}
		for _, k := range vmouse_cfg.Get("RIGHT_CLICK_KEYS").MustStringArray() {
			if k == key_name {
				is_equ_right = true
				break
			}
		}

		if is_mouse_button || is_equ_left || is_equ_right {
			var fake_btn int32
			if is_equ_left || key_name == "BTN_LEFT" {
				fake_btn = 0x110
			} else if is_equ_right || key_name == "BTN_RIGHT" {
				fake_btn = 0x111
			} else {
				fake_btn_u, _ := friendly_name_2_keycode[key_name]
				fake_btn = int32(fake_btn_u)
			}

			self.u_input_control(UInput_mouse_btn, fake_btn, up_down)
			return
		}
	}

	if self.macro_handler != nil {
		self.macro_handler.HandleInput(key_name, up_down)
	}

	if key_name == "BTN_SELECT" {
		if up_down == DOWN || up_down == UP {
			self.BTN_SELECT_UP_DOWN = up_down
		}
	}
	if self.BTN_SELECT_UP_DOWN == DOWN {
		if key_name == "BTN_RS" && up_down == UP {
			self.switch_map_mode()
			return
		}
	}

	if self.KEYBOARD_SWITCH_KEY_NAME_S[key_name] {
		if up_down == UP {
			self.switch_map_mode()
		}
	}

	if self.POINTER_SWITCH_KEY_NAME_S[key_name] && self.map_on {
		if up_down == DOWN {
			self.pointer_switch_key_status[key_name] = true
			if !self.pointer_is_out_temp {
				// [V3.4.6] 切出前，强制洗地鼠标按键状态
				self.release_mouse_and_equiv_keys_state()

				self.pointer_is_out_temp = true
				self.map_switch_signal <- false

				self.view_lock.Lock()
				if self.view_id >= 0 {
					self.view_id = self.touch_release(self.view_id)
				}
				
				// [V3.4.6] 按键切出指针时，强制重置视角坐标和锚点归中
				self.view_anchor_x = self.view_init_x
				self.view_anchor_y = self.view_init_y
				self.reset_view_position(self.view_init_x, self.view_init_y)
				
				self.view_lock.Unlock()

				self.scroll_slider_lock.Lock()
				if self.scroll_slider_id >= 0 {
					self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
				}
				self.scroll_slider_lock.Unlock()
			}
		} else if up_down == UP {
			delete(self.pointer_switch_key_status, key_name)
			if len(self.pointer_switch_key_status) == 0 {
				self.pointer_is_out_temp = false
				self.map_switch_signal <- true
			}
		}
	}

	recoil_cfg := self.config.Get("GLOBAL_RECOIL")
	if recoil_cfg.Get("ENABLE").MustBool(false) {
		for _, resetKey := range recoil_cfg.Get("RESET_SPEED_KEYS").MustStringArray() {
			if key_name == resetKey && up_down == DOWN {
				self.current_recoil_speed = self.base_recoil_speed
				logger.Info("压枪速度已重置")
			}
		}
		isTrigger := false
		for _, k := range recoil_cfg.Get("TRIGGER_KEYS").MustStringArray() {
			if key_name == k {
				self.recoil_trigger_status[k] = (up_down == DOWN)
				isTrigger = true
				break
			}
		}
		if !isTrigger {
			for _, k := range recoil_cfg.Get("SCOPE_KEYS").MustStringArray() {
				if key_name == k {
					self.recoil_scope_status[k] = (up_down == DOWN)
					break
				}
			}
		}
		triggerPressed := false
		for _, v := range self.recoil_trigger_status {
			if v {
				triggerPressed = true
				break
			}
		}
		scopePressed := false
		for _, v := range self.recoil_scope_status {
			if v {
				scopePressed = true
				break
			}
		}

		if triggerPressed {
			if !self.recoil_scope_mode || scopePressed {
				self.recoil_active = true
			} else {
				self.recoil_active = false
			}
		} else {
			self.recoil_active = false
		}
	}

	action, has_action := self.config.Get("KEY_MAPS").CheckGet(key_name)

	handle_shift := func() bool {
		if self.wheel_shift_enable && key_name == "KEY_LEFTSHIFT" {
			if self.shift_press_toggle {
				if up_down == DOWN {
					self.shift_state = !self.shift_state
				}
			} else if self.shift_release_toggle {
				if up_down == UP {
					self.shift_state = !self.shift_state
				}
			} else {
				if up_down == DOWN {
					self.shift_state = true
				} else if up_down == UP {
					self.shift_state = false
				}
			}
			self.wasd_up_down_statues[4] = self.shift_state
			return true
		}
		return false
	}

	handle_wasd := func() bool {
		for i := 0; i < 4; i++ {
			if self.wheel_wasd[i] == key_name {
				if up_down == DOWN {
					self.wasd_up_down_statues[i] = true
				} else if up_down == UP {
					self.wasd_up_down_statues[i] = false
				}
				self.wheel_last_input_time = time.Now()
				return true
			}
		}
		return false
	}

	if self.map_on {
		if handle_wasd() {
			return
		}
		if handle_shift() {
			return
		}

		if has_action {
			state, ok := self.key_action_state_save.Load(key_name)
			action_type := action.Get("TYPE").MustString()

			need_up_event := map[string]bool{
				"PRESS":               true,
				"AUTO_FIRE":           true,
				"MULT_PRESS":          true,
				"SYNC_VIEW_RESET":     true,
				"SYNC_BACKPACK":       true,
				"PRESS_RELEASE_CLICK": true,
				"SMART_TOGGLE":        true,
			}

			allow_reentry := map[string]bool{
				"SEQUENTIAL_PRESS": true,
				"CLICK_VIEW_RESET": true,
				"BACKPACK_TOGGLE":  true,
				"SMART_TOGGLE":     true, 
			}

			if (up_down == UP && !ok && !need_up_event[action_type]) ||
				(up_down == DOWN && ok && action_type != "PRESS" && !allow_reentry[action_type]) {
			} else {
				self.execute_key_action(time.Now(), key_name, up_down, action, state)
			}
		}

	} else {
		if self.wheel_penetrating {
			if handle_wasd() {
				return
			}
			if handle_shift() {
				return
			}
		}

		if has_action {
			action_type := action.Get("TYPE").MustString()
			if action_type == "RECOIL_SPEED_SET" {
				state, _ := self.key_action_state_save.Load(key_name)
				self.execute_key_action(time.Now(), key_name, up_down, action, state)
			}
		}

		if jsconfig, ok := self.joystickInfo[dev_name]; ok {
			if joystick_btn_map_key_name, ok := jsconfig.Get("MAP_KEYBOARD").CheckGet(key_name); ok {
				self.handel_key_up_down(joystick_btn_map_key_name.MustString(), up_down, dev_name+"_joystick_mapped")
			}
		} else {
			if code, ok := friendly_name_2_keycode[key_name]; ok {
				self.u_input_control(UInput_key_event, int32(code), int32(up_down))
				if up_down == DOWN {
					self.real_key_down_state.Store(key_name, true)
				} else {
					self.real_key_down_state.Delete(key_name)
				}
			}
		}
	}
}

func (self *TouchHandler) execute_key_action(start time.Time, key_name string, up_down int32, action *simplejson.Json, state interface{}) {
	action_type := action.Get("TYPE").MustString()
	isWheelKey := key_name == "REL_WHEEL_DOWN" || key_name == "REL_WHEEL_UP" || key_name == "REL_HWHEEL_DOWN" || key_name == "REL_HWHEEL_UP"
	if isWheelKey {
		if action_type == "PRESS" || action_type == "AUTO_FIRE" {
			return
		}
	}

	defer logger.Debugf("key[%s]%s\t%v\t%v", key_name, UDF[up_down], action, time.Since(start))

	switch action_type {
	case "PRESS":
		if up_down == DOWN {
			x := int32(action.Get("POS").GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
			y := int32(action.Get("POS").GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
			x_jit, y_jit := self.apply_key_jitter(x, y)
			self.key_action_state_save.Store(key_name, self.touch_require(x_jit, y_jit, true))
		} else if up_down == UP {
			if tid, ok := state.(int32); ok {
				self.touch_release(tid)
			}
			self.key_action_state_save.Delete(key_name)
		}

	case "CLICK":
		if up_down == DOWN {
			go (func() {
				action_check, ok := self.config.Get("KEY_MAPS").CheckGet(key_name)
				if !ok {
					return
				}
				duration := 18
				if interval := action_check.Get("INTERVAL").MustArray(); len(interval) > 0 {
					duration = action_check.Get("INTERVAL").GetIndex(0).MustInt(18)
				}
				x := int32(action_check.Get("POS").GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(action_check.Get("POS").GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})()
		}

	case "AUTO_FIRE":
		if up_down == DOWN {
			if _, ok := self.key_action_state_save.Load(key_name); ok {
				return
			}
			cancel_ch := make(chan bool, 1)
			self.key_action_state_save.Store(key_name, cancel_ch)

			x := int32(action.Get("POS").GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
			y := int32(action.Get("POS").GetIndex(1).MustFloat64() * float64(self.rel_screen_y))

			go (func() {
				for {
					select {
					case <-cancel_ch:
						return
					default:
					}

					action_check, ok := self.config.Get("KEY_MAPS").CheckGet(key_name)
					if !ok {
						return
					}
					interval_array := action_check.Get("INTERVAL").MustArray()
					max_dur, max_int, min_dur, min_int := 18, 20, 10, 10
					if len(interval_array) >= 2 {
						max_dur = action_check.Get("INTERVAL").GetIndex(0).MustInt(18)
						max_int = action_check.Get("INTERVAL").GetIndex(1).MustInt(20)
					}
					if len(interval_array) >= 4 {
						min_dur = action_check.Get("INTERVAL").GetIndex(2).MustInt(max_dur / 2)
						min_int = action_check.Get("INTERVAL").GetIndex(3).MustInt(max_int / 2)
					}
					down_time := self.get_random_duration(max_dur, min_dur)
					interval_time := self.get_random_duration(max_int, min_int)
					x_jit, y_jit := self.apply_key_jitter(x, y)
					tid := self.touch_require(x_jit, y_jit, true)

					time.Sleep(down_time)
					self.touch_release(tid)

					select {
					case <-cancel_ch:
						return
					case <-time.After(interval_time):
					}
				}
			})()

		} else if up_down == UP {
			if ch, ok := state.(chan bool); ok {
				select {
				case ch <- true:
				default:
				}
			}
			self.key_action_state_save.Delete(key_name)
		}

	case "MULT_PRESS":
		if up_down == DOWN {
			if _, ok := self.key_action_state_save.Load(key_name); ok {
				return
			}
			cancel_ch := make(chan bool, 1)
			self.key_action_state_save.Store(key_name, cancel_ch)

			go (func() {
				action_check, ok := self.config.Get("KEY_MAPS").CheckGet(key_name)
				if !ok {
					return
				}
				pos_s := action_check.Get("POS_S").MustArray()
				if len(pos_s) == 0 {
					return
				}
				interval := 0
				if len(action_check.Get("INTERVAL").MustArray()) > 0 {
					interval = action_check.Get("INTERVAL").GetIndex(0).MustInt(0)
				}

				tid_list := make([]int32, 0)
				defer func() {
					for _, tid := range tid_list {
						self.touch_release(tid)
					}
				}()

				for i := range pos_s {
					select {
					case <-cancel_ch:
						return
					default:
					}

					if i > 0 && interval > 0 {
						select {
						case <-cancel_ch:
							return
						case <-time.After(time.Duration(interval) * time.Millisecond):
						}
					}
					x := int32(action_check.Get("POS_S").GetIndex(i).GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
					y := int32(action_check.Get("POS_S").GetIndex(i).GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
					x_jit, y_jit := self.apply_key_jitter(x, y)
					tid := self.touch_require(x_jit, y_jit, true)
					tid_list = append(tid_list, tid)
				}
				<-cancel_ch
			})()
		} else if up_down == UP {
			if ch, ok := state.(chan bool); ok {
				select {
				case ch <- true:
				default:
				}
			}
			self.key_action_state_save.Delete(key_name)
		}

	case "DRAG":
		if up_down == DOWN {
			go (func() {
				action, ok := self.config.Get("KEY_MAPS").CheckGet(key_name)
				if !ok {
					return
				}
				pos_s := action.Get("POS_S").MustArray()
				if len(pos_s) < 2 {
					return
				}
				interval_time := action.Get("INTERVAL").GetIndex(0).MustInt(18)
				init_x := int32(action.Get("POS_S").GetIndex(0).GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				init_y := int32(action.Get("POS_S").GetIndex(0).GetIndex(1).MustFloat64() * float64(self.rel_screen_y))

				tid := self.touch_require(init_x, init_y, true)
				defer self.touch_release(tid)

				time.Sleep(time.Duration(interval_time) * time.Millisecond)
				for index := 1; index < len(pos_s); index++ {
					x := int32(action.Get("POS_S").GetIndex(index).GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
					y := int32(action.Get("POS_S").GetIndex(index).GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
					self.touch_move(tid, x, y, true)
					time.Sleep(time.Duration(interval_time) * time.Millisecond)
				}
			})()
		}

	case "SYNC_VIEW_RESET":
		if up_down == DOWN {
			self.view_lock.Lock()
			self.key_action_state_save.Store(key_name, true)
			self.view_resetting_lock = true
			if self.view_id >= 0 {
				self.view_id = self.touch_release(self.view_id)
			}
			self.view_id = -1
			new_x := int32(action.Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe)
			new_y := int32(action.Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe)
			
			// [V3.4.6] 触发去程，动态变更锚点，使甜甜圈跟随
			self.view_anchor_x = new_x
			self.view_anchor_y = new_y
			self.reset_view_position(new_x, new_y, 0.5)
			
			self.delayed_view_require()
			self.view_lock.Unlock()
		} else if up_down == UP {
			if _, ok := state.(bool); ok {
				self.view_lock.Lock()
				self.view_resetting_lock = false
				if self.view_id >= 0 {
					self.view_id = self.touch_release(self.view_id)
				}
				self.view_id = -1
				
				// [V3.4.6] 触发回程，动态还原锚点，使甜甜圈回到初始化中心
				self.view_anchor_x = self.view_init_x
				self.view_anchor_y = self.view_init_y
				self.reset_view_position(self.view_init_x, self.view_init_y, 1.0)
				
				self.delayed_view_require()
				self.view_lock.Unlock()
			}
			self.key_action_state_save.Delete(key_name)
		}

	case "CLICK_VIEW_RESET":
		if up_down == DOWN {
			_, loaded := self.key_action_state_save.Load(key_name)
			if !loaded { // 第一次点击：去程
				self.view_lock.Lock()
				self.key_action_state_save.Store(key_name, true)
				self.view_resetting_lock = true
				if self.view_id >= 0 {
					self.view_id = self.touch_release(self.view_id)
				}
				self.view_id = -1
				new_x := int32(action.Get("POS").GetIndex(0).MustFloat64() * 0x7ffffffe)
				new_y := int32(action.Get("POS").GetIndex(1).MustFloat64() * 0x7ffffffe)
				
				// [V3.4.6] 去程变更锚点
				self.view_anchor_x = new_x
				self.view_anchor_y = new_y
				self.reset_view_position(new_x, new_y, 0.5)
				
				self.delayed_view_require()
				self.view_lock.Unlock()
			} else { // 第二次点击：回程
				self.view_lock.Lock()
				self.view_resetting_lock = false
				if self.view_id >= 0 {
					self.view_id = self.touch_release(self.view_id)
				}
				self.view_id = -1
				
				// [V3.4.6] 回程还原锚点
				self.view_anchor_x = self.view_init_x
				self.view_anchor_y = self.view_init_y
				self.reset_view_position(self.view_init_x, self.view_init_y, 1.0)
				
				self.delayed_view_require()
				self.view_lock.Unlock()
				self.key_action_state_save.Delete(key_name)
			}
		}

	case "BACKPACK_TOGGLE":
		if up_down == DOWN {
			var pos_to_click *simplejson.Json
			if !self.pointer_is_out_temp {
				// [V3.4.6] 切出前执行前置清理
				self.release_mouse_and_equiv_keys_state()

				pos_to_click = action.Get("POS")
				self.key_action_state_save.Store(key_name, true)
				self.pointer_is_out_temp = true
				self.map_switch_signal <- false

				self.view_lock.Lock()
				if self.view_id >= 0 {
					self.view_id = self.touch_release(self.view_id)
				}
				
				// [V3.4.6] 切出指针，强制视角归中与锚点还原
				self.view_anchor_x = self.view_init_x
				self.view_anchor_y = self.view_init_y
				self.reset_view_position(self.view_init_x, self.view_init_y)
				
				self.view_lock.Unlock()
				
				self.scroll_slider_lock.Lock()
				if self.scroll_slider_id >= 0 {
					self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
				}
				self.scroll_slider_lock.Unlock()
			} else {
				pos_to_click = action.Get("POS_B")
				if pos_to_click.Interface() == nil {
					pos_to_click = action.Get("POS")
				}
				self.key_action_state_save.Delete(key_name)
				self.pointer_is_out_temp = false
				self.map_switch_signal <- true
			}

			go (func(pos *simplejson.Json) {
				action_check, ok := self.config.Get("KEY_MAPS").CheckGet(key_name)
				if !ok {
					return
				}
				duration := action_check.Get("CLICK_DURATION").MustInt(18)
				if duration == 0 {
					duration = 8
				}
				x := int32(pos.GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(pos.GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				// [V3.4.6] 增加抖动
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})(pos_to_click)
		}

	case "SYNC_BACKPACK":
		if up_down == DOWN {
			if !self.pointer_is_out_temp {
				// [V3.4.6] 切出前执行前置清理
				self.release_mouse_and_equiv_keys_state()

				self.pointer_is_out_temp = true
				self.map_switch_signal <- false
				
				self.view_lock.Lock()
				if self.view_id >= 0 {
					self.view_id = self.touch_release(self.view_id)
				}
				
				// [V3.4.6] 切出指针，强制视角归中与锚点还原
				self.view_anchor_x = self.view_init_x
				self.view_anchor_y = self.view_init_y
				self.reset_view_position(self.view_init_x, self.view_init_y)
				
				self.view_lock.Unlock()
				
				self.scroll_slider_lock.Lock()
				if self.scroll_slider_id >= 0 {
					self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
				}
				self.scroll_slider_lock.Unlock()
			}
			go (func() {
				duration := action.Get("CLICK_DURATION").MustInt(50)
				x := int32(action.Get("POS").GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(action.Get("POS").GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				// [V3.4.6] 增加抖动
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})()
		} else if up_down == UP {
			if self.pointer_is_out_temp {
				self.pointer_is_out_temp = false
				self.map_switch_signal <- true
			}
			go (func() {
				duration := action.Get("CLICK_DURATION").MustInt(50)
				pos_b := action.Get("POS_B")
				if pos_b.Interface() == nil {
					pos_b = action.Get("POS")
				}
				x := int32(pos_b.GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(pos_b.GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				// [V3.4.6] 增加抖动
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})()
		}

	case "SMART_TOGGLE":
		separat := action.Get("SEPARAT").MustBool()
		release_mouse := action.Get("RELEASE_MOUSE").MustBool()
		execTouch := action.Get("TOUCH").MustBool()
		x_start := int32(action.Get("POS_S").GetIndex(0).GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
		y_start := int32(action.Get("POS_S").GetIndex(0).GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
		x_end := int32(action.Get("POS_S").GetIndex(1).GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
		y_end := int32(action.Get("POS_S").GetIndex(1).GetIndex(1).MustFloat64() * float64(self.rel_screen_y))

		if separat {
			if up_down == DOWN {
				_, loaded := self.key_action_state_save.Load(key_name)
				if !loaded {
					self.key_action_state_save.Store(key_name, true)
					if release_mouse && !self.pointer_is_out_temp {
						// [V3.4.6] 切出前执行前置清理
						self.release_mouse_and_equiv_keys_state()

						self.pointer_is_out_temp = true
						self.map_switch_signal <- false
						
						self.view_lock.Lock()
						if self.view_id >= 0 {
							self.view_id = self.touch_release(self.view_id)
						}
						
						// [V3.4.6] 切出指针，强制视角归中与锚点还原
						self.view_anchor_x = self.view_init_x
						self.view_anchor_y = self.view_init_y
						self.reset_view_position(self.view_init_x, self.view_init_y)
						
						self.view_lock.Unlock()
						
						self.scroll_slider_lock.Lock()
						if self.scroll_slider_id >= 0 {
							self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
						}
						self.scroll_slider_lock.Unlock()
					}
					if execTouch {
						go (func() {
							// [V3.4.6] 增加抖动
							x_jit, y_jit := self.apply_key_jitter(x_start, y_start)
							tid := self.touch_require(x_jit, y_jit, true)
							time.Sleep(time.Duration(8) * time.Millisecond)
							self.touch_release(tid)
						})()
					}
				} else {
					self.key_action_state_save.Delete(key_name)
					if release_mouse && self.pointer_is_out_temp {
						self.pointer_is_out_temp = false
						self.map_switch_signal <- true
					}
					if execTouch {
						go (func() {
							// [V3.4.6] 增加抖动
							x_jit, y_jit := self.apply_key_jitter(x_end, y_end)
							tid := self.touch_require(x_jit, y_jit, true)
							time.Sleep(time.Duration(8) * time.Millisecond)
							self.touch_release(tid)
						})()
					}
				}
			}
		} else {
			if up_down == DOWN {
				if release_mouse && !self.pointer_is_out_temp {
					// [V3.4.6] 切出前执行前置清理
					self.release_mouse_and_equiv_keys_state()

					self.pointer_is_out_temp = true
					self.map_switch_signal <- false
					
					self.view_lock.Lock()
					if self.view_id >= 0 {
						self.view_id = self.touch_release(self.view_id)
					}
					
					// [V3.4.6] 切出指针，强制视角归中与锚点还原
					self.view_anchor_x = self.view_init_x
					self.view_anchor_y = self.view_init_y
					self.reset_view_position(self.view_init_x, self.view_init_y)
					
					self.view_lock.Unlock()
					
					self.scroll_slider_lock.Lock()
					if self.scroll_slider_id >= 0 {
						self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
					}
					self.scroll_slider_lock.Unlock()
				}
				if execTouch {
					go (func() {
						// [V3.4.6] 增加抖动
						x_jit, y_jit := self.apply_key_jitter(x_start, y_start)
						tid := self.touch_require(x_jit, y_jit, true)
						time.Sleep(time.Duration(8) * time.Millisecond)
						self.touch_release(tid)
					})()
				}
			} else if up_down == UP {
				if release_mouse && self.pointer_is_out_temp {
					self.pointer_is_out_temp = false
					self.map_switch_signal <- true
				}
				if execTouch {
					go (func() {
						// [V3.4.6] 增加抖动
						x_jit, y_jit := self.apply_key_jitter(x_end, y_end)
						tid := self.touch_require(x_jit, y_jit, true)
						time.Sleep(time.Duration(8) * time.Millisecond)
						self.touch_release(tid)
					})()
				}
			}
		}

	case "SEQUENTIAL_PRESS":
		if up_down == DOWN {
			var all_pos []*simplejson.Json
			if pos := action.Get("POS"); pos.Interface() != nil {
				all_pos = append(all_pos, pos)
			}
			if pos_s_list, ok := action.CheckGet("POS_S"); ok {
				for i := range pos_s_list.MustArray() {
					all_pos = append(all_pos, action.Get("POS_S").GetIndex(i))
				}
			}
			if len(all_pos) == 0 {
				return
			}

			last_time, has_last := self.key_action_state_save.Load(key_name + "_time")
			if has_last {
				cooldown := action.Get("COOLDOWN").MustInt(0)
				if cooldown > 0 && time.Since(last_time.(time.Time)) < time.Duration(cooldown)*time.Millisecond {
					return
				}
			}
			current_index := 0
			if idx_val, ok := self.key_action_state_save.Load(key_name + "_seq_index"); ok {
				current_index = idx_val.(int)
			}
			if current_index >= len(all_pos) {
				current_index = 0
			}
			pos_to_click := all_pos[current_index]
			next_index := (current_index + 1) % len(all_pos)

			self.key_action_state_save.Store(key_name+"_seq_index", next_index)
			self.key_action_state_save.Store(key_name, true)
			self.key_action_state_save.Store(key_name+"_time", time.Now())

			go (func() {
				action_check, ok := self.config.Get("KEY_MAPS").CheckGet(key_name)
				if !ok {
					return
				}
				duration := action_check.Get("CLICK_DURATION").MustInt(18)
				if duration == 0 {
					duration = 8
				}
				x := int32(pos_to_click.GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(pos_to_click.GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				// [V3.4.6] 增加抖动
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})()
		}

	case "PRESS_RELEASE_CLICK":
		if up_down == DOWN {
			go (func() {
				duration := action.Get("CLICK_DURATION").MustInt(50)
				x := int32(action.Get("POS").GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(action.Get("POS").GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				// [V3.4.6] 增加抖动
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})()
		} else if up_down == UP {
			go (func() {
				duration := action.Get("CLICK_DURATION").MustInt(50)
				x := int32(action.Get("POS_B").GetIndex(0).MustFloat64() * float64(self.rel_screen_x))
				y := int32(action.Get("POS_B").GetIndex(1).MustFloat64() * float64(self.rel_screen_y))
				// [V3.4.6] 增加抖动
				x_jit, y_jit := self.apply_key_jitter(x, y)
				tid := self.touch_require(x_jit, y_jit, true)
				defer self.touch_release(tid)
				time.Sleep(time.Duration(duration) * time.Millisecond)
			})()
		}

	case "RECOIL_SPEED_SET":
		if up_down == DOWN {
			target_speed := action.Get("VALUE").MustFloat64()
			self.current_recoil_speed = target_speed
			logger.Infof("压枪速度热更新: %v", target_speed)
		}
	}
}