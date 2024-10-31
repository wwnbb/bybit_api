package qant_api_bybit

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Info(format string, v ...any)
	Debug(format string, v ...any)
	Error(format string, v ...any)
	Fatal(format string, v ...any)
}

type StdLogger struct {
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	name        string
}

func BasicLogger(name string) Logger {
	return StdLogger{
		infoLogger: log.New(os.Stdout, fmt.Sprintf("INFO:  %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		debugLogger: log.New(os.Stdout, fmt.Sprintf("DEBUG: %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		errorLogger: log.New(os.Stderr, fmt.Sprintf("ERROR: %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
		fatalLogger: log.New(os.Stderr, fmt.Sprintf("FATAL: %s > ", name),
			log.Lmsgprefix|log.LstdFlags|log.Lmicroseconds),
	}
}

func (l StdLogger) Info(format string, v ...any)  { l.infoLogger.Printf(format, v...) }
func (l StdLogger) Debug(format string, v ...any) { l.debugLogger.Printf(format, v...) }
func (l StdLogger) Error(format string, v ...any) { l.errorLogger.Printf(format, v...) }
func (l StdLogger) Fatal(format string, v ...any) { l.fatalLogger.Fatalf(format, v...) }
