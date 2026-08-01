package main

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

// Version: V3.4.6
type v_mouse_controller struct {
	touchHandlerInstance *TouchHandler
	uinput_in            chan *u_input_control_pack
	uinput_out           chan *u_input_control_pack
	working              bool
	left_downing         bool
	left_btn_press_count int32 // [V3.4.6] 新增左键引用计数，解决多设备同时点击的粘连悖论
	mouse_x              int32
	mouse_y              int32
	udp_write_ch         chan []byte
	mouse_id             int32
	screen_x             int32
	screen_y             int32
	map_switch_signal    chan bool
	wheel_move_chan      chan int32
	scroll_lock          sync.Mutex
}

func init_v_mouse_controller(
	touchHandlerInstance *TouchHandler,
	u_input_control_ch chan *u_input_control_pack,
	fileted_u_input_control_ch chan *u_input_control_pack,
	map_switch_signal chan bool,
	addr net.UDPAddr,
) *v_mouse_controller {
	udp_write_ch := make(chan []byte)
	go (func() {
		localAddr, err := net.ResolveUDPAddr("udp", ":0") // 使用随机本地端口
		if err != nil {
			logger.Errorf("解析本地地址失败: %s", err.Error())
			os.Exit(3)
		}
		socket, err := net.ListenUDP("udp", localAddr)
		if err != nil {
			logger.Errorf("udp error : %v", err)
			return
		}
		if err != nil {
			logger.Errorf("连接v_mouse失败 : %s", err.Error())
			os.Exit(3)
		}
		defer socket.Close()
		localAddrReal := socket.LocalAddr().(*net.UDPAddr)
		logger.Infof("v_mouse服务启动在端口: %d", localAddrReal.Port)

		udpRecvCh := make(chan []byte, 100)
		go func() {
			readBuffer := make([]byte, 32)
			for {
				select {
				case <-global_close_signal:
					close(udpRecvCh)
					return
				default:
					n, remoteAddr, err := socket.ReadFromUDP(readBuffer)
					if err != nil {
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							continue
						}
						logger.Errorf("读取UDP数据失败: %s", err.Error())
						continue
					}
					if n > 0 {
						data := make([]byte, n)
						copy(data, readBuffer[:n])
						global_device_orientation = int32(data[0])
						logger.Debugf("从 %s 接收到 vpoint上报屏幕方向 %d ", remoteAddr.String(), int32(data[0]))
					}
				}
			}
		}()

		for {
			select {
			case <-global_close_signal:
				return
			default:
				data := <-udp_write_ch
				socket.WriteToUDP(data, &addr)
			}
		}
	})()

	screen_x, screen_y := int32(0), int32(0)
	if global_is_wordking_remote {
		screen_x = global_screen_x
		screen_y = global_screen_y
	} else {
		screen_x, screen_y = get_wm_size()
	}
	wheel_move_chan := make(chan int32, 100)

	return &v_mouse_controller{
		touchHandlerInstance: touchHandlerInstance,
		uinput_in:            u_input_control_ch,
		uinput_out:           fileted_u_input_control_ch,
		working:              true,
		left_downing:         false,
		left_btn_press_count: 0,
		mouse_x:              0,
		mouse_y:              0,
		udp_write_ch:         udp_write_ch,
		mouse_id:             -1,
		screen_x:             screen_x,
		screen_y:             screen_y,
		map_switch_signal:    map_switch_signal,
		wheel_move_chan:      wheel_move_chan,
		scroll_lock:          sync.Mutex{},
	}
}

func (self *v_mouse_controller) get_max_xy_val() (int32, int32) {
	if global_is_wordking_remote {
		return global_screen_x, global_screen_y
	} else {
		if global_device_orientation == 0 || global_device_orientation == 2 {
			return self.screen_x, self.screen_y
		} else {
			return self.screen_y, self.screen_x
		}
	}
}

func (self *v_mouse_controller) get_config() (resetPos []float64, mouseSpeed []float64) {
	cfg := self.touchHandlerInstance.config.Get("V_MOUSE_SETTINGS")

	resetPos = []float64{0.5, 0.5}
	mouseSpeed = []float64{1.0, 1.0}

	if rpArray, err := cfg.Get("RESET_POS").Array(); err == nil && len(rpArray) >= 2 {
		rX := cfg.Get("RESET_POS").GetIndex(0).MustFloat64(0.5)
		rY := cfg.Get("RESET_POS").GetIndex(1).MustFloat64(0.5)
		resetPos = []float64{rX, rY}
	}

	if msArray, err := cfg.Get("MOUSE_SPEED").Array(); err == nil && len(msArray) >= 2 {
		sX := cfg.Get("MOUSE_SPEED").GetIndex(0).MustFloat64(1.0)
		sY := cfg.Get("MOUSE_SPEED").GetIndex(1).MustFloat64(1.0)
		mouseSpeed = []float64{sX, sY}
	}
	return
}

