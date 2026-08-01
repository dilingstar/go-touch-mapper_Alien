package main

import (
	"time"
)

// Version: V3.4.5
// loop_handel_recoil 全局压枪处理循环
// 负责按照设定的速度自动向下移动视角
// 频率: 5ms (200Hz)
func (self *TouchHandler) loop_handel_recoil() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-global_close_signal:
			return
		case <-ticker.C:
			// [V3.4.5] 增加拦截器：开启映射 + 激活压枪 + 速度大于0 + 【指针未切出】
			if self.map_on && self.recoil_active && self.current_recoil_speed > 0 && !self.pointer_is_out_temp {
				// 注入向下的位移 (Y轴正向)
				// 速度系数校准：假设 current_recoil_speed 100 代表非常快 (100 px/unit)
				// 这里直接将 speed 作为基础像素位移传递给 handel_view_move
				// 5ms 执行一次
				// 缩放系数 0.2 -> Speed 10 = 2px/5ms = 400px/s
				
				move_y := int32(self.current_recoil_speed * 0.2)
				if move_y > 0 {
					// 压枪只涉及 Y 轴移动，X 轴为 0
					self.handel_view_move(0, move_y)
				}
			}
		}
	}
}