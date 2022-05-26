package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"icapeg/config"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const (
	minLogLengthToFlush int = 1048576
	minTimeToFlush          = 10
	srvName                 = "icap-server"
)

type ZLogger struct {
	Logger           zerolog.Logger
	LogContent       *bytes.Buffer
	LogFlushTime     float64
	LoggingServerURL string
	LogStartTime     time.Time
	LoggingServer    *LoggingServer
	FileID           string
	LogMetaData      map[string]string
	TransitionID     string
}

// ExceedLoggingTime : ExceedLoggingTime checks if minimum time is reached to send logs to logging server.
func (z *ZLogger) ExceedLoggingTime() bool {
	duration := time.Since(z.LogStartTime).Seconds()
	return duration >= z.LogFlushTime*minTimeToFlush && len(z.LogContent.Bytes()) > 0
}

// ExceedLogSize : ExceedLogSize checks if logging content exceeds min length to send logs to logging server.
func (z *ZLogger) ExceedLogSize() bool {
	return len(z.LogContent.Bytes()) > minLogLengthToFlush
}

// readLogFiles : read the content of the logging file and create a glasswall logging format
func (z *ZLogger) readLogFiles() (tLog TransactionalLog, err error) {
	tLog.Events.Logs = map[string]TransactionalLogEvent{}
	tLog.Events.Type = srvName
	tLog.Metadata = z.LogMetaData

	scanner := bufio.NewScanner(z.LogContent)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		c := zLogOutput{}
		b := scanner.Bytes()
		err = json.Unmarshal(b, &c)
		if err != nil {
			return TransactionalLog{}, fmt.Errorf("could not unmarshall logs %v", err)
		}
		tLog.Events.Logs[c.Message] = TransactionalLogEvent{
			Value:        c.Value,
			Duration:     c.Duration,
			ErrorMessage: c.ErrorMessage,
			Time:         c.Time,
		}
		tLog.Events.Duration += c.Duration
	}
	return
}

// NewZLogger : create new zero logger object
func NewZLogger(conf *config.AppConfig) (*ZLogger, error) {
	zLogger := new(ZLogger)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("could not get the host name %v", err)
	}
	logMetaData := map[string]string{
		"processed-by": hostname,
		"service_name": srvName,
	}
	zLogger.LoggingServerURL = conf.LoggingServerURL
	zLogger.LogFlushTime = conf.LoggingFlushDuration
	zLogger.LogStartTime = time.Now()
	// setting time format in logs to be in Epoch
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// setting the logging level
	zLevel, err := zerolog.ParseLevel(conf.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("logging level %s provided is not supported by zerolog: %w", conf.LogLevel, err)
	}
	zLogger.LogContent = &bytes.Buffer{}
	zerolog.SetGlobalLevel(zLevel)
	multiWriter := zerolog.MultiLevelWriter(zLogger.LogContent, zerolog.ConsoleWriter{Out: os.Stdout})
	zlog.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
	zLogger.LoggingServer = NewLoggerClient()
	zLogger.LogMetaData = logMetaData
	return zLogger, nil
}
