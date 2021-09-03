package ymlogger

import (
	"time"
)

// LogLevel defines the severity for LOG data type
type LogLevel byte

const (
	// DEBUG for debug level statements
	DEBUG LogLevel = iota
	// INFO for info level statements
	INFO
	// ERROR for error level statements
	ERROR
	// CRITICAL for critical level statements
	CRITICAL
)

func (logLevel LogLevel) String() string {
	switch logLevel {
	case CRITICAL:
		return "CRITICAL"
	case ERROR:
		return "ERROR"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

//FromString returns the enum based on the string of the log severity
func (logLevel LogLevel) FromString(severity string) LogLevel {
	switch severity {
	case "CRITICAL":
		return CRITICAL
	case "ERROR":
		return ERROR
	case "INFO":
		return INFO
	case "DEBUG":
		return DEBUG
	default:
		return INFO
	}
}

//BaseLoggerData defines the base structure with common fields
//that is built for sending out the message
type BaseLoggerData struct {
	RequestID   string
	LogTime     time.Time
	ProcessName string
	Hostname    string
	ProcessID   int
}

// LogData defines the specific fields for the log data type
type LogData struct {
	BaseLogger BaseLoggerData
	Level      string
	FileName   string
	LineNum    int
	Msg        string
}

// LoggerConf defines the service specific config for logger
type LoggerConf struct {
	ProcessName string `json:"process_name"`
	LogSeverity string `json:"log_severity"`
	LogFileName string `json:"log_file_name"`
	ConsoleLog  bool   `json:"console_log"`
}
