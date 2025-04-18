package main

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/hellobchain/nginxgo/common/constant"
	"github.com/hellobchain/nginxgo/nginxgo"
	"github.com/hellobchain/nginxgo/pkg/utils"
	"github.com/hellobchain/wswlog/wlogging"
	"github.com/spf13/pflag"
)

var logger = wlogging.MustGetLoggerWithoutName()
var engine *nginxgo.Engine // 声明一个 Engine 变量
var mu sync.Mutex          // 互斥锁，用于保护 engine 变量的访问
func main() {
	// 监听SIGHUP信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	command = strings.TrimSpace(command)
	operateCMD(command, 1)
	utils.HandleSignals(logger)
	pidStr := strconv.Itoa(os.Getpid())
	os.WriteFile("nginxgo.pid", []byte(pidStr), 0644)
	for {
		select {
		case <-signalChan:
			// 收到 SIGHUP 信号，执行重置操作
			logger.Info("nginxgo: received SIGHUP signal, resetting engine")
			operateCMD(constant.CMD_RESET, -1)
		}
	}
}

var command string
var flags *pflag.FlagSet

const flagCommand = "commond"

func init() {
	flags = &pflag.FlagSet{}
	flags.StringVar(&command, flagCommand, "start", "操作命令,--command=start 启动服务,--command=stop 停止服务,--command=reset 重置服务,--command=help 查看帮助")
	// 定义一个简写的参数
	pflag.StringVarP(&command, "c", "c", "start", "操作命令,--command=start 启动服务,--command=stop 停止服务,--command=reset 重置服务,--command=help 查看帮助")
	pflag.Parse()
}

func operateCMD(command string, pid int) {
	switch command {
	case constant.CMD_START:
		mu.Lock()
		if engine == nil {
			engine = nginxgo.Init()
			go engine.Start()
		} else {
			logger.Warn("nginxgo: engine already started")
		}
		mu.Unlock()
	case constant.CMD_RESET:
		if pid > 0 {
			pid, err := readPidFile()
			if err != nil {
				logger.Error("nginxgo: read pid file error:", err)
			} else {
				sendSignalHup(pid)
			}
			os.Exit(1)
		} else {
			mu.Lock()
			if engine != nil {
				engine.Reset()
			} else {
				logger.Warn("nginxgo: engine not started yet")
			}
			mu.Unlock()
		}
	case constant.CMD_STOP:
		if pid > 0 {
			pid, err := readPidFile()
			if err != nil {
				logger.Error("nginxgo: read pid file error:", err)
			} else {
				sendSignalTerm(pid)
			}
			os.Exit(1)
		} else {
			mu.Lock()
			if engine != nil {
				engine.Stop()
				engine = nil // 释放引用，方便重新启动
			} else {
				logger.Warn("nginxgo: engine not started yet")
			}
			mu.Unlock()
		}
	case constant.CMD_HELP:
		logger.Info("nginxgo -c=help")
		logger.Info("nginxgo -c=start")
		logger.Info("nginxgo -c=stop")
		logger.Info("nginxgo -c=reset")
	default:
		logger.Error("nginxgo: error command")
	}
}

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
	data, err := os.ReadFile("nginxgo.pid")
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
