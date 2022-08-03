package logger

import (
	"github.com/dtm-labs/logger"
)

var (
	// WithLogger replaces default logger
	WithLogger = logger.WithLogger
	// InitLog is an initialization for a logger
	// level can be: debug info warn error
	InitLog = logger.InitLog
	// InitLog2 specify advanced log config
	InitLog2 = logger.InitLog2
	// Debugf log to level debug
	Debugf = logger.Debugf

	// Infof log to level info
	Infof = logger.Infof

	// Warnf log to level warn
	Warnf = logger.Warnf
	// Errorf log to level error
	Errorf = logger.Errorf

	// FatalfIf log to level error
	FatalfIf = logger.FatalfIf

	// FatalIfError if err is not nil, then log to level fatal and call os.Exit
	FatalIfError = logger.FatalIfError
)
