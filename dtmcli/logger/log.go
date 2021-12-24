package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger = nil

func init() {
	InitLog()
}

// InitLog is a initialization for a logger
func InitLog() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if os.Getenv("DTM_DEBUG") != "" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	p, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		log.Fatal("create logger failed: ", err)
	}
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
	logger.Fatalf(fmt, args...)
}

// FatalIfError if err is not nil, then log to level fatal and call os.Exit
func FatalIfError(err error) {
	FatalfIf(err != nil, "fatal error: %v", err)
}