func (self *v_mouse_controller) main_loop() {
	for {
		select {
		case <-global_close_signal:
			return
		case map_on := <-self.map_switch_signal:
			self.working = !map_on
			if self.working {
				resetPos, _ := self.get_config()
				max_x, max_y := self.get_max_xy_val()

				self.mouse_x = int32(resetPos[0] * float64(max_x))
				self.mouse_y = int32(resetPos[1] * float64(max_y))

				self.display_mouse_control(true, self.left_downing, self.mouse_x, self.mouse_y)
			} else {
				self.display_mouse_control(false, false, 0, 0)
				// [V3.4.6] 彻底清零状态
				if self.left_downing || self.left_btn_press_count > 0 {
					self.left_downing = false
					self.left_btn_press_count = 0
					if self.mouse_id != -1 {
						self.touchHandlerInstance.touch_release(self.mouse_id)
						self.mouse_id = -1
					}
				}
			}
		case data := <-self.uinput_in:
			if data.action == UInput_mouse_move {
				self.on_mouse_move(data.arg1, data.arg2)
			} else if data.action == UInput_mouse_wheel {
				if data.arg1 == REL_WHEEL {
					go self.on_hwheel_action(data.arg2)
				}
			} else {
				if data.arg1 >= 0x110 && data.arg1 <= 0x117 {
					if data.arg1 == 0x110 {
						self.on_left_btn(data.arg2)
					}
				} else {
					self.uinput_out <- data
				}
			}
		}
	}
}

func (self *v_mouse_controller) display_mouse_control(show, downing bool, abs_x, abs_y int32) {
	var show_int int32
	if show {
		show_int = 1
	} else {
		show_int = 0
	}
	var downing_int int32
	if downing {
		downing_int = 1
	} else {
		downing_int = 0
	}
	fmt_str := fmt.Sprintf("%d,%d,%d,%d,%d", abs_x, abs_y, show_int, downing_int, global_device_orientation)
	self.udp_write_ch <- []byte(fmt_str)
}

func (self *v_mouse_controller) on_mouse_move(rel_x, rel_y int32) {
	if self.working {
		_, mouseSpeed := self.get_config()

		adj_x := int32(float64(rel_x) * mouseSpeed[0])
		adj_y := int32(float64(rel_y) * mouseSpeed[1])

		self.mouse_x += adj_x
		self.mouse_y += adj_y

		max_x, max_y := self.get_max_xy_val()
		if self.mouse_x < 0 {
			self.mouse_x = 0
		}
		if self.mouse_y < 0 {
			self.mouse_y = 0
		}
		if self.mouse_x > max_x {
			self.mouse_x = max_x
		}
		if self.mouse_y > max_y {
			self.mouse_y = max_y
		}
		self.display_mouse_control(true, self.left_downing, self.mouse_x, self.mouse_y)
		if self.left_downing && self.mouse_id != -1 {
			self.touchHandlerInstance.touch_move(self.mouse_id, int32(int64(self.mouse_x)*0x7ffffffe/int64(max_x)), int32(int64(self.mouse_y)*0x7ffffffe/int64(max_y)), false)
		}
	}
}

func (self *v_mouse_controller) on_left_btn(up_down int32) {
	if self.working {
		if up_down == DOWN {
			self.left_btn_press_count++
			// [V3.4.6] 只有当第一个源按下时，才真正触发屏幕点击
			if self.left_btn_press_count == 1 {
				self.left_downing = true
				max_x, max_y := self.get_max_xy_val()
				self.mouse_id = self.touchHandlerInstance.touch_require(int32(int64(self.mouse_x)*0x7ffffffe/int64(max_x)), int32(int64(self.mouse_y)*0x7ffffffe/int64(max_y)), false)
			}
		} else {
			self.left_btn_press_count--
			if self.left_btn_press_count <= 0 {
				self.left_btn_press_count = 0
				self.left_downing = false
				if self.mouse_id != -1 {
					self.touchHandlerInstance.touch_release(self.mouse_id)
					self.mouse_id = -1
				}
			}
		}
		self.on_mouse_move(0, 0)
	}
}

func (self *v_mouse_controller) on_hwheel_action(value int32) {
	if self.working {
		self.wheel_move_chan <- value
	}
}

