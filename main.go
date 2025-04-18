package main

import (
	"strings"
	"sync"

	"github.com/hellobchain/nginxgo/common/constant"
	"github.com/hellobchain/nginxgo/nginxgo"
	"github.com/hellobchain/nginxgo/pkg/utils"
	"github.com/hellobchain/wswlog/wlogging"
	"github.com/spf13/pflag"
)

var logger = wlogging.MustGetLoggerWithoutName()

func main() {
	var engine *nginxgo.Engine // 声明一个 Engine 变量
	var mu sync.Mutex          // 互斥锁，用于保护 engine 变量的访问
	command = strings.TrimSpace(command)
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
		mu.Lock()
		if engine != nil {
			engine.Reset()
		} else {
			logger.Warn("nginxgo: engine not started yet")
		}
		mu.Unlock()
	case constant.CMD_STOP:
		mu.Lock()
		if engine != nil {
			engine.Stop()
			engine = nil // 释放引用，方便重新启动
		} else {
			logger.Warn("nginxgo: engine not started yet")
		}
		mu.Unlock()
	case constant.CMD_HELP:
		logger.Info("nginxgo: help")
		logger.Info("nginxgo: start")
		logger.Info("nginxgo: stop")
		logger.Info("nginxgo: reset")
		logger.Info("nginxgo: help")
	default:
		logger.Error("nginxgo: error command")
	}
	utils.HandleSignals(logger)
	for {
	}
}

var command string
var flags *pflag.FlagSet

const flagCommand = "commond"

func init() {
	flags = &pflag.FlagSet{}
	flags.StringVar(&command, flagCommand, "start",
		"操作命令,--command=start 启动服务,--command=stop 停止服务,--command=reset 重置服务,--command=help 查看帮助")
}
