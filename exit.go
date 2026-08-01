package main

import (
	"os"
	"syscall"
)

// Version: V3.5.0
// TriggerSafeExit 触发程序的平滑退出机制
// 给自身进程发送 SIGTERM 信号，复用 main.go 尾部监听系统的优雅退出流程，确保所有触点及资源被安全释放。
func TriggerSafeExit() {
	logger.Warnf("接收到结束程序指令，正在清理状态与释放设备...")
	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGTERM)
}