package mgobase

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

type logger interface {
	Print(v ...interface{})
	Warn(v ...interface{})
}

// mgoLogger implements the mgo.log_Logger interface.
type mgoLogger struct {
	logger
}

func (l *mgoLogger) Output(calldepth int, s string) error {
	l.logger.Print(s)
	return nil
}

var (
	globalLogger *mgoLogger
)

// SetDebug sets the mgo.SetDebug.
func SetDebug(debug bool) {
	mgo.SetDebug(debug)
}

// SetLogger sets the logger so that mgobase can logger the important info.
func SetLogger(logger logger) {
	if logger != nil {
		globalLogger = &mgoLogger{
			logger: logger,
		}
	}
}

func info(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Print(v...)
	}
}

func infof(s string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Print(fmt.Sprintf(s, v...))
	}
}

func infoln(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Print(fmt.Sprintln(v...))
	}
}

func warn(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(v...)
	}
}

func warnf(s string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(fmt.Sprintf(s, v...))
	}
}

func warnln(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(fmt.Sprintln(v...))
	}
}
