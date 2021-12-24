package logger

import (
	"os"
	"testing"
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
