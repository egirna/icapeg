package logger

import "time"

type TransactionalLogEvent struct {
	Value        string        `json:"value"`
	Duration     float32       `json:"duration"`
	ErrorMessage string        `json:"error,omitempty"`
	Time         time.Duration `json:"time,omitempty"`
}

type TransactionalLog struct {
	Metadata map[string]string `json:"metadata,omitempty"`
	Events   Events            `json:"events,omitempty"`
}

type Events struct {
	Duration float32                          `json:"duration,omitempty"`
	Type     string                           `json:"type,omitempty"`
	Logs     map[string]TransactionalLogEvent `json:"logs,omitempty"`
}

type zLogOutput struct {
	Value        string        `json:"value"`
	Duration     float32       `json:"duration"`
	Message      string        `json:"message,omitempty"`
	Time         time.Duration `json:"time"`
	ErrorMessage string        `json:"error,omitempty"`
}
