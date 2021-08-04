package dtmcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// P2E panic to error
func P2E(perr *error) {
	if x := recover(); x != nil {
		if e, ok := x.(error); ok {
			*perr = e
		} else {
			panic(x)
		}
	}
}

// E2P error to panic
func E2P(err error) {
	if err != nil {
		panic(err)
	}
}

// CatchP catch panic to error
func CatchP(f func()) (rerr error) {
	defer P2E(&rerr)
	f()
	return nil
}

// PanicIf name is clear
func PanicIf(cond bool, err error) {
	if cond {
		panic(err)
	}
}

// MustAtoi 走must逻辑
func MustAtoi(s string) int {
	r, err := strconv.Atoi(s)
	if err != nil {
		E2P(errors.New("convert to int error: " + s))
	}
	return r
}

// OrString return the first not empty string
func OrString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// If ternary operator
func If(condition bool, trueObj interface{}, falseObj interface{}) interface{} {
	if condition {
		return trueObj
	}
	return falseObj
}

// MustMarshal checked version for marshal
func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	E2P(err)
	return b
}

// MustMarshalString string version of MustMarshal
func MustMarshalString(v interface{}) string {
	return string(MustMarshal(v))
}

// MustUnmarshal checked version for unmarshal
func MustUnmarshal(b []byte, obj interface{}) {
	err := json.Unmarshal(b, obj)
	E2P(err)
}

// MustUnmarshalString string version of MustUnmarshal
func MustUnmarshalString(s string, obj interface{}) {
	MustUnmarshal([]byte(s), obj)
}

// MustRemarshal marshal and unmarshal, and check error
func MustRemarshal(from interface{}, to interface{}) {
	b, err := json.Marshal(from)
	E2P(err)
	err = json.Unmarshal(b, to)
	E2P(err)
}

// Logf 输出日志
func Logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
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
	fmt.Printf("%s %s:%d %s\n", ts, path.Base(file), line, msg)
}

// LogRedf 采用红色打印错误类信息
func LogRedf(fmt string, args ...interface{}) {
	Logf("\x1b[31m\n"+fmt+"\x1b[0m\n", args...)
}

// RestyClient the resty object
var RestyClient = resty.New()

func init() {
	// RestyClient.SetTimeout(3 * time.Second)
	// RestyClient.SetRetryCount(2)
	// RestyClient.SetRetryWaitTime(1 * time.Second)
	RestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.URL = MayReplaceLocalhost(r.URL)
		Logf("requesting: %s %s %v %v", r.Method, r.URL, r.Body, r.QueryParam)
		return nil
	})
	RestyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		Logf("requested: %s %s %s", r.Method, r.URL, resp.String())
		return nil
	})
}

// CheckRestySuccess panic if error or resp not success
func CheckRestySuccess(resp *resty.Response, err error) {
	E2P(err)
	if !strings.Contains(resp.String(), "SUCCESS") {
		panic(fmt.Errorf("resty response not success: %s", resp.String()))
	}
}

// MustGetwd must version of os.Getwd
func MustGetwd() string {
	wd, err := os.Getwd()
	E2P(err)
	return wd
}

// GetCurrentCodeDir name is clear
func GetCurrentCodeDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

// GetProjectDir name is clear
func GetProjectDir() string {
	_, file, _, _ := runtime.Caller(1)
	for ; !strings.HasSuffix(file, "/dtm"); file = filepath.Dir(file) {
	}
	return file
}

// GetFuncName get current call func name
func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// MayReplaceLocalhost when run in docker compose, change localhost to host.docker.internal for accessing host network
func MayReplaceLocalhost(host string) string {
	if os.Getenv("IS_DOCKER_COMPOSE") != "" {
		return strings.Replace(host, "localhost", "host.docker.internal", 1)
	}
	return host
}
