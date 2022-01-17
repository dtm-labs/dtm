package logger

import (
	"os"
	"testing"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
)

func TestInitLog(t *testing.T) {
	os.Setenv("DTM_DEBUG", "1")
	InitLog("debug")
	Debugf("a debug msg")
	Infof("a info msg")
	Warnf("a warn msg")
	Errorf("a error msg")
	FatalfIf(false, "nothing")
	FatalIfError(nil)
}

func TestWithLogger(t *testing.T) {
	logger := zap.NewExample().Sugar()
	WithLogger(logger)
	Debugf("a debug msg")
	Infof("a info msg")
	Warnf("a warn msg")
	Errorf("a error msg")
	FatalfIf(false, "nothing")
	FatalIfError(nil)
}

func TestInitRotateLog(t *testing.T) {
	os.Setenv("DTM_DEBUG", "1")
	ll := lumberjack.Logger{
		Filename:   "test.log",
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
	}
	InitRotateLog("debug", &ll)
	Debugf("a debug msg")
	Infof("a info msg")
	Warnf("a warn msg")
	Errorf("a error msg")
	FatalfIf(false, "nothing")
	FatalIfError(nil)
	s := lumberjackSink{&ll}
	_ = s.Sync()
	_ = os.Remove("test.log")
}
