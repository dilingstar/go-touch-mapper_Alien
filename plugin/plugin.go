//go:build ignore || plugin
// +build ignore plugin

package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

func uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

func readUint32(reader *bufio.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := reader.Read(buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

func writeString(writer *bufio.Writer, s string) error {
	lenBytes := uint32ToBytes(uint32(len(s)))
	writer.Write(lenBytes)
	writer.WriteString(s)
	return writer.Flush()
}

func readString(reader *bufio.Reader) (string, error) {
	len, err := readUint32(reader)
	if err != nil {
		return "", err
	}
	buf := make([]byte, len)
	if _, err := reader.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeInt32List(writer *bufio.Writer, list []int32) error {
	buffer := make([]byte, 4*len(list)+4)
	binary.LittleEndian.PutUint32(buffer[0:], uint32(len(list)))
	for i, v := range list {
		binary.LittleEndian.PutUint32(buffer[4+i*4:], uint32(v))
	}
	writer.Write(buffer)
	return writer.Flush()
}

func readInt32List(reader *bufio.Reader) ([]int32, error) {
	len, err := readUint32(reader)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4*len)
	if _, err := reader.Read(buf); err != nil {
		return nil, err
	}
	list := make([]int32, len)
	for i := range list {
		list[i] = int32(binary.LittleEndian.Uint32(buf[i*4:]))
	}
	return list, nil
}

func main() {
	unixAddr, err := net.ResolveUnixAddr("unix", "@uds_go_touch_mapper_plugin")
	if err != nil {
		fmt.Println("ResolveUnixAddr failed : ", err.Error())
		return
	}
	conn, err := net.DialUnix("unix", nil, unixAddr)
	if err != nil {
		fmt.Println("DialUnix failed : ", err.Error())
		return
	}
	fmt.Println("initPlugin success, conn: ", conn)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writeString(writer, PluginID)
	writeString(writer, PluginConfigTemplate)
	userConfigString, err := readString(reader)
	if err != nil {
		fmt.Println("读取用户配置失败:", err.Error())
		return
	}
	userConfig := make(map[string]interface{})
	if err := json.Unmarshal([]byte(userConfigString), &userConfig); err != nil {
		fmt.Println("解析用户配置失败:", err.Error())
		return
	} else {
		fmt.Println("用户配置:", userConfig)
	}

	customWheel := initWheel()
	customWheel.update_user_config(userConfig)
	for {
		cmd, _ := reader.ReadByte()
		switch cmd {
		case 0xf1:
			// fmt.Println("收到 update_user_config")
			userConfigString, err := readString(reader)
			if err != nil {
				fmt.Println("读取用户配置失败:", err.Error())
				return
			}
			userConfig := make(map[string]interface{})
			if err := json.Unmarshal([]byte(userConfigString), &userConfig); err != nil {
				fmt.Println("解析用户配置失败:", err.Error())
				return
			} else {
				fmt.Println("用户配置:", userConfig)
				customWheel.update_user_config(userConfig)
			}
		case 0xf2:
			// fmt.Println("收到 update_wheel_config")
			wheelConfigList, _ := readInt32List(reader)
			customWheel.update_wheel_config(wheelConfigList[0], wheelConfigList[1], wheelConfigList[2], wheelConfigList[3], wheelConfigList[4], wheelConfigList[5])
		case 0xf3:
			// fmt.Println("收到 get_wheel_move_target")
			payload, _ := readInt32List(reader)
			x, y := customWheel.get_wheel_move_target(payload[0], payload[1], payload[2], payload[3], payload[4] == 1)
			returnBytes := make([]byte, 9)
			// returnBytes[0] == uint8(0xf3)
			binary.LittleEndian.PutUint32(returnBytes[1:5], uint32(x))
			binary.LittleEndian.PutUint32(returnBytes[5:9], uint32(y))
			writer.Write(returnBytes)
			writer.Flush()
		}
	}

}
