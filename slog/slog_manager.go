package slog

import (
	"fmt"
	"time"
)

const (
	LOG_VERSION = "slog-v0.1"
	LOG_MAJOR   = 3
	LOG_MINOR   = 0
	LOG_BUILD   = 1
)

type Level int

const (
	INFO Level = iota
	DEBUG
	WARNING
	ERROR
)

var (
	levelStrings = [...]string{"INFO", "DEBG", "WARN", "EROR"}
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

type LogManager struct {
	Name  string
	level Level
}

func init() {
	InitLogWriter()
}

func (this *LogManager) post_log(debuglv Level, is_log_file bool, name string, format string, args ...interface{}) {
	if this.level > debuglv {
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

var _log_manager_default = &LogManager{"default", INFO}

func LogInfo(name string, format string, args ...interface{}) {
	_log_manager_default.post_log(INFO, false, name, format, args...)
}

func LogDebug(name string, format string, args ...interface{}) {
	_log_manager_default.post_log(DEBUG, true, name, format, args...)
}

func LogWarning(name string, format string, args ...interface{}) {
	_log_manager_default.post_log(WARNING, true, name, format, args...)
}

func LogError(name string, format string, args ...interface{}) {
	_log_manager_default.post_log(ERROR, true, name, format, args...)
}

func SetLogPath(path string) {
	_log_writer.SetPath(path)
}

func Close() {
	_log_writer.LogWrite(nil)
	_log_writer.end_wait.Wait()
}
