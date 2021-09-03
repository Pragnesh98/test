package ymlogger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

//muLogger is the mutext for the logger writer. Serializes the writes to the file
var muLogger sync.Mutex
var logger io.Writer

//processName is a global variable in this package for process name. Will be printed with each log message.
var processName string

//logSeverity is a global variable in this package for the minimum log level.
//Logs will only be printed if they are above this level
var logSeverity = DEBUG

//logFileName is a global variable in this package for the log filename.
var logFileName string

//hostname is a global variable in this package which holds the hostname reported by the kernel.
var hostname string

//processID is a global variable in this package which holds the pid of process reported by the kernel.
var processID int

//consoleLog is a global variable in this package which will hold the choice of console logging
var consoleLog bool

// LogError logs all the error level statments
func LogError(requestID string, v ...interface{}) {
	Log(requestID, 2, ERROR, v...)
}

//LogCritical logs all the critical level statements
func LogCritical(requestID string, v ...interface{}) {
	Log(requestID, 2, CRITICAL, v...)
}

//LogInfo logs all the info level statements
func LogInfo(requestID string, v ...interface{}) {
	Log(requestID, 2, INFO, v...)
}

//LogDebug logs all the debug level statements
func LogDebug(requestID string, v ...interface{}) {
	Log(requestID, 2, DEBUG, v...)
}

//LogErrorf logs all the error level statements in given format
func LogErrorf(requestID string, format string, v ...interface{}) {
	Logf(requestID, 2, ERROR, format, v...)
}

//LogCriticalf logs all the critical level statements in given format
func LogCriticalf(requestID string, format string, v ...interface{}) {
	Logf(requestID, 2, CRITICAL, format, v...)
}

//LogInfof logs all the info level statements in given format
func LogInfof(requestID string, format string, v ...interface{}) {
	Logf(requestID, 2, INFO, format, v...)
}

//LogDebugf logs all the debug level statements in given format
func LogDebugf(requestID string, format string, v ...interface{}) {
	Logf(requestID, 2, DEBUG, format, v...)
}

//Log logs all statements without formatting
func Log(requestID string, stackLevel int, logLevel LogLevel, v ...interface{}) {
	var msg = ""
	if len(v) > 0 {
		msg = fmt.Sprint(v...)
	}
	if _, filename, line, ok := runtime.Caller(stackLevel); ok == true {
		loggerLog(requestID, filename, line, logLevel, msg)
		return
	}
	loggerLog(requestID, "", 0, logLevel, msg)
}

//Logf logs all statements without formatting
func Logf(requestID string, stackLevel int, logLevel LogLevel, format string, v ...interface{}) {
	var msg = ""
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	if _, filename, line, ok := runtime.Caller(stackLevel); ok == true {
		loggerLog(requestID, filename, line, logLevel, msg)
		return
	}
	loggerLog(requestID, "", 0, logLevel, msg)
}

func loggerLog(
	requestID string,
	filename string,
	line int,
	level LogLevel,
	v string,
) {
	if level < logSeverity {
		return
	}
	exolog := LogData{
		BaseLogger: BaseLoggerData{
			RequestID:   requestID,
			LogTime:     time.Now(),
			Hostname:    hostname,
			ProcessName: processName,
			ProcessID:   processID,
		},
		Level:    level.String(),
		FileName: filepath.Base(filename),
		LineNum:  line,
		Msg:      v,
	}
	pushJSONByteStream(exolog)
}

func pushJSONByteStream(
	exolog interface{},
) {
	byteStream, err := json.Marshal(exolog)
	if err != nil {
		log.Println("Logger: Unable to marshal the JSON")
		return
	}

	/* When init doesn't happen before logger gets called */
	if logger == nil {
		initConn()
		if logger == nil {
			log.Println("Logger : Handle couldn't be initialised ")
			return
		}
	}

	//Getting a lock here in order to support multi-threading
	var len int
	muLogger.Lock()
	var buf []byte
	buf = append(byteStream, "\n"...)
	len, err = logger.Write(buf)
	muLogger.Unlock()
	if err != nil {
		log.Printf("Got error while logging. Length written. %d Cause: %s", len, err.Error())
	}
}

// InitYMLogger initializes the logger with service specific config
func InitYMLogger(l LoggerConf) error {
	processName = l.ProcessName
	logFileName = l.LogFileName
	consoleLog = l.ConsoleLog
	logSeverity = logSeverity.FromString(l.LogSeverity)
	hostname, _ = os.Hostname()
	processID = os.Getpid()
	return initConn()
}

func initConn() (err error) {
	/* Set the logger on this connection */
	if logFileName == "" && consoleLog == true {
		logger = os.Stdout
	} else {
		logger, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Println("Unable to open the log file", logFileName)
		}
	}
	return err
}
