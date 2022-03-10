## Logger

Logger handles logging of `go-icap-server`. It uses [zerolog](https://github.com/rs/zerolog) as logging client, collect icap-server logs and sends it to the GlassWall Logging Service.

## Log Level Supported

| Level      | Explanation                               |
|------------|-------------------------------------------|
| `debug`    | DebugLevel defines debug log level        |
| `info`     | InfoLevel defines info log level          |
| `warn`     | WarnLevel defines warn log level          |
| `error`    | ErrorLevel defines error log level        |
| `fatal`    | FatalLevel defines fatal log level        |
| `panic`    | PanicLevel defines panic log level        |
| `disabled` | Disabled disables the zlogger              |
| ``         | Empty/NoLevel defines an absent log level |

## TimeFormat Supported

| TimeFormat            | Explanation                                                       |
|-----------------------|-------------------------------------------------------------------|
| `TimeFormatUnix`      | time fields serialized as Unix timestamp integers                 |
| `TimeFormatUnixMs`    | time fields serialized as Unix timestamp integers in milliseconds |
| `TimeFormatUnixMicro` | time fields serialized as Unix timestamp integers in microseconds |

```text
Note: If time format not set explicitly the format will be string e.g 2006-01-02T15:04:05Z07:00
```

## Flush Logs to External Service

Logger flushes logs to external service `http://logging-transaction-logging-api:8080/upload` on two conditions:
- If the duration since start of the application cross a threshold (`log_flush_duration`)
- If the log content cross a threshold (`minLogLengthToFlush` currently set to `1048576`)

Log Content schema looks like:
```text
Value        string        `json:"value"`
Duration     float32       `json:"duration"`
Message      string        `json:"message,omitempty"`
Time         time.Duration `json:"time"`
ErrorMessage string        `json:"error_message,omitempty"`
```