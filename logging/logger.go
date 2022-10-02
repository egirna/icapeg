package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger *zap.Logger

func InitializeLogger(logLevel string, writeLogsToConsole bool) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	os.Mkdir("./logs", os.ModePerm)

	logFile, _ := os.OpenFile("logs/logs.json", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel, _ := zapcore.ParseLevel(logLevel)
	var core zapcore.Core
	if writeLogsToConsole {
		consoleEncoder := zapcore.NewConsoleEncoder(config)
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		)
	}

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
