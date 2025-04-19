package nginxgo

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

func sendSignalHup(targetPID int) {
	// 获取目标进程
	process, err := os.FindProcess(targetPID)
	if err != nil {
		logger.Error("Failed to find process:", err)
		return
	}
	// 向目标进程发送 SIGHUP 信号
	err = process.Signal(syscall.SIGHUP)
	if err != nil {
		logger.Error("Failed to send SIGHUP signal:", err)
		return
	}
}

func sendSignalTerm(targetPID int) {
	// 获取目标进程
	process, err := os.FindProcess(targetPID)
	if err != nil {
		logger.Error("Failed to find process:", err)
		return
	}
	// 向目标进程发送 SIGTERM 信号
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		logger.Error("Failed to send SIGTERM signal:", err)
		return
	}
}

func readPidFile() (int, error) {
	data, err := os.ReadFile(pidFilePath)
	if err != nil {
		return -1, err
	}
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return -1, err
	}
	return pid, nil
}
