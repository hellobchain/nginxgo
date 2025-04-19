package nginxgo

import (
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/hellobchain/nginxgo/common/constant"
	"github.com/hellobchain/nginxgo/core"
	"github.com/hellobchain/nginxgo/pkg/utils"
	"github.com/hellobchain/wswlog/wlogging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var logger = wlogging.MustGetLoggerWithoutName()
var engine *core.Engine // 声明一个 Engine 变量
var mu sync.Mutex       // 互斥锁，用于保护 engine 变量的访问
const (
	// flagNameOfConfigFilepath 是配置文件路径的标志名称
	flagNameOfConfigFilepath = "nginxgo-config"

	// flagNameShortHandOFConfigFilepath 是配置文件路径的短名称
	flagNameShortHandOFConfigFilepath = "c"

	// readPidFile 读取 pid 文件
	flagNameOfPidFilePath = "pid-file"

	// pidFilePath 是 pid 文件路径
	flagNameShortHandOfPidFilePath = "p"
)

var pidFilePath string

func operateCMD(command string, pid int) {
	switch command {
	case constant.CMD_START:
		mu.Lock()
		if engine == nil {
			engine = core.Init()
			go engine.Start()
		} else {
			logger.Warn("nginxgo: engine already started")
		}
		mu.Unlock()
		handleResetSignals()
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
		logger.Info("nginxgo help")
		logger.Info("nginxgo start")
		logger.Info("nginxgo stop")
		logger.Info("nginxgo reset")
	default:
		logger.Error("nginxgo: error command")
	}
}

func StartMain() {
	mainCmd := &cobra.Command{Use: "nginxgo"}
	mainCmd.AddCommand(startCMD())
	mainCmd.AddCommand(stopCMD())
	mainCmd.AddCommand(resetCMD())
	mainCmd.AddCommand(helpCMD())
	err := mainCmd.Execute()
	if err != nil {
		logger.Error("nginxgo: error command")
	}
}

func startCMD() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Startup nginxgo",
		Long:  "Startup nginxgo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			operateCMD(constant.CMD_START, -1)
			logger.Error("nginxgo exit")
			return nil
		},
	}

	attachFlags(startCmd, []string{flagNameOfConfigFilepath})
	return startCmd
}

func handleResetSignals() {
	// 监听SIGHUP信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	utils.HandleSignals(logger)
	pidStr := strconv.Itoa(os.Getpid())
	os.WriteFile(pidFilePath, []byte(pidStr), 0644)
	for range signalChan {
		// 收到 SIGHUP 信号，执行重置操作
		logger.Info("nginxgo: received SIGHUP signal, resetting engine")
		operateCMD(constant.CMD_RESET, -1)
	}
}

func stopCMD() *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "stop nginxgo",
		Long:  "stop nginxgo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			operateCMD(constant.CMD_STOP, 1)
			return nil
		},
	}
	attachFlags(stopCmd, []string{flagNameOfConfigFilepath})
	return stopCmd
}

func resetCMD() *cobra.Command {
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "reset nginxgo",
		Long:  "reset nginxgo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			operateCMD(constant.CMD_RESET, 1)
			return nil
		},
	}
	attachFlags(resetCmd, []string{flagNameOfConfigFilepath, flagNameOfPidFilePath})
	return resetCmd
}

func helpCMD() *cobra.Command {
	helpCmd := &cobra.Command{
		Use:   "help",
		Short: "help nginxgo",
		Long:  "help nginxgo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			operateCMD(constant.CMD_HELP, -1)
			return nil
		},
	}
	return helpCmd
}

func initFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.StringVarP(&core.NginxConfigFilepath, flagNameOfConfigFilepath, flagNameShortHandOFConfigFilepath,
		"./configs/config.cfg", "specify config file path, if not set, default use ./configs/config.cfg")
	flags.StringVarP(&pidFilePath, flagNameOfPidFilePath, flagNameShortHandOfPidFilePath,
		"./nginxgo.pid", "specify pid file path, if not set, default use ./nginxgo.pid")
	return flags
}

func attachFlags(cmd *cobra.Command, flagNames []string) {
	flags := initFlagSet()
	cmdFlags := cmd.Flags()
	for _, flagName := range flagNames {
		if flag := flags.Lookup(flagName); flag != nil {
			cmdFlags.AddFlag(flag)
		}
	}
}
