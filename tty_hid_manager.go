package main

import (
	"encoding/binary"

	"go.bug.st/serial"
)

func handel_touch_using_hid_manager(port serial.Port) touch_control_func {
	var buf [16]byte
	buf[0] = 0x55
	buf[1] = 0xaa
	buf[2] = 0x0c
	buf[3] = 0x03
	setReport := func(action uint8, id uint8, x, y uint32) {
		buf[4] = action
		buf[5] = id
		binary.LittleEndian.PutUint32(buf[6:10], x)
		binary.LittleEndian.PutUint32(buf[10:14], y)
		buf[14] = 0
		buf[15] = buf[2] ^ buf[3] ^ buf[4] ^ buf[5] ^ buf[6] ^ buf[7] ^ buf[8] ^ buf[9] ^ buf[10] ^ buf[11] ^ buf[12] ^ buf[13] ^ buf[14]
	}

	return func(control_data touch_control_pack) {
		switch control_data.action {
		case TouchActionRequire, TouchActionMove:
			x, y := rotateAbsoluteXY(control_data.x, control_data.y)
			setReport(0x01, uint8(control_data.id), uint32(x), uint32(y))
			port.Write(buf[:])
		case TouchActionRelease:
			setReport(0x00, uint8(control_data.id), 0, 0)
			port.Write(buf[:])
		case TouchActionResetResolution:
			setReport(0x03, uint8(control_data.id), uint32(control_data.x), uint32(control_data.y))
			port.Write(buf[:])
		}
	}
}

// func handel_touch_using_hid_manager(port serial.Port) touch_control_func {
// 	var buf [16]byte
// 	buf[0] = 0x55
// 	buf[1] = 0xaa
// 	buf[2] = 0x0c
// 	buf[3] = 0x03
// 	setReport := func(action uint8, id uint8, x, y uint32) {
// 		buf[4] = action
// 		buf[5] = id
// 		binary.LittleEndian.PutUint32(buf[6:10], x)
// 		binary.LittleEndian.PutUint32(buf[10:14], y)
// 		buf[14] = 0
// 		buf[15] = buf[2] ^ buf[3] ^ buf[4] ^ buf[5] ^ buf[6] ^ buf[7] ^ buf[8] ^ buf[9] ^ buf[10] ^ buf[11] ^ buf[12] ^ buf[13] ^ buf[14]
// 		logger.Errorf("buf: %x", buf)
// 	}

// 	// 创建UDP连接
// 	udpAddr, err := net.ResolveUDPAddr("udp", "192.168.3.255:61068")
// 	if err != nil {
// 		logger.Errorf("Error resolving UDP address: %v", err)
// 	}
// 	conn, err := net.DialUDP("udp", nil, udpAddr)
// 	if err != nil {
// 		logger.Errorf("Error creating UDP connection: %v", err)
// 	}

// 	return func(control_data touch_control_pack) {
// 		switch control_data.action {
// 		case TouchActionRequire, TouchActionMove:
// 			x, y := rotateAbsoluteXY(control_data.x, control_data.y)
// 			setReport(0x01, uint8(control_data.id), uint32(x), uint32(y))
// 			_, err := conn.Write(buf[:])
// 			if err != nil {
// 				logger.Errorf("Error sending UDP data: %v", err)
// 			}
// 		case TouchActionRelease:
// 			setReport(0x00, uint8(control_data.id), 0, 0)
// 			_, err := conn.Write(buf[:])
// 			if err != nil {
// 				logger.Errorf("Error sending UDP data: %v", err)
// 			}
// 		case TouchActionResetResolution:
// 			setReport(0x03, uint8(control_data.id), uint32(control_data.x), uint32(control_data.y))
// 			_, err := conn.Write(buf[:])
// 			if err != nil {
// 				logger.Errorf("Error sending UDP data: %v", err)
// 			}
// 		}
// 	}
// }
