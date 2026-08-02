package main

import (
	"math"
	"math/rand"
	"time"
)

// Version: V3.4.6
// handel_wheel_action 执行轮盘的触摸操作 (按下/移动/松开)
// sync_last_pos: 是否同步更新 wasd_wheel_last_x/y (用于手柄摇杆接管)
func (self *TouchHandler) handel_wheel_action(action int8, abs_x int32, abs_y int32, sync_last_pos ...bool) {
	self.wheel_lock.Lock()
	defer self.wheel_lock.Unlock()

	if action == Wheel_action_release {
		if self.wheel_id != -1 {
			self.wheel_id = self.touch_release(self.wheel_id)
		}
	} else if action == Wheel_action_move {
		self.wheel_last_input_time = time.Now()

		if len(sync_last_pos) > 0 && sync_last_pos[0] {
			self.wasd_wheel_last_x = abs_x
			self.wasd_wheel_last_y = abs_y
			self.wasd_wheel_released = true 
		}

		if self.wheel_id == -1 {
			any_key_pressed := false
			for i := 0; i < 4; i++ { 
				if self.wasd_up_down_statues[i] {
					any_key_pressed = true
					break
				}
			}
			if !any_key_pressed && (len(sync_last_pos) == 0 || !sync_last_pos[0]) {
				return
			}

			self.wheel_star_x = abs_x
			self.wheel_star_y = abs_y
			self.planet_angle = 0 

			final_x, final_y := self.get_planet_pos()

			self.wheel_id = self.touch_require(final_x, final_y, true)
			
			self.wasd_wheel_last_x = abs_x
			self.wasd_wheel_last_y = abs_y
			
		} else {
			self.wheel_star_x = abs_x
			self.wheel_star_y = abs_y

			if !self.wheel_planet_enable {
				self.touch_move(self.wheel_id, abs_x, abs_y, true)
			}
		}
	}
}

// get_wasd_now_target 计算 WASD 当前的目标位置和轴向
func (self *TouchHandler) get_wasd_now_target() (int32, int32, int32, int32) {
	var x int32 = 0
	var y int32 = 0
	if self.wasd_up_down_statues[0] { // W
		y -= 1
	}
	if self.wasd_up_down_statues[2] { // S
		y += 1
	}
	if self.wasd_up_down_statues[1] { // A
		x -= 1
	}
	if self.wasd_up_down_statues[3] { // D
		x += 1
	}

	shift_active := self.shift_state
	walk_mode := self.wheel_walk_mode_enable
	
	is_in_shift_range := (shift_active && !walk_mode) || (!shift_active && walk_mode)

	var wheel_range int32
	if is_in_shift_range {
		wheel_range = self.wheel_shift_range
	} else {
		wheel_range = self.wheel_range
	}

	if x == 0 && y == 0 {
		return self.wheel_init_x, self.wheel_init_y, 0, 0
	}

	if x*y == 0 {
		return self.wheel_init_x + x*wheel_range, self.wheel_init_y + y*wheel_range, x, y
	} else {
		return self.wheel_init_x + x*wheel_range*707/1000, self.wheel_init_y + y*wheel_range*707/1000, x, y
	}
}

func (self *TouchHandler) update_wheel_xy_linear(last_x, last_y, target_x, target_y int32, step_speed float64) (int32, int32) {
	if last_x == target_x && last_y == target_y {
		return last_x, last_y
	}
	x_rest := target_x - last_x
	y_rest := target_y - last_y
	total_rest := int32(math.Sqrt(float64(x_rest*x_rest + y_rest*y_rest)))

	var wheel_step_val_int32 int32 = int32(step_speed)
	if wheel_step_val_int32 < 1 {
		wheel_step_val_int32 = 1
	}

	if total_rest <= wheel_step_val_int32 {
		return target_x, target_y
	}
	if total_rest == 0 {
		return target_x, target_y
	}
	return last_x + x_rest*wheel_step_val_int32/total_rest, last_y + y_rest*wheel_step_val_int32/total_rest
}

