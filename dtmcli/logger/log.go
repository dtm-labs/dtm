package logger

import (
	"encoding/json"
	"errors"
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

// DefaultLogOutput is the default configuration for log output.
const (
	DefaultLogOutput = "default"
	StdErrLogOutput  = "stderr"
	StdOutLogOutput  = "stdout"
)

func init() {
	InitLog(os.Getenv("LOG_LEVEL"), nil, 0, "")
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
func InitLog(level string, outputs []string, logRotationEnable int64, logRotateConfigJSON string) {
	if len(outputs) == 0 {
		outputs = []string{DefaultLogOutput}
	}

	// parse outputs
	outputPaths := make([]string, 0)
	for _, v := range outputs {
		switch v {
		case DefaultLogOutput:
			outputPaths = append(outputPaths, StdOutLogOutput)

		case StdErrLogOutput:
			outputPaths = append(outputPaths, StdErrLogOutput)

		case StdOutLogOutput:
			outputPaths = append(outputPaths, StdOutLogOutput)

		default:
			var path string
			if logRotationEnable != 0 {
				// append rotate scheme to logs managed by lumberjack log rotation
				if v[0:1] == "/" {
					path = fmt.Sprintf("lumberjack:/%%2F%s", v[1:])
				} else {
					path = fmt.Sprintf("lumberjack:/%s", v)
				}
			} else {
				path = v
			}
			outputPaths = append(outputPaths, path)
		}
	}

	// setup log rotation
	if logRotationEnable != 0 {
		setupLogRotation(outputs, logRotateConfigJSON)
	}

	config := loadConfig(level)
	config.OutputPaths = outputPaths
	p, err := config.Build(zap.AddCallerSkip(1))
	FatalIfError(err)
	logger = p.Sugar()
}

type lumberjackSink struct {
	*lumberjack.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

// setupLogRotation initializes log rotation for a single file path target.
func setupLogRotation(logOutputs []string, logRotateConfigJSON string) {
	var lumberjackSink lumberjackSink
	outputFilePaths := 0
	for _, v := range logOutputs {
		switch v {
		case "stdout", "stderr":
			continue
		default:
			outputFilePaths++
		}
	}
	// log rotation requires file target
	if len(logOutputs) == 1 && outputFilePaths == 0 {
		FatalIfError(fmt.Errorf("log outputs requires a single file path when LogRotationConfigJSON is defined"))
	}
	// support max 1 file target for log rotation
	if outputFilePaths > 1 {
		FatalIfError(fmt.Errorf("log outputs requires a single file path when LogRotationConfigJSON is defined"))
	}

	if err := json.Unmarshal([]byte(logRotateConfigJSON), &lumberjackSink); err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		var syntaxError *json.SyntaxError
		switch {
		case errors.As(err, &syntaxError):
			FatalIfError(fmt.Errorf("improperly formatted log rotation config: %w", err))
		case errors.As(err, &unmarshalTypeError):
			FatalIfError(fmt.Errorf("invalid log rotation config: %w", err))
		}
	}
	err := zap.RegisterSink("lumberjack", func(u *url.URL) (zap.Sink, error) {
		lumberjackSink.Filename = u.Path[1:]
		return &lumberjackSink, nil
	})
	FatalIfError(err)
	return
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
