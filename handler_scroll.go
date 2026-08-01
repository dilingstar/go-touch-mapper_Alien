package main

import (
	"math"
	"math/rand"
	"time"
)

// Version: V3.4.5

// loop_auto_release_scroll_slider 滚轮滑块自动释放协程
// 如果一段时间没有滚动输入，自动释放触点，模拟手指离开屏幕
func (self *TouchHandler) loop_auto_release_scroll_slider() {
	for {
		select {
		case <-global_close_signal:
			return
		default:
			self.scroll_slider_lock.Lock()
			if self.scroll_slider_id >= 0 {
				// 使用配置的延迟释放时间
				delay := time.Duration(self.scroll_slider_release_delay_ms) * time.Millisecond
				if delay < 20*time.Millisecond {
					delay = 20 * time.Millisecond // 最小防抖
				}

				if time.Since(self.scroll_slider_last_scroll_time) > delay {
					self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
				}
			}
			self.scroll_slider_lock.Unlock()
			time.Sleep(time.Duration(20) * time.Millisecond)
		}
	}
}

// quick_click 快速点击 (用于普通滚轮映射，如切换武器)
func (self *TouchHandler) quick_click(keyname string) {
	self.handel_key_up_down(keyname, DOWN, "MOUSE_WHEEL")
	time.Sleep(time.Duration(50) * time.Millisecond)
	self.handel_key_up_down(keyname, UP, "MOUSE_WHEEL")
}

// handel_scroll_slider 处理滚轮滑块逻辑
// direction: 滚动方向 (-1 或 1)，对应滚轮的上下
func (self *TouchHandler) handel_scroll_slider(direction int32) {

	if !self.scroll_slider_enable {
		// 未启用滑块模式，回退到普通按键映射 (REL_WHEEL_UP/DOWN)
		if direction < 0 {
			go self.quick_click("REL_WHEEL_UP")
		} else if direction > 0 {
			go self.quick_click("REL_WHEEL_DOWN")
		}
		return
	}

	self.scroll_slider_lock.Lock()
	defer self.scroll_slider_lock.Unlock()

	self.scroll_slider_last_scroll_time = time.Now()

	// 1. 超时重置位置 (模拟手指抬起后重新放回中心)
	if time.Since(self.scroll_slider_last_reset_time) > self.scroll_slider_timeout_duration {
		self.scroll_slider_current_y = self.scroll_slider_init_y
	}
	self.scroll_slider_last_reset_time = time.Now()

	final_x, final_y := self.scroll_slider_init_x, self.scroll_slider_current_y

	is_new_press := (self.scroll_slider_id == -1)

	// 2. 随机落点 (防检测) - 仅在起步时
	if is_new_press && self.scroll_slider_random_enable {
		offset_x, offset_y := self.get_random_offset(self.scroll_slider_random_radius_px)
		final_x += offset_x
		final_y += offset_y
	}

	// 3. 计算速度 (直接使用固定速度)
	current_speed_px := self.scroll_slider_speed_px

	// 4. 计算目标Y坐标
	target_y := final_y + (direction * current_speed_px)

	// 5. 边界检查 (撞墙检测)
	boundary_hit := false
	if target_y < self.scroll_slider_bound_up {
		target_y = self.scroll_slider_init_y
		boundary_hit = true
	} else if target_y > self.scroll_slider_bound_down {
		target_y = self.scroll_slider_init_y
		boundary_hit = true
	}

	self.scroll_slider_current_y = target_y
	final_y = target_y

	// 6. 应用曲线偏移 (使滑动轨迹不是直线)
	if self.scroll_slider_curve_enable {
		curve_calc_angle := (float64(final_y-self.scroll_slider_init_y) * 0.1) * self.scroll_slider_curve_freq
		curve_offset := math.Sin(curve_calc_angle) * float64(self.scroll_slider_curve_amount_px)
		final_x += int32(curve_offset)
	}

	// 7. 应用动态噪点
	if self.scroll_slider_noise_enable {
		noise_range := self.scroll_slider_noise_intensity * float64(self.rel_screen_x)
		final_x += int32((rand.Float64() - 0.5) * noise_range)
	}

	// 8. 执行触控操作
	if is_new_press || boundary_hit {
		// 如果撞墙了，且当前有触点，先释放旧触点
		if boundary_hit && self.scroll_slider_id >= 0 {
			self.scroll_slider_id = self.touch_release(self.scroll_slider_id)
		}

		self.scroll_slider_id = -2 // 标记为等待中 (死区)
		
		// 异步延迟按下 (模拟撞墙后的停顿)
		go (func(x, y int32, is_boundary_hit bool) {
			var delay time.Duration
			if is_boundary_hit {
				delay = self.get_delay_ms(
					self.scroll_slider_delay_reset_ms,
					self.scroll_slider_delay_reset_min_ms,
					self.scroll_slider_delay_random_enable,
				)
			} else {
				delay = 0
			}

			if delay > 0 {
				time.Sleep(delay)
			}

			self.scroll_slider_lock.Lock()
			defer self.scroll_slider_lock.Unlock()

			if self.scroll_slider_id != -2 {
				return // 状态已改变，取消操作
			}

			self.scroll_slider_id = self.touch_require(x, y, true)
		})(final_x, final_y, boundary_hit)

	} else {
		// 正常移动
		if self.scroll_slider_id >= 0 {
			self.touch_move(self.scroll_slider_id, final_x, final_y, true)
		}
	}
}

