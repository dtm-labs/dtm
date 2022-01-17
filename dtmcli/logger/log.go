package logger

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//var logger *zap.SugaredLogger = nil

var logger Logger

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

type lumberjackSink struct {
	*lumberjack.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

// WithLogger replaces default logger
func WithLogger(log Logger) {
	logger = log
}

// InitLog is an initialization for a logger
// level can be: debug info warn error
func InitLog(level string) {
	config := loadConfig(level)
	p, err := config.Build(zap.AddCallerSkip(1))
	FatalIfError(err)
	logger = p.Sugar()
}

// InitRotateLog is an initialization for a rotated logger by lumberjack
func InitRotateLog(logLevel string, ll *lumberjack.Logger) {
	config := loadConfig(logLevel)
	config.OutputPaths = []string{fmt.Sprintf("lumberjack:%s", ll.Filename), "stdout"}
	err := zap.RegisterSink("lumberjack", func(*url.URL) (zap.Sink, error) {
		return lumberjackSink{
			Logger: ll,
		}, nil
	})
	FatalIfError(err)

	p, err := config.Build(zap.AddCallerSkip(1))
	FatalIfError(err)
	logger = p.Sugar()
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
	log.Fatalf(fmt, args...)
}

// FatalIfError if err is not nil, then log to level fatal and call os.Exit
func FatalIfError(err error) {
	FatalfIf(err != nil, "fatal error: %v", err)
}