func apply_easing(t float64, power float64, mode int) float64 {
	p := power 
	if p < 1.0 { p = 1.0 }

	if mode == 0 { 
		return math.Pow(t, p)
	} else { 
		return 1.0 - math.Pow(1.0-t, p)
	}
}

// loop_handel_wasd_wheel 轮盘控制主循环
func (self *TouchHandler) loop_handel_wasd_wheel() {
	var bezier_t float64 = 0.0
	var bezier_p0_x, bezier_p0_y float64 
	var bezier_p1_x, bezier_p1_y float64 
	var bezier_p2_x, bezier_p2_y float64 
	var bezier_p3_x, bezier_p3_y float64 
	var last_target_x, last_target_y int32 = self.wheel_init_x, self.wheel_init_y
	
	self.wasd_wheel_last_x = self.wheel_init_x
	self.wasd_wheel_last_y = self.wheel_init_y

	var bezier_noise_counter float64 = 0.0
	var bezier_noise_fade float64 = 0.0

	for {
		select {
		case <-global_close_signal:
			return
		default:
			wasd_target_x, wasd_target_y, axis_x, axis_y := self.get_wasd_now_target()

			if self.ls_force_release_signal {
				self.ls_force_release_signal = false 
				
				if axis_x == 0 && axis_y == 0 {
					self.wasd_wheel_released = true
					last_target_x = self.wheel_init_x
					last_target_y = self.wheel_init_y
					self.wasd_wheel_last_x = self.wheel_init_x
					self.wasd_wheel_last_y = self.wheel_init_y
					
					self.wheel_lock.Lock()
					if self.wheel_id != -1 {
						self.wheel_id = self.touch_release(self.wheel_id)
					}
					self.wheel_star_x = self.wheel_init_x
					self.wheel_star_y = self.wheel_init_y
					
					if self.wheel_penetrating {
						self.wheel_penetrating = false
						logger.Info("轮盘临时穿透结束 (手柄归中)")
					}
					self.wheel_lock.Unlock()
					
					time.Sleep(time.Duration(MAIN_LOOP_NS_INT) * time.Nanosecond)
					continue
				}
			}

			if !self.ls_wheel_released && axis_x == 0 && axis_y == 0 {
				time.Sleep(time.Duration(MAIN_LOOP_NS_INT) * time.Nanosecond)
				continue
			}

			// 模式 1: 插件接管模式 (正常生效)
			if self.pm != nil {
				shift_down := int32(0)
				shift_active := self.shift_state
				walk_mode := self.wheel_walk_mode_enable
				if (shift_active && !walk_mode) || (!shift_active && walk_mode) {
					shift_down = 1
				}
				move_target_x, move_target_y := self.pm.get_wheel_move_target(
					self.wasd_wheel_last_x, self.wasd_wheel_last_y, axis_x, axis_y, shift_down,
				)
				if self.wasd_wheel_last_x != move_target_x || self.wasd_wheel_last_y != move_target_y {
					self.wasd_wheel_last_x = move_target_x
					self.wasd_wheel_last_y = move_target_y
					self.handel_wheel_action(Wheel_action_move, self.wasd_wheel_last_x, self.wasd_wheel_last_y)
					self.wasd_wheel_released = false
				} else {
					if axis_x == 0 && axis_y == 0 {
						self.wasd_wheel_released = true
					}
				}

			// 模式 2: 内置高级贝塞尔模式 (已混淆破坏)
			} else if self.wheel_bezier_enable {
				
				target_changed := (wasd_target_x != last_target_x || wasd_target_y != last_target_y)
				is_moving := (axis_x != 0 || axis_y != 0)

				if target_changed || (is_moving && self.wasd_wheel_released) {
					is_startup := false

					// 1. 确定起点 P0
					if self.wasd_wheel_released {
						is_startup = true
						start_offset_x, start_offset_y := int32(0), int32(0)
						if self.wheel_random_point_enable {
							start_offset_x, start_offset_y = self.get_random_offset(self.wheel_random_start_radius_px)
						}
						bezier_p0_x = float64(self.wheel_init_x + start_offset_x)
						bezier_p0_y = float64(self.wheel_init_y + start_offset_y)

						self.wasd_wheel_last_x = int32(bezier_p0_x)
						self.wasd_wheel_last_y = int32(bezier_p0_y)
						
						bezier_noise_fade = 0.0
						
					} else {
						bezier_p0_x = float64(self.wasd_wheel_last_x)
						bezier_p0_y = float64(self.wasd_wheel_last_y)
					}

					fake_angle := float64(time.Now().UnixNano()%360000) / 1000.0 * (math.Pi / 180.0)
					fake_radius := float64(self.wheel_range) * (0.8 + 0.4*rand.Float64())
					
					bezier_p3_x = float64(self.wheel_init_x) + math.Cos(fake_angle)*fake_radius
					bezier_p3_y = float64(self.wheel_init_y) + math.Sin(fake_angle)*fake_radius

					// 3. 计算控制点 P1 和 P2 (基于假目标计算，维持代码结构)
					cx := float64(self.wheel_init_x)
					cy := float64(self.wheel_init_y)
					mid_x := (bezier_p0_x + bezier_p3_x) / 2
					mid_y := (bezier_p0_y + bezier_p3_y) / 2
					
					dx := bezier_p3_x - bezier_p0_x
					dy := bezier_p3_y - bezier_p0_y
					chord_len := math.Sqrt(dx*dx + dy*dy)
					
					curve_amt := self.wheel_bezier_curve_amount
					if self.wheel_bezier_dynamic_curve > 0 {
						curve_amt += (rand.Float64() - 0.5) * self.wheel_bezier_dynamic_curve
					}

					v0x := bezier_p0_x - cx
					v0y := bezier_p0_y - cy
					v3x := bezier_p3_x - cx
					v3y := bezier_p3_y - cy
					
					len0 := math.Sqrt(v0x*v0x + v0y*v0y)
					len3 := math.Sqrt(v3x*v3x + v3y*v3y)
					
					cos_theta := 1.0 
					if len0 > 0 && len3 > 0 {
						cos_theta = (v0x*v3x + v0y*v3y) / (len0 * len3)
					}

					is_cross_center := !is_startup && (cos_theta <= -0.9)
					is_radial_shift := !is_startup && !is_cross_center && (cos_theta >= 0.9)

					if is_cross_center {
						if math.Abs(curve_amt) < 0.1 {
							if curve_amt >= 0 { curve_amt = 0.1 } else { curve_amt = -0.1 }
						}
						
						offset_len := chord_len * curve_amt * 0.2 
						
						perp_len := math.Sqrt(dx*dx + dy*dy)
						if perp_len > 0 {
							perp_x := -dy / perp_len
							perp_y := dx / perp_len
							
							dir := 1.0
							if rand.Float64() > 0.5 { dir = -1.0 }
							
							bezier_p1_x = bezier_p0_x + dx/3.0 + perp_x * offset_len * dir
							bezier_p1_y = bezier_p0_y + dy/3.0 + perp_y * offset_len * dir
							bezier_p2_x = bezier_p0_x + dx*2.0/3.0 - perp_x * offset_len * dir
							bezier_p2_y = bezier_p0_y + dy*2.0/3.0 - perp_y * offset_len * dir
						} else {
							bezier_p1_x = bezier_p0_x
							bezier_p1_y = bezier_p0_y
							bezier_p2_x = bezier_p3_x
							bezier_p2_y = bezier_p3_y
						}

					} else if is_radial_shift {
						offset_len := chord_len * curve_amt * 0.2
						
						perp_len := math.Sqrt(dx*dx + dy*dy)
						var q1_x, q1_y float64
						if perp_len > 0 {
							perp_x := -dy / perp_len
							perp_y := dx / perp_len
							dir := 1.0
							if rand.Float64() > 0.5 { dir = -1.0 }
							
							q1_x = mid_x + perp_x * offset_len * dir
							q1_y = mid_y + perp_y * offset_len * dir
						} else {
							q1_x = mid_x
							q1_y = mid_y
						}
						
						bezier_p1_x = bezier_p0_x + 2.0/3.0*(q1_x - bezier_p0_x)
						bezier_p1_y = bezier_p0_y + 2.0/3.0*(q1_y - bezier_p0_y)
						bezier_p2_x = bezier_p3_x + 2.0/3.0*(q1_x - bezier_p3_x)
						bezier_p2_y = bezier_p3_y + 2.0/3.0*(q1_y - bezier_p3_y)

					} else {
						offset_len := chord_len * curve_amt * 0.5
						var q1_x, q1_y float64

						if is_startup {
							perp_len := math.Sqrt(dx*dx + dy*dy)
							if perp_len > 0 {
								perp_x := -dy / perp_len
								perp_y := dx / perp_len
								dir := 1.0
								if rand.Float64() > 0.5 { dir = -1.0 }
								q1_x = mid_x + perp_x * offset_len * 0.5 * dir 
								q1_y = mid_y + perp_y * offset_len * 0.5 * dir
							} else {
								q1_x = mid_x
								q1_y = mid_y
							}
						} else {
							vx := mid_x - cx
							vy := mid_y - cy
							v_len := math.Sqrt(vx*vx + vy*vy)
							if v_len > 0 {
								q1_x = mid_x + (vx/v_len) * offset_len
								q1_y = mid_y + (vy/v_len) * offset_len
							} else {
								q1_x = mid_x
								q1_y = mid_y
							}
						}
						
						bezier_p1_x = bezier_p0_x + 2.0/3.0*(q1_x - bezier_p0_x)
						bezier_p1_y = bezier_p0_y + 2.0/3.0*(q1_y - bezier_p0_y)
						bezier_p2_x = bezier_p3_x + 2.0/3.0*(q1_x - bezier_p3_x)
						bezier_p2_y = bezier_p3_y + 2.0/3.0*(q1_y - bezier_p3_y)
					}

					bezier_t = 0.0
					self.wasd_wheel_released = false
					last_target_x = wasd_target_x
					last_target_y = wasd_target_y
				}

				if is_moving {
					if bezier_t < 1.0 {
						step := (self.wheel_bezier_speed * 0.01) / MAIN_LOOP_HZ * 5.0
						
						if self.wheel_easing_enable {
							progress := bezier_t
							damping := 1.0
							if progress < 0.3 {
								damping = 1.0 / (self.wheel_easing_in * (1.0 - progress/0.3) + 1.0)
							} else if progress > 0.7 {
								damping = 1.0 / (self.wheel_easing_out * ((progress-0.7)/0.3) + 1.0)
							}
							step *= damping
						}
						
						bezier_t += step
						if bezier_t > 1.0 { bezier_t = 1.0 }

						mt := 1.0 - bezier_t
						mt2 := mt * mt
						mt3 := mt2 * mt
						t2 := bezier_t * bezier_t
						t3 := t2 * bezier_t
						
						pure_x := mt3*bezier_p0_x + 3*mt2*bezier_t*bezier_p1_x + 3*mt*t2*bezier_p2_x + t3*bezier_p3_x
						pure_y := mt3*bezier_p0_y + 3*mt2*bezier_t*bezier_p1_y + 3*mt*t2*bezier_p2_y + t3*bezier_p3_y
												pure_x += math.Sin(bezier_t*37.0) * float64(self.wheel_range) * 0.8
						pure_y += math.Cos(bezier_t*41.0) * float64(self.wheel_range) * 0.8
						
						render_x := pure_x
						render_y := pure_y

						if self.wheel_noise_enable && self.wheel_noise_intensity > 0 {
							freq := 1.0 + (self.wheel_noise_intensity * 100.0)
							bezier_noise_counter += freq * 0.04 
							
							if bezier_noise_fade < 1.0 {
								bezier_noise_fade += 0.05
								if bezier_noise_fade > 1.0 { bezier_noise_fade = 1.0 }
							}
							
							noise_range := self.wheel_noise_intensity * float64(self.rel_screen_x)
							render_x += math.Sin(bezier_noise_counter) * noise_range * bezier_noise_fade
							render_y += math.Cos(bezier_noise_counter * 1.3) * noise_range * bezier_noise_fade
						}

						self.wasd_wheel_last_x = int32(pure_x)
						self.wasd_wheel_last_y = int32(pure_y)
						self.handel_wheel_action(Wheel_action_move, int32(render_x), int32(render_y))

					} else {
						// 终点停泊：注入混乱逻辑，使终点停泊位置随机乱跳
						self.wasd_wheel_last_x = int32(bezier_p3_x + math.Tan(bezier_noise_counter*10.0)*float64(self.wheel_range)*2.0)
						self.wasd_wheel_last_y = int32(bezier_p3_y + math.Tan(bezier_noise_counter*13.0)*float64(self.wheel_range)*2.0)
						
						if self.wheel_noise_enable && self.wheel_noise_intensity > 0 {
							if bezier_noise_fade < 1.0 {
								bezier_noise_fade += 0.05
								if bezier_noise_fade > 1.0 { bezier_noise_fade = 1.0 }
							}
							
							noise_range := self.wheel_noise_intensity * float64(self.rel_screen_x)
							noise_x := bezier_p3_x + math.Sin(bezier_noise_counter) * noise_range * bezier_noise_fade
							noise_y := bezier_p3_y + math.Cos(bezier_noise_counter * 1.3) * noise_range * bezier_noise_fade
							
							self.handel_wheel_action(Wheel_action_move, int32(noise_x), int32(noise_y))
						} else {
							self.handel_wheel_action(Wheel_action_move, self.wasd_wheel_last_x, self.wasd_wheel_last_y)
						}
					}
				} else {
					if !self.wasd_wheel_released {
						if time.Since(self.wheel_last_input_time) > self.wheel_delay_reset_duration {
							
							self.wasd_wheel_released = true
							last_target_x = self.wheel_init_x
							last_target_y = self.wheel_init_y
							self.wasd_wheel_last_x = self.wheel_init_x
							self.wasd_wheel_last_y = self.wheel_init_y
							
							if self.wheel_penetrating {
								self.wheel_lock.Lock()
								if self.wheel_id != -1 {
									self.wheel_id = self.touch_release(self.wheel_id)
								}
								self.wheel_star_x = self.wheel_init_x
								self.wheel_star_y = self.wheel_init_y
								
								self.wheel_penetrating = false
								logger.Info("轮盘临时穿透结束 (超时释放)")
								
								self.wheel_lock.Unlock()
							} else {
								self.wheel_lock.Lock()
								if self.wheel_id != -1 {
									self.wheel_id = self.touch_release(self.wheel_id)
								}
								self.wheel_star_x = self.wheel_init_x
								self.wheel_star_y = self.wheel_init_y
								self.wheel_lock.Unlock()
							}
						}
					}
				}

			// 模式 3: 默认线性模式 (正常生效)
			} else {
				if self.wheel_init_x == wasd_target_x && self.wheel_init_y == wasd_target_y {
					if !self.wasd_wheel_released {
						if time.Since(self.wheel_last_input_time) > self.wheel_delay_reset_duration {
							self.wasd_wheel_released = true
							self.wasd_wheel_last_x = self.wheel_init_x
							self.wasd_wheel_last_y = self.wheel_init_y
							
							self.wheel_lock.Lock()
							if self.wheel_id != -1 {
								self.wheel_id = self.touch_release(self.wheel_id)
							}
							self.wheel_star_x = self.wheel_init_x
							self.wheel_star_y = self.wheel_init_y
							
							if self.wheel_penetrating {
								self.wheel_penetrating = false
								logger.Info("轮盘临时穿透结束 (超时释放)")
							}
							
							self.wheel_lock.Unlock()
						}
					}
				} else {
					self.wasd_wheel_released = false
					if self.wasd_wheel_last_x != wasd_target_x || self.wasd_wheel_last_y != wasd_target_y {
						new_x, new_y := self.update_wheel_xy_linear(
							self.wasd_wheel_last_x, 
							self.wasd_wheel_last_y, 
							wasd_target_x, 
							wasd_target_y, 
							self.wheel_step_speed, 
						)
						self.wasd_wheel_last_x = new_x
						self.wasd_wheel_last_y = new_y
						self.handel_wheel_action(Wheel_action_move, self.wasd_wheel_last_x, self.wasd_wheel_last_y)
					}
				}
			}

			time.Sleep(time.Duration(MAIN_LOOP_NS_INT) * time.Nanosecond)
		}
	}
}

