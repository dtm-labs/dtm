package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger = nil

func init() {
	InitLog("info")
}

// InitLog is a initialization for a logger
// level can be: debug info warn error
func InitLog(level string) {
	config := zap.NewProductionConfig()
	err := config.Level.UnmarshalText([]byte(level))
	FatalIfError(err)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if os.Getenv("DTM_DEBUG") != "" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	p, err := config.Build(zap.AddCallerSkip(1))
	FatalIfError(err)
	logger = p.Sugar()
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
