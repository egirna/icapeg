package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"icapeg/readValues"
	"os"
)

var Logger *zap.Logger
var logLevels map[string]zapcore.Level

func initLogLevels() {
	logLevels = make(map[string]zapcore.Level)
	logLevels["debug"] = zapcore.DebugLevel
	logLevels["info"] = zapcore.InfoLevel
	logLevels["warn"] = zapcore.WarnLevel
	logLevels["error"] = zapcore.ErrorLevel
	logLevels["fatal"] = zapcore.FatalLevel
}

func getZapCoreLevel(logLevel string) zapcore.Level {
	return logLevels[logLevel]
}
func InitLogger() {
	initLogLevels()
	logLevel := readValues.ReadValuesString("app.log_level")
	printToConsole := readValues.ReadValuesBool("app.write_logs_to_console")
	logFile := os.Stdout
	if !printToConsole {
		logFile, _ = os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY, 0644)
	}
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	writer := zapcore.AddSync(logFile)
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, getZapCoreLevel(logLevel)),
	)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	Logger.Debug("logging is configured successfully")
}
