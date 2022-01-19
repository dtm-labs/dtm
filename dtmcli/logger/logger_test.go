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

	InitLog2("debug", "test.log,stderr", 0, "")
	Debugf("a debug msg to console and file")

	InitLog2("debug", "test2.log,/tmp/dtm-test1.log,/tmp/dtm-test.log,stdout,stderr", 1,
		"{\"maxsize\": 1, \"maxage\": 1, \"maxbackups\": 1, \"compress\": false}")
	Debugf("a debug msg to /tmp/dtm-test.log and test2.log and stdout and stderr")

	// _ = os.Remove("test.log")
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