// handel_rel_event 处理相对移动事件 (鼠标移动、滚轮)
// 这是相对输入设备的分发入口
func (self *TouchHandler) handel_rel_event(x int32, y int32, HWhell int32, Wheel int32) {
	// =======================================================================
	// [V3.4.5] 如果处于临时切出指针状态，强行拦截所有鼠标相对事件给 v_mouse
	// 完美解决背包内鼠标移动乱转视角，滚动乱划屏幕的冲突
	// =======================================================================
	if self.pointer_is_out_temp {
		if x != 0 || y != 0 {
			self.u_input_control(UInput_mouse_move, x, y)
		}
		if HWhell != 0 {
			self.u_input_control(UInput_mouse_wheel, REL_HWHEEL, HWhell)
		}
		if Wheel != 0 {
			self.u_input_control(UInput_mouse_wheel, REL_WHEEL, Wheel)
		}
		return // 拦截完毕，直接退出
	}

	// 1. 处理鼠标移动 -> 转发给视角模块
	if x != 0 || y != 0 {
		if self.map_on {
			self.handel_view_move(x, y)
		} else {
			self.u_input_control(UInput_mouse_move, x, y)
		}
	}

	// 2. 处理水平滚轮 (HWheel)
	if HWhell != 0 {
		if self.map_on {
			if HWhell > 0 {
				go self.quick_click("REL_HWHEEL_UP")
			} else if HWhell < 0 {
				go self.quick_click("REL_HWHEEL_DOWN")
			}
		} else {
			self.u_input_control(UInput_mouse_wheel, REL_HWHEEL, HWhell)
		}
	}

	// 3. 处理垂直滚轮 (Wheel) -> 转发给滑块模块 或 按键模块
	if Wheel != 0 {
		if self.map_on {
			var key_name string
			if Wheel < 0 {
				key_name = "REL_WHEEL_DOWN"
			} else {
				key_name = "REL_WHEEL_UP"
			}

			// 检查是否配置了特殊按键类型 (如多点映射，则优先触发映射)
			if action, ok := self.config.Get("KEY_MAPS").CheckGet(key_name); ok {
				action_type := action.Get("TYPE").MustString()

				// 只要是按键映射类型都走按键逻辑，不走滑块
				if action_type != "" {
					state, _ := self.key_action_state_save.Load(key_name)
					self.execute_key_action(time.Now(), key_name, DOWN, action, state)
				} else {
					go self.handel_scroll_slider(Wheel * -1)
				}
			} else {
				go self.handel_scroll_slider(Wheel * -1)
			}

		} else {
			self.u_input_control(UInput_mouse_wheel, REL_WHEEL, Wheel)
		}
	}
}