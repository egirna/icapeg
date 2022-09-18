package logging

import (
	"fmt"
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
	_, err := os.Create("logs/logs.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	logFile, _ := os.OpenFile("logs/logs.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
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
