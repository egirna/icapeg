package logger

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

// The log levels
const (
	LogLevelInfo  = "info"
	LogLevelDebug = "debug"
	LogLevelError = "error"
	LogLevelNone  = "none"
)

// Logger is the one responsible for perfoming the log writes
type Logger struct {
	AllowedLogLevels map[string]bool
}

var (
	logFile  *os.File
	logLevel string
)

// NewLogger is the factory function for generating a new Logger instance
func NewLogger(allowedLogLevels ...string) *Logger {
	l := &Logger{
		AllowedLogLevels: make(map[string]bool),
	}

	for _, ll := range allowedLogLevels {
		l.AllowedLogLevels[ll] = true
	}
	return l
}

// SetLogLevel sets the log level for the app
func SetLogLevel(l string) {
	logLevel = l
}

// SetLogFile prepares a text file for logs
func SetLogFile(filename string) error {
	if filename == "" {
		filename = "logs.txt"
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

// LogToScreen logs the parameters to screen
func (l *Logger) LogToScreen(a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	log.SetOutput(os.Stdout)
	log.Println(a...)
}

// LogToFile logs the parameters to the log file
func (l *Logger) LogToFile(a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	log.SetOutput(logFile)
	log.Println(a...)
}

// LogfToScreen logs the formatted parameters to screen
func (l *Logger) LogfToScreen(str string, a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	log.SetOutput(os.Stdout)
	log.Printf(str, a...)
}

// LogfToFile logs the formatted parameters to the log file
func (l *Logger) LogfToFile(str string, a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	log.SetOutput(logFile)
	log.Printf(str, a...)
}

// LogToAll logs the given parameters to both the screen and the file
func (l *Logger) LogToAll(a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	log.SetOutput(os.Stdout)
	log.Println(a...)
	log.SetOutput(logFile)
	log.Println(a...)
}

// LogfToAll logs the given formatted parameters to both the screen and the file
func (l *Logger) LogfToAll(str string, a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	log.SetOutput(os.Stdout)
	log.Printf(str, a...)
	log.SetOutput(logFile)
	log.Printf(str, a...)
}

// LogFatalToScreen logs the given parameters to the screen with a fatal
func LogFatalToScreen(a ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Fatal(a...)
}

// LogFatalToFile logs the given parameters to the file with a fatal
func LogFatalToFile(a ...interface{}) {
	log.SetOutput(logFile)
	log.Fatal(a...)
}

// DumpToFile spew dumps the parameters to the file
func (l *Logger) DumpToFile(a ...interface{}) {

	if logLevel == LogLevelNone {
		return
	}

	if _, allowed := l.AllowedLogLevels[logLevel]; !allowed && len(l.AllowedLogLevels) > 0 {
		return
	}

	spew.Fdump(logFile, a...)
}

// LogFile returns the log file instance
func LogFile() *os.File {
	return logFile
}
