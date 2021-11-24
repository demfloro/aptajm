package main

import (
	"io"
	"log"
)

type botLogger struct {
	debug  bool
	logger *log.Logger
}

func (b botLogger) Log(v ...interface{}) {
	b.logger.Print(v...)
}

func (b botLogger) Logf(format string, v ...interface{}) {
	b.logger.Printf(format, v...)
}

func (b botLogger) Debug(format string, v ...interface{}) {
	if !b.debug {
		return
	}
	b.logger.Printf(format, v...)
}

func newLogger(writer io.Writer, debug bool) botLogger {
	return botLogger{
		debug:  debug,
		logger: log.New(writer, "", log.Lshortfile),
	}
}
