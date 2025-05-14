package qant_api_bybit

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelError
	LogLevelFatal
)

type Logger interface {
	Info(format string, v ...any)
	Debug(format string, v ...any)
	Error(format string, v ...any)
	Fatal(format string, v ...any)
	SetLogLevel(level LogLevel)
}

type StdLogger struct {
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	name        string
	logLevel    LogLevel
}

func BasicLogger(name string) Logger {
	return &StdLogger{
		infoLogger: log.New(os.Stdout, fmt.Sprintf("INFO:  %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		debugLogger: log.New(os.Stdout, fmt.Sprintf("DEBUG: %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		errorLogger: log.New(os.Stderr, fmt.Sprintf("ERROR: %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		fatalLogger: log.New(os.Stderr, fmt.Sprintf("FATAL: %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		logLevel: LogLevelDebug,
	}
}

func (l *StdLogger) SetLogLevel(level LogLevel) {
	l.logLevel = level
}

func (l *StdLogger) Info(format string, v ...any) {
	if l.logLevel <= LogLevelInfo {
		l.infoLogger.Printf(format, v...)
	}
}

func (l *StdLogger) Debug(format string, v ...any) {
	if l.logLevel <= LogLevelDebug {
		l.debugLogger.Printf(format, v...)
	}
}

func (l *StdLogger) Error(format string, v ...any) {
	if l.logLevel <= LogLevelError {
		l.errorLogger.Printf(format, v...)
	}
}

func (l *StdLogger) Fatal(format string, v ...any) {
	if l.logLevel <= LogLevelFatal {
		l.fatalLogger.Fatalf(format, v...)
	}
}
