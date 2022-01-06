package logger

import (
	"os"
	"testing"

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
