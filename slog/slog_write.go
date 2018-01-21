package slog

import (
	"os"
	"sync"
)

type FileWriter struct {
	// The opened file
	filename string
	file     *os.File
}

type LogWriter struct {
	rec      chan *LogRecord
	rot      chan bool
	buf      []byte
	filter   map[string]*FileWriter
	end_wait sync.WaitGroup
}

func (this *LogWriter) LogWrite(rec *LogRecord) {
	this.rec <- rec
}

var _log_writer = &LogWriter{
	rec:    make(chan *LogRecord, LogBufferLength),
	rot:    make(chan bool),
	filter: make(map[string]*FileWriter),
	buf:    make([]byte, 2048),
}

func WriteRecover(log *LogRecord) {
	if e := recover(); e != nil {

	}
}

func (this *LogWriter) writeFile(log *LogRecord) {
	defer WriteRecover(log)

	file_writer, ok := _log_writer.filter[log.Name]
	if ok == false {
		fd, err := os.OpenFile(log.Name+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err == nil {
			file_writer = &FileWriter{log.Name, fd}
		}
		_log_writer.filter[log.Name] = file_writer
	}

	file_writer.file.Write(_log_writer.buf)
}

func (this *LogWriter) writeStd(log *LogRecord) {
	os.Stdout.Write(_log_writer.buf)
}

func InitLogWriter() {
	go run()
}

func run() {
	defer defer_write_close()
	_log_writer.end_wait.Add(1)

	for {

		log, ok := <-_log_writer.rec
		if ok == false || log == nil {
			goto wait_close
		}

		_log_writer.buf = _log_writer.buf[:0]
		//_log_writer.buf = append(_log_writer.buf, "\x1b[031m"...)
		formatHeader(&_log_writer.buf, log.Level, log.Created)
		_log_writer.buf = append(_log_writer.buf, log.Message...)
		//_log_writer.buf = append(_log_writer.buf, "\x1b[0m"...)
		_log_writer.buf = append(_log_writer.buf, '\n')

		if log.logfile == true {
			_log_writer.writeFile(log)
		}

		_log_writer.writeStd(log)

	}
wait_close:
	_log_writer.end_wait.Done()
}

func defer_write_close() {
	for _, filter := range _log_writer.filter {
		if filter.file != nil {
			filter.file.Sync()
			filter.file.Close()
		}
	}
}
