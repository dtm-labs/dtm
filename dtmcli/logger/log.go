package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime/debug"
	"strings"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//var logger *zap.SugaredLogger = nil

var logger Logger

const (
	// StdErr is the default configuration for log output.
	StdErr = "stderr"
	// StdOut configuration for log output
	StdOut = "stdout"
)

func init() {
	InitLog(os.Getenv("LOG_LEVEL"))
}

// Logger logger interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// WithLogger replaces default logger
func WithLogger(log Logger) {
	logger = log
}

// InitLog is an initialization for a logger
// level can be: debug info warn error
func InitLog(level string) {
	InitLog2(level, StdOut, 0, "")
}

// InitLog2 specify advanced log config
func InitLog2(level string, outputs string, logRotationEnable int64, logRotateConfigJSON string) {
	outputPaths := strings.Split(outputs, ",")
	for i, v := range outputPaths {
		if logRotationEnable != 0 && v != StdErr && v != StdOut {
			outputPaths[i] = fmt.Sprintf("lumberjack://%s", v)
		}
	}

	if logRotationEnable != 0 {
		setupLogRotation(logRotateConfigJSON)
	}

	config := loadConfig(level)
	config.OutputPaths = outputPaths
	p, err := config.Build(zap.AddCallerSkip(1))
	FatalIfError(err)
	logger = p.Sugar()
}

type lumberjackSink struct {
	lumberjack.Logger
}

func (*lumberjackSink) Sync() error {
	return nil
}

// setupLogRotation initializes log rotation for a single file path target.
func setupLogRotation(logRotateConfigJSON string) {
	err := zap.RegisterSink("lumberjack", func(u *url.URL) (zap.Sink, error) {
		var conf lumberjackSink
		err := json.Unmarshal([]byte(logRotateConfigJSON), &conf)
		FatalfIf(err != nil, "bad config LogRotateConfigJSON: %v", err)
		conf.Filename = u.Host + u.Path
		return &conf, nil
	})
	FatalIfError(err)
}

func loadConfig(logLevel string) zap.Config {
	config := zap.NewProductionConfig()
	err := config.Level.UnmarshalText([]byte(logLevel))
	FatalIfError(err)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if os.Getenv("DTM_DEBUG") != "" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return config
}

// Debugf log to level debug
func Debugf(fmt string, args ...interface{}) {
	logger.Debugf(fmt, args...)
}

// Infof log to level info
func Infof(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// Warnf log to level warn
func Warnf(fmt string, args ...interface{}) {
	logger.Warnf(fmt, args...)
}

// Errorf log to level error
func Errorf(fmt string, args ...interface{}) {
	logger.Errorf(fmt, args...)
}

// FatalfIf log to level error
func FatalfIf(cond bool, fmt string, args ...interface{}) {
	if !cond {
		return
	}
	debug.PrintStack()
	log.Fatalf(fmt, args...)
}

// FatalIfError if err is not nil, then log to level fatal and call os.Exit
func FatalIfError(err error) {
	FatalfIf(err != nil, "fatal error: %v", err)
}
