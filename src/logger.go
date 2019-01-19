package main

import (
	"fmt"
	"github.com/google/logger"
	"io/ioutil"
)

type DaemonLogger struct {
	*logger.Logger
	verbose bool
}

func NewLogger(flags int, verbose bool) *DaemonLogger {
	instance := &DaemonLogger{
		Logger: logger.Init("Verbose", true, false, ioutil.Discard),
		verbose: verbose,
	}
	logger.SetFlags(flags)

	return instance
}

func (l *DaemonLogger) Verbosef(message string, sprintf ...interface{}) {
	if l.verbose {
		if len(sprintf) > 0 {
			message = fmt.Sprintf(message, sprintf...)
		}

		l.InfoDepth(1, message)
	}
}
