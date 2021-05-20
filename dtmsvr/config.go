package dtmsvr

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yedf/dtm/common"
)

// formatter 自定义formatter
type formatter struct{}

// Format 进行格式化
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer = &bytes.Buffer{}
	if entry.Buffer != nil {
		b = entry.Buffer
	}
	n := time.Now()
	ts := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%03d", n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), n.Second(), n.Nanosecond()/1000000)
	var file string
	var line int
	for i := 1; ; i++ {
		_, file, line, _ = runtime.Caller(i)
		if strings.Contains(file, "dtm") {
			break
		}
	}
	b.WriteString(fmt.Sprintf("%s %s:%d %s\n", ts, path.Base(file), line, entry.Message))
	return b.Bytes(), nil
}

var configLoaded = false

func LoadConfig() {
	if configLoaded {
		return
	}
	configLoaded = true
	logrus.SetFormatter(&formatter{})
	_, file, _, _ := runtime.Caller(0)
	viper.SetConfigFile(filepath.Dir(file) + "/dtmsvr.yml")
	err := viper.ReadInConfig()
	common.PanicIfError(err)
}
