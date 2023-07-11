package logging

import (
	"log/syslog"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerSettings struct {
	LogLevel         string
	LogFilePath      string
	LogFileMaxSize   int
	LogFileMaxBackus int
	LogFileMaxAge    int
	LogToSyslog      bool
	LogToConsole     bool
}

var Logger *zap.Logger

func InitializeLogger(loggerSettings LoggerSettings) {
	logLevel, _ := zapcore.ParseLevel(loggerSettings.LogLevel)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   loggerSettings.LogFilePath,
		MaxSize:    loggerSettings.LogFileMaxSize,
		MaxBackups: loggerSettings.LogFileMaxBackus,
		MaxAge:     loggerSettings.LogFileMaxAge,
	})
	syslogWriter, _ := syslog.New(syslog.LOG_ERR|syslog.LOG_LOCAL0, "icapeg")

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		fileWriter,
		logLevel,
	)
	syslogCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(syslogWriter),
		logLevel,
	)
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		logLevel,
	)

	core := zapcore.NewTee(
		fileCore,
	)
	if loggerSettings.LogToSyslog {
		core = zapcore.NewTee(
			core,
			syslogCore,
		)
	}
	if loggerSettings.LogToConsole {
		core = zapcore.NewTee(
			core,
			consoleCore,
		)
	}

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
