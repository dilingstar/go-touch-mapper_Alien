package main

import (
	"github.com/kenshaw/evdev"
)

// Version: V3.4.5
// 陀螺仪与运动传感器模块 (作为原版实验性功能的独立存放点)

// 全局的传感器范围参数 { dev_name: { x:[-100,100] , y:[-200,200]  } }
var global_motion_sensors_range = make(map[string]map[uint16][]int32)

// handel_gyro_events 用于处理被识别为 type_motion_sensors 的设备事件
// 这是从原版 v5.0.5 提取的实验性草稿代码，直接封存，后续可随时接入具体映射逻辑
func (self *TouchHandler) handel_gyro_events(events []*evdev.Event, dev_name string) {
	for _, event := range events {
		// 原版草稿注释：
		// 手持手柄水平放置桌面上，USB-c口水平向前
		// 向左为x正
		// 向下为y正
		// 向前为z正

		//陀螺仪 重力加速度在每个方向上的分量，可判断手柄在三维空间的姿态
		// AbsoluteX 重力在x轴的分量
		// AbsoluteY 重力在y轴的分量
		// AbsoluteZ 重力在z轴的分量

		//角速度 左手定则：握住对应的轴，拇指指向轴正方向，手柄沿握拳的方向旋转为正
		// AbsoluteRX 俯仰角 向左为x正
		// AbsoluteRY 偏航角 向下为y正
		// AbsoluteRZ 滚转角 向前为z正

		// 调试打印草稿示例：
		// testType := uint16(evdev.AbsoluteX)
		// if event.Code == testType {
		//     harfLen := 30
		//     fmt.Print(strings.Repeat("=", harfLen+int(float64(harfLen*int(event.Value))/32768)) + "|" + strings.Repeat("=", harfLen-int(float64(harfLen*int(event.Value))/32768)) + fmt.Sprintf("(%d,%d)", global_motion_sensors_range[dev_name][event.Code][0], global_motion_sensors_range[dev_name][event.Code][1]) + "\r")
		// }

		_ = event // 忽略未使用报错
	}
}