func (self *v_mouse_controller) loop_handel_v_mouse_wheel_move() {
	var wheel_touch_id int32 = -1
	var current_pos_y int32 = 0
	var last_scroll_time time.Time
	var last_reset_time time.Time
	var curve_counter float64 = 0

	release_ticker := time.NewTicker(20 * time.Millisecond)
	defer release_ticker.Stop()

	for {
		select {
		case <-global_close_signal:
			return

		case raw_val := <-self.wheel_move_chan:
			self.scroll_lock.Lock()

			cfg := self.touchHandlerInstance.config.Get("V_MOUSE_SETTINGS").Get("SCROLL_CONFIG")
			invert := self.touchHandlerInstance.config.Get("V_MOUSE_SETTINGS").Get("ENABLE_INVERT_SCROLL").MustBool(true)

			speed_mult := cfg.Get("SPEED").MustFloat64(1.0)
			non_reset_ms := time.Duration(cfg.Get("NON_RESET_MS").MustInt(300)) * time.Millisecond
			reset_delay_ms := time.Duration(cfg.Get("RESET_DELAY_MS").MustInt(50)) * time.Millisecond

			last_scroll_time = time.Now()

			if time.Since(last_reset_time) > non_reset_ms {
				current_pos_y = 0
				curve_counter = 0
			}
			last_reset_time = time.Now()

			var delta int32
			base_step := 40.0 * speed_mult
			if invert {
				delta = int32(float64(raw_val) * base_step)
			} else {
				delta = int32(float64(raw_val) * -base_step)
			}

			current_pos_y += delta

			max_x, max_y := self.get_max_xy_val()
			center_x_px := self.mouse_x
			center_y_px := self.mouse_y

			var offset_x_px int32 = 0
			curve_enable := cfg.Get("CURVE_ENABLE").MustBool(false)
			if curve_enable {
				amount := cfg.Get("CURVE_AMOUNT").MustFloat64(0.005)
				freq := cfg.Get("CURVE_FREQ").MustFloat64(1.0)
				curve_counter += 0.2 * freq
				offset_pct := math.Sin(curve_counter) * amount
				offset_x_px = int32(offset_pct * float64(max_x))
			}
			noise_enable := cfg.Get("DYNAMIC_NOISE_ENABLE").MustBool(false)
			if noise_enable {
				amount := cfg.Get("DYNAMIC_NOISE_AMOUNT").MustFloat64(0.002)
				noise_px := (rand.Float64() - 0.5) * 2 * amount * float64(max_x)
				offset_x_px += int32(noise_px)
			}

			target_x_px := center_x_px + offset_x_px
			target_y_px := center_y_px + current_pos_y

			boundary_hit := false
			if target_y_px < 0 || target_y_px > max_y {
				boundary_hit = true
				current_pos_y = 0
				target_y_px = center_y_px
			}

			final_x := int32(int64(target_x_px) * 0x7ffffffe / int64(max_x))
			final_y := int32(int64(target_y_px) * 0x7ffffffe / int64(max_y))

			is_new_press := (wheel_touch_id == -1)

			if is_new_press || boundary_hit {
				if boundary_hit && wheel_touch_id != -1 {
					self.touchHandlerInstance.touch_release(wheel_touch_id)
				}

				wheel_touch_id = -2

				go func(req_x, req_y int32, is_boundary bool, delay_val time.Duration) {
					var wait time.Duration = 0
					if is_boundary {
						wait = delay_val
					}

					if wait > 0 {
						time.Sleep(wait)
					}

					self.scroll_lock.Lock()
					defer self.scroll_lock.Unlock()

					if wheel_touch_id == -2 {
						if is_boundary {
							wheel_touch_id = -1
						} else {
							wheel_touch_id = self.touchHandlerInstance.touch_require(req_x, req_y, false)
						}
					}
				}(final_x, final_y, boundary_hit, reset_delay_ms)

			} else {
				if wheel_touch_id >= 0 {
					self.touchHandlerInstance.touch_move(wheel_touch_id, final_x, final_y, false)
				}
			}

			self.scroll_lock.Unlock()

		case <-release_ticker.C:
			self.scroll_lock.Lock()
			if wheel_touch_id >= 0 {
				cfg := self.touchHandlerInstance.config.Get("V_MOUSE_SETTINGS").Get("SCROLL_CONFIG")
				release_delay := time.Duration(cfg.Get("RELEASE_DELAY_MS").MustInt(50)) * time.Millisecond

				if time.Since(last_scroll_time) > release_delay {
					self.touchHandlerInstance.touch_release(wheel_touch_id)
					wheel_touch_id = -1
				}
			}
			self.scroll_lock.Unlock()
		}
	}
}