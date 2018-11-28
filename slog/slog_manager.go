package slog

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Level uint32

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

var (
	levelStrings = [...]string{"DEBG", "INFO", "WARN", "EROR"}
)

func init() {
	go _log_writer.Run()
}

const LogBufferLength = 65535

func (l Level) String() string {
	if l < 0 || uint32(l) > uint32(len(levelStrings)) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
}

var log_pool *sync.Pool = &sync.Pool{New: func() interface{} { return new(LogRecord) }}

type LogRecord struct {
	Level   Level
	Created time.Time
	logfile bool
	Name    string
	Message string
}

func (this *LogRecord) Reset() {
	this.Name = ""
	this.Message = ""
}

var LOG_LEVEL Level = DEBUG

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

	rec := log_pool.Get().(*LogRecord)

	rec.Level = debuglv
	rec.logfile = is_log_file
	rec.Created = time.Now()
	rec.Name = name
	rec.Message = msg

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

func DebugLevel() Level {
	return Level(atomic.LoadUint32((*uint32)(&LOG_LEVEL)))
}

func SetDebugLevel(level Level) {
	atomic.StoreUint32((*uint32)(&LOG_LEVEL), uint32(level))
}

func SetLogPath(path string) {
	_log_writer.InitPath(path)
}

func Close() {
	_log_writer.LogWrite(nil)
	_log_writer.end_wait.Wait()
}

func Run() {
	SetLogPath("log")
}