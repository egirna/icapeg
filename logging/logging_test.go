package logging

import (
	"os"
	"testing"
)

func TestInitializeLogger(t *testing.T) {
	// Test case 1: Initialize logger with console logging enabled
	InitializeLogger("debug", true)
	if Logger == nil {
		t.Error("Logger not initialized")
	}

	// Test case 2: Initialize logger with console logging disabled
	InitializeLogger("debug", false)
	if Logger == nil {
		t.Error("Logger not initialized")
	}

	// Test case 3: Check if logs directory and file are created
	_, err := os.Stat("./logs/logs.json")
	if os.IsNotExist(err) {
		t.Error("Logs directory or file does not exist")
	}

	// Add more test cases as needed...
}

// Example usage of the logger in another test function
func TestLoggerUsage(t *testing.T) {
	// Initialize logger
	InitializeLogger("debug", true)

	// Example log message
	Logger.Info("This is an example log message")

	// Add assertions or checks based on your actual usage of the logger
}
