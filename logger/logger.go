package logger

import (
	"log"
	"os"
)

var (
	logFile *os.File
)

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
func LogToScreen(a ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Println(a...)
}

// LogToFile logs the parameters to the log file
func LogToFile(a ...interface{}) {
	log.SetOutput(logFile)
	log.Println(a...)
}

// LogfToScreen logs the formatted parameters to screen
func LogfToScreen(str string, a ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Printf(str, a...)
}

// LogfToFile logs the formatted parameters to the log file
func LogfToFile(str string, a ...interface{}) {
	log.SetOutput(logFile)
	log.Printf(str, a...)
}

// LogToAll logs the given parameters to both the screen and the file
func LogToAll(a ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Println(a...)
	log.SetOutput(logFile)
	log.Println(a...)
}

// LogfToAll logs the given formatted parameters to both the screen and the file
func LogfToAll(str string, a ...interface{}) {
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

// LogFile returns the log file instance
func LogFile() *os.File {
	return logFile
}