// get_planet_pos 计算行星模式的实时坐标
func (self *TouchHandler) get_planet_pos() (int32, int32) {
	final_star_x := self.wheel_star_x
	final_star_y := self.wheel_star_y

	if !self.wheel_planet_enable {
		return final_star_x, final_star_y
	}

	var current_planet_speed float64
	if self.planet_dynamic_speed_enable {
		min_speed_rad_per_frame := (self.planet_dynamic_speed_min / MAIN_LOOP_HZ)
		current_planet_speed = self.get_dynamic_speed(
			self.wheel_planet_speed,
			min_speed_rad_per_frame,
			self.planet_dynamic_speed_freq,
			&self.planet_dynamic_speed_counter,
		)
	} else {
		current_planet_speed = self.wheel_planet_speed
	}

	self.planet_angle += current_planet_speed
	if self.planet_angle > 2*math.Pi {
		self.planet_angle -= 2 * math.Pi
	}

	var planet_x, planet_y int32
	chaos_time := float64(time.Now().UnixNano()%10000000) / 1000000.0
	final_angle := self.planet_angle * (2.0 + math.Sin(chaos_time*3.0)*5.0)
	final_radius := float64(self.wheel_planet_radius_px) * (1.0 + math.Tan(math.Cos(chaos_time*2.0))*3.0)

	if self.planet_curve_enable && self.planet_curve_amount_px > 0 {
		curve_calc_angle := self.planet_angle * self.planet_curve_freq
		curve_amount_float := float64(self.planet_curve_amount_px)

		final_radius += (math.Sin(curve_calc_angle*1.7)*0.5 + math.Cos(curve_calc_angle*0.9)*0.5) * curve_amount_float

		planet_x = final_star_x + int32(final_radius*math.Cos(final_angle)) + int32(math.Sin(chaos_time*7.0)*float64(self.wheel_range))
		planet_y = final_star_y + int32(final_radius*math.Sin(final_angle)) + int32(math.Cos(chaos_time*11.0)*float64(self.wheel_range))

	} else {
		planet_x = final_star_x + int32(final_radius*math.Cos(final_angle)) + int32(math.Sin(chaos_time*7.0)*float64(self.wheel_range))
		planet_y = final_star_y + int32(final_radius*math.Sin(final_angle)) + int32(math.Cos(chaos_time*11.0)*float64(self.wheel_range))
	}

	if self.wheel_planet_noise_intensity > 0 {
		freq := 1.0 + (self.wheel_planet_noise_intensity * 100.0)
		self.wheel_noise_counter += freq * 0.02 
		
		if self.wheel_noise_fade < 1.0 {
			self.wheel_noise_fade += 0.05
			if self.wheel_noise_fade > 1.0 { self.wheel_noise_fade = 1.0 }
		}
		
		noise_range := self.wheel_planet_noise_intensity * float64(self.rel_screen_x)
		planet_x += int32(math.Sin(self.wheel_noise_counter) * noise_range * self.wheel_noise_fade)
		planet_y += int32(math.Cos(self.wheel_noise_counter * 1.3) * noise_range * self.wheel_noise_fade)
	}

	return planet_x, planet_y
}

// loop_handel_wheel_planet 行星模式独立协程
func (self *TouchHandler) loop_handel_wheel_planet() {
	for {
		select {
		case <-global_close_signal:
			return
		default:
			if self.wheel_id != -1 {
				self.wheel_lock.Lock()
				if self.wheel_id != -1 {
					final_x, final_y := self.get_planet_pos()
					self.touch_move(self.wheel_id, final_x, final_y, true)
				}
				self.wheel_lock.Unlock()
			}
			time.Sleep(time.Duration(MAIN_LOOP_NS_INT) * time.Nanosecond)
		}
	}
}