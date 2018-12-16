package slog

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type FileWriter struct {
	// The opened file
	filename string
	file     *os.File
}

type LogWriter struct {
	rec       chan *LogRecord
	init      chan string
	init_path string
	b_init    int32
	buf       []byte
	filter    map[string]*FileWriter
	end_wait  sync.WaitGroup
	path      string
	dayTimer  *time.Timer
}

func (this *LogWriter) LogWrite(rec *LogRecord) {
	this.rec <- rec
}

var _log_writer = &LogWriter{
	rec:    make(chan *LogRecord, LogBufferLength),
	init:   make(chan string),
	filter: make(map[string]*FileWriter),
	buf:    make([]byte, 2048),
	path:   "",
	b_init: 0,
}

func (this *LogWriter) writeFile(log *LogRecord) {
	file_writer, ok := _log_writer.filter[log.Name]
	if ok == false {
		path := path.Join(this.path, log.Name+".log")
		fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			return
		}
		file_writer = &FileWriter{log.Name, fd}
		_log_writer.filter[log.Name] = file_writer
	}

	file_writer.file.Write(_log_writer.buf)
}

func (this *LogWriter) writeStd(log *LogRecord) {
	os.Stdout.Write(_log_writer.buf)
}

func (this *LogWriter) InitPath(p string) {
	if atomic.CompareAndSwapInt32(&this.b_init, 0, 1) == true {
		this.init <- p
	}
}

func (this *LogWriter) MakePath() {
	ok := false
	if file, err := os.Stat(this.path); err != nil {
		ok = os.IsExist(err)
	} else {
		ok = file.IsDir()
	}

	if ok == false {
		os.MkdirAll(this.path, 0x777)
	}
}

func (this *LogWriter) MakeLogTimePath() {
	now := time.Now()
	dirName := fmt.Sprintf("%d-%02d-%02d",
		now.Year(),
		now.Month(),
		now.Day())

	this.path = path.Join(this.init_path, dirName)

	this.MakePath()
	zero_time := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	delayTime := 24*60*60 - (now.Unix() - zero_time.Unix())

	if this.dayTimer == nil {
		this.dayTimer = time.NewTimer(time.Duration(delayTime) * time.Second)
	} else {
		this.dayTimer.Reset(time.Duration(delayTime) * time.Second)
	}
	this.syncLog()
	this.filter = make(map[string]*FileWriter)
}

func (this *LogWriter) Run(b bool) {
	defer this.Recover()
	this.end_wait.Add(1)
	defer this.end_wait.Done()
	if b == true {
		this.init_path = <-this.init
	}
	this.MakeLogTimePath()
	this.doRun()
}

func (this *LogWriter) doRun() {
	var logRecord *LogRecord
	var ok bool
	for {
		select {
		case logRecord, ok = <-this.rec:
		case <-this.dayTimer.C:
			this.MakeLogTimePath()
			continue
		}
		if ok == false || logRecord == nil {
			goto wait_close
		}
		this.buf = this.buf[:0]
		//_log_writer.buf = append(_log_writer.buf, "\x1b[031m"...)
		formatHeader(&this.buf, logRecord.Level, logRecord.Created)
		this.buf = append(this.buf, logRecord.Message...)
		//_log_writer.buf = append(_log_writer.buf, "\x1b[0m"...)
		this.buf = append(this.buf, "\r\n"...)

		if logRecord.logfile == true {
			this.writeFile(logRecord)
		}

		if logRecord.Level == INFO || DebugLevel() == DEBUG {
			this.writeStd(logRecord)
		}

		logRecord.Reset()
		log_pool.Put(logRecord)
	}
wait_close:
	this.DoClose()
}

func (this *LogWriter) Recover() {
	if r := recover(); r != nil {
		LogError("log_manager", "log worker recover[%v]", r)
		debug.PrintStack()
		go this.Run(false)
	}
}

func (this *LogWriter) syncLog() {
	for _, filter := range this.filter {
		filter.file.Sync()
		filter.file.Close()
	}
}

func (this *LogWriter) DoClose() {
	this.syncLog()
}
