package slog

import (
	"fmt"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

var (
	levelStrings = [...]string{"DEBG", "INFO", "WARN", "EROR"}
)

const LogBufferLength = 65535

func (l Level) String() string {
	if l < 0 || int(l) > len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
}

type LogRecord struct {
	Level   Level
	Created time.Time
	logfile bool
	Name    string
	Message string
}

var LOG_LEVEL Level = DEBUG

func init() {
	InitLogWriter()
}

func post_log(debuglv Level, is_log_file bool, name string, format string, args ...interface{}) {
	if LOG_LEVEL > debuglv {
		return
	}

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	rec := &LogRecord{
		Level:   debuglv,
		logfile: is_log_file,
		Created: time.Now(),
		Name:    name,
		Message: msg,
	}

	_log_writer.LogWrite(rec)
}

func LogDebug(name string, format string, args ...interface{}) {
	post_log(DEBUG, true, name, format, args...)
}

func LogInfo(name string, format string, args ...interface{}) {
	post_log(INFO, false, name, format, args...)
}

func LogWarning(name string, format string, args ...interface{}) {
	post_log(WARNING, true, name, format, args...)
}

func LogError(name string, format string, args ...interface{}) {
	post_log(ERROR, true, name, format, args...)
}

func SetLogPath(path string) {
	_log_writer.SetPath(path)
}

func Close() {
	_log_writer.LogWrite(nil)
	_log_writer.end_wait.Wait()
}
