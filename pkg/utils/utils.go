package utils

import (
	"bytes"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func handleSignals(handlers map[os.Signal]func(), logger Logger) {
	var signals []os.Signal
	for sig := range handlers {
		signals = append(signals, sig)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)

	for sig := range signalChan {
		logger.Debugf("Received signal: %d (%s)", sig, sig)
		handlers[sig]()
	}
}

func addPlatformSignals(sigs map[os.Signal]func(), logger Logger) map[os.Signal]func() {
	sigs[syscall.SIGUSR1] = func() { logGoRoutines(logger) }
	return sigs
}

type Logger interface {
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	DPanicw(msg string, kvPairs ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(msg string, kvPairs ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, kvPairs ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, kvPairs ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, kvPairs ...interface{})
	Panic(args ...interface{})
	Panicf(template string, args ...interface{})
	Panicw(msg string, kvPairs ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, kvPairs ...interface{})
	Warning(args ...interface{})
	Warningf(template string, args ...interface{})
	// for backwards compatibility
	Critical(args ...interface{})
	Criticalf(template string, args ...interface{})
	Notice(args ...interface{})
	Noticef(template string, args ...interface{})
}

func captureGoRoutines() (string, error) {
	var buf bytes.Buffer
	err := pprof.Lookup("goroutine").WriteTo(&buf, 2)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func logGoRoutines(logger Logger) {
	output, err := captureGoRoutines()
	if err != nil {
		logger.Errorf("failed to capture go routines: %s", err)
		return
	}

	logger.Debugf("Go routines report:\n%s", output)
}

func HandleSignals(logger Logger) {
	go handleSignals(addPlatformSignals(map[os.Signal]func(){
		syscall.SIGINT: func() {
			logger.Info("释放资源")
			os.Exit(0)
		},
		syscall.SIGTERM: func() {
			logger.Info("释放资源")
			os.Exit(0)
		},
	}, logger), logger)
}
