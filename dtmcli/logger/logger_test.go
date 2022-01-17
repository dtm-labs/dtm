package logger

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestInitLog(t *testing.T) {
	os.Setenv("DTM_DEBUG", "1")
	InitLog("debug", nil, 0, "")
	Debugf("a debug msg")
	Infof("a info msg")
	Warnf("a warn msg")
	Errorf("a error msg")
	FatalfIf(false, "nothing")
	FatalIfError(nil)

	InitLog("debug", []string{"test.log", "stdout"}, 0, "")
	Debugf("a debug msg to console and file")
	Infof("a info msg to console and file")
	Warnf("a warn msg to console and file")
	Errorf("a error msg to console and file")

	InitLog("debug", []string{"stdout", "stderr"}, 0, "")
	Debugf("a debug msg to stdout and stderr")
	Infof("a info msg to stdout and stderr")
	Warnf("a warn msg to stdout and stderr")
	Errorf("a error msg to stdout and stderr")

	InitLog("debug", []string{"test.log", "stdout"}, 1,
		"{\"maxsize\": 1, \"maxage\": 1, \"maxbackups\": 1, \"compress\": false}")
	Debugf("a debug msg to console and file with rotation")
	Infof("a info msg to console and file with rotation")
	Warnf("a warn msg to console and file with rotation")
	Errorf("a error msg to console and file with rotation")

	_ = os.Remove("test.log")
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
