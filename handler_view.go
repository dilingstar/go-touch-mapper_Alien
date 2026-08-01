package main

import (
	"time"
)

// Version: V3.4.6

// handel_view_move 处理视角移动
// offset_x, offset_y: 移动的相对距离
func (self *TouchHandler) handel_view_move(offset_x int32, offset_y int32) {
	self.view_lock.Lock()
	defer self.view_lock.Unlock()

	// 正在重置中(初始化申请阶段)，忽略输入
	if self.view_id == -2 {
		return
	}

	if self.measure_sensitivity_mode {
		self.total_move_x += offset_x
		self.total_move_y += offset_y
		logger.Infof("total_move_x:%v\ttotal_move_y:%v", self.total_move_x, self.total_move_y)
	}
	self.auto_release_view_counter = 0

	// 如果当前没有持有视角触点，尝试申请
	if self.view_id == -1 {
		self.delayed_view_require()
		return 
	}

	// 计算下一帧的坐标 (使用高精度坐标 0x7ffffffe)
	next_x := int64(self.view_current_x) + int64(offset_x)*int64(self.view_speed_x)
	next_y := int64(self.view_current_y) + int64(offset_y)*int64(self.view_speed_y)

	var reset_required bool = false

	if self.view_range_limited { 
		// 1. 边界检查 (屏幕边缘)
		if next_x <= 0 || next_x >= int64(self.screen_x) || next_y <= 0 || next_y >= int64(self.screen_y) {
			reset_required = true
		}

		// 2. 视角重置半径检查 (甜甜圈区域)
		if !reset_required && self.view_reset_radius_enable {
			// [V3.4.6] 半径检测的中心基准修改为动态锚点 view_anchor，使得甜甜圈跟随重置键移动
			diff_x := next_x - int64(self.view_anchor_x)
			diff_y := next_y - int64(self.view_anchor_y)

			// 转换回像素单位进行半径比较
			dist_x_px := (diff_x * int64(self.rel_screen_x)) / 0x7ffffffe
			dist_y_px := (diff_y * int64(self.rel_screen_y)) / 0x7ffffffe

			dist_sq := float64(dist_x_px*dist_x_px + dist_y_px*dist_y_px)

			radius_px := float64(self.view_reset_radius_px)
			
			// 如果处于重置键触发的去程状态(view_resetting_lock=true)，半径减半
			if self.view_resetting_lock {
				radius_px *= 0.6
			}

			thickness_px := float64(self.view_reset_radius_thickness_px)

			inner_radius := radius_px - thickness_px
			if inner_radius < 0 {
				inner_radius = 0
			}
			inner_sq := inner_radius * inner_radius
			outer_sq := radius_px * radius_px

			if dist_sq >= inner_sq && dist_sq <= outer_sq {
				reset_required = true
			}
		}
	}

	if reset_required {
		// 如果处于重置键触发的去程状态，撞墙不重置，而是直接停下(模拟撞墙)
		if self.view_resetting_lock {
			return 
		}

		// 需要重置：释放当前触点 -> 重置坐标 -> 重新申请
		if self.view_id >= 0 {
			self.view_id = self.touch_release(self.view_id)
		}
		self.view_id = -1 

		// [V3.4.6] 撞墙重置的位置，使用动态锚点，不再死板地回到 init 中心
		self.reset_view_position(self.view_anchor_x, self.view_anchor_y)

		self.delayed_view_require()

	} else {
		// 正常移动
		self.view_current_x = int32(next_x)
		self.view_current_y = int32(next_y)

		if self.view_id >= 0 {
			self.touch_move(self.view_id, self.view_current_x, self.view_current_y, false)
		}
	}
}

// delayed_view_require 延迟申请视角触点 (支持随机延迟)
func (self *TouchHandler) delayed_view_require() {
	if self.view_id != -1 {
		return
	}

	self.view_id = -2 // 标记为"正在申请中"

	go (func() {
		var delay time.Duration
		if self.view_delay_reset_enable {
			delay = self.get_delay_ms(
				self.view_delay_reset_ms,
				self.view_delay_reset_min_ms,
				self.view_delay_reset_random_enable,
			)
			time.Sleep(delay)
		}

		self.view_lock.Lock()
		defer self.view_lock.Unlock()

		if self.view_id != -2 {
			return 
		}

		self.view_id = self.touch_require(self.view_current_x, self.view_current_y, false)
	})()
}

// reset_view_position 重置视角坐标 (支持随机位置)
// 增加参数 scale_factor，用于控制随机范围的缩放 (默认1.0，重置键去程时0.5)
func (self *TouchHandler) reset_view_position(new_x int32, new_y int32, scale_factor ...float64) {
	final_x, final_y := new_x, new_y

	if self.view_random_reset_enable {
		scale := 1.0
		if len(scale_factor) > 0 {
			scale = scale_factor[0]
		}
		
		// 应用缩放后的半径
		scaled_radius := int32(float64(self.view_random_reset_radius_px) * scale)
		offset_x, offset_y := self.get_random_offset(scaled_radius)
		
		scale_factor_x := int64(0x7ffffffe) / int64(self.rel_screen_x)
		scale_factor_y := int64(0x7ffffffe) / int64(self.rel_screen_y)

		final_x += int32(int64(offset_x) * scale_factor_x)
		final_y += int32(int64(offset_y) * scale_factor_y)
	}

	self.view_current_x = final_x
	self.view_current_y = final_y
}

// auto_handel_view_release 自动释放视角触点协程
func (self *TouchHandler) auto_handel_view_release() {
	for {
		select {
		case <-global_close_signal:
			return
		default:
			timeout := self.view_auto_release_ms
			enable := self.view_auto_release_enable

			if !enable {
				time.Sleep(time.Duration(200) * time.Millisecond)
				continue
			}
			
			// 如果处于重置键锁定状态，禁止自动释放
			if self.view_resetting_lock {
				time.Sleep(time.Duration(50) * time.Millisecond)
				continue
			}

			if timeout > 0 {
				self.view_lock.Lock()
				if self.view_id >= 0 { 
					self.auto_release_view_counter += 1
					if self.auto_release_view_counter > int32(timeout/50) {
						self.auto_release_view_counter = 0
						self.view_id = self.touch_release(self.view_id)
						// [V3.4.6] 自动释放后，回归点使用动态锚点
						self.reset_view_position(self.view_anchor_x, self.view_anchor_y)
					}
				} else {
					self.auto_release_view_counter = 0
				}
				self.view_lock.Unlock()
			}
			time.Sleep(time.Duration(50) * time.Millisecond) 
		}
	}
}

// loop_handel_rs_move 右摇杆(RS)视角移动协程
func (self *TouchHandler) loop_handel_rs_move() {
	for {
		select {
		case <-global_close_signal:
			return
		default:
			rs_x, rs_y := self.getStick("RS")
			if rs_x != 0.5 || rs_y != 0.5 {
				// [V3.4.5] 增加指针模式拦截
				if self.map_on && !self.pointer_is_out_temp {
					self.handel_view_move(int32((rs_x-0.5)*self.rs_speed_x), int32((rs_y-0.5)*self.rs_speed_y))
				} else {
					self.u_input_control(UInput_mouse_move, int32((rs_x-0.5)*24), int32((rs_y-0.5)*24))
				}
			}
			time.Sleep(time.Duration(4) * time.Millisecond) 
		}
	}
}