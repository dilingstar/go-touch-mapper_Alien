// v0.3.2
package main

import (
	"encoding/binary"
	"net"
	"sync"
	"time"
)

// mode: "tcp", "udp"
// addr: "127.0.0.1:61071"
func handel_touch_using_app_manager(mode string, addr string) touch_control_func {
	var conn net.Conn
	var mu sync.Mutex
	var isConnecting bool

	var connect func()
	connect = func() {
		mu.Lock()
		if isConnecting {
			mu.Unlock()
			return
		}
		isConnecting = true
		mu.Unlock()

		for {
			select {
			case <-global_close_signal:
				return
			default:
				var c net.Conn
				var err error

				switch mode {
				case "tcp":
					c, err = net.Dial("tcp4", addr)
				case "udp":
					c, err = net.Dial("udp4", addr)
				default:
					logger.Errorf("未知 App 桥接模式: %s", mode)
					return
				}

				if err == nil {
					mu.Lock()
					conn = c
					isConnecting = false
					mu.Unlock()
					logger.Infof(">>> 已成功连接到 App 桥接接口 [%s]: %s <<<", mode, addr)
					
					// UDP 无状态不需要探活，TCP 需要探活
					if mode == "tcp" {
						go func(currentConn net.Conn) {
							buf := make([]byte, 1)
							_, err := currentConn.Read(buf)
							if err != nil {
								logger.Warnf("与 App [%s] 的连接已断开，进入后台静默重连...", mode)
								mu.Lock()
								if conn == currentConn {
									conn.Close()
									conn = nil
									go connect()
								}
								mu.Unlock()
							}
						}(c)
					}
					return
				}
				time.Sleep(2 * time.Second)
			}
		}
	}

	go connect()

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
		mu.Lock()
		currentConn := conn
		mu.Unlock()

		if currentConn == nil {
			return
		}

		switch control_data.action {
		case TouchActionRequire, TouchActionMove:
			x, y := rotateAbsoluteXY(control_data.x, control_data.y)
			setReport(0x01, uint8(control_data.id), uint32(x), uint32(y))
		case TouchActionRelease:
			setReport(0x00, uint8(control_data.id), 0, 0)
		case TouchActionResetResolution:
			setReport(0x03, uint8(control_data.id), uint32(control_data.x), uint32(control_data.y))
		default:
			return
		}

		_, err := currentConn.Write(buf[:])
		if err != nil && mode == "tcp" {
			mu.Lock()
			if conn == currentConn {
				conn.Close()
				conn = nil
				go connect()
			}
			mu.Unlock()
		}
	}
}