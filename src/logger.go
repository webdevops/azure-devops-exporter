package main

import (
	"os"
	"log"
	"fmt"
)

const (
	LoggerLogPrefix = ""
	LoggerLogPrefixError = "[ERROR] "
)

type DaemonLogger struct {
	*log.Logger
}

var (
	Verbose bool
)

func CreateDaemonLogger(flags int) *DaemonLogger {
	return &DaemonLogger{log.New(os.Stdout, LoggerLogPrefix, flags)}
}

func CreateDaemonErrorLogger(flags int) *DaemonLogger {
	return &DaemonLogger{log.New(os.Stderr, LoggerLogPrefix, flags)}
}

func (l *DaemonLogger) Verbose(message string, sprintf ...interface{}) {
	if Verbose {
		if len(sprintf) > 0 {
			message = fmt.Sprintf(message, sprintf...)
		}

		l.Println(message)
	}
}

func (l *DaemonLogger) Messsage(message string, sprintf ...interface{}) {
	if len(sprintf) > 0 {
		message = fmt.Sprintf(message, sprintf...)
	}

	l.Println(message)
}

// Log error object as message
func (l *DaemonLogger) Error(msg string, err error) {
	l.Println(fmt.Sprintf("%v%v: %v", LoggerLogPrefixError, msg, err))
}
