package log

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/murlokito/gophercoin/log/internal/json"
)

const (
	defaultBufSize     = 500
	defaultDetailsSize = 5
)

var mu sync.Mutex

type logMessage struct {
	json.Encoder
	buf []byte
}

func (msg *logMessage) writeDetails(details []Detail) []byte {
	var err error
	for i := 0; i < len(details); i++ {
		detail := details[i]
		if detail.key != "" {
			msg.buf = msg.ObjKey(msg.buf, detail.key)
		}

		msg.buf, err = msg.ValueInterface(msg.buf, detail.value)
		if err != nil {
			msg.buf = msg.ValueString(msg.buf, "could not serialize "+detail.key)
		}

		if i < len(details)-1 {
			msg.buf = msg.Comma(msg.buf)
		}
	}

	return msg.buf
}

// Conf configurations to setup the logger
type Conf struct {
	Level  Level
	Writer io.Writer
}

// Log performs the needed logs
type Log struct {
	pool   *sync.Pool
	level  Level
	writer io.Writer
	fields []Detail
	err    error
}

// New returns a new Log
func New(conf Conf) *Log {
	log := initLog(conf.Level, conf.Writer)
	log.pool = &sync.Pool{
		New: func() interface{} {
			return &logMessage{
				buf: make([]byte, 0, defaultBufSize),
			}
		},
	}
	return log
}

func initLog(level Level, writer io.Writer) *Log {
	log := new(Log)
	log.level = level
	log.writer = writer
	log.fields = make([]Detail, 0, defaultDetailsSize)
	return log
}

// WithError stores the error for later logging
func (log *Log) WithError(err error) Logger {
	l := initLog(log.level, log.writer)
	l.pool = log.pool
	l.fields = log.fields
	l.err = err
	return l
}

// WithDetails stores the details for later logging
func (log *Log) WithDetails(details ...Detail) Logger {
	l := initLog(log.level, log.writer)
	l.pool = log.pool
	l.err = log.err
	l.fields = append(l.fields, log.fields...)
	for _, detail := range details {
		switch detail.detailType {
		case field:
			if detail.key == "" {
				continue
			}

			l.fields = append(l.fields, detail)
		}
	}

	return l
}

// Debug prints debug level logs
func (log *Log) Debug(format string, args ...interface{}) {
	log.print(DebugLevel, format, args...)
}

// Info prints info level logs
func (log *Log) Info(format string, args ...interface{}) {
	log.print(InfoLevel, format, args...)
}

// Error prints error level logs
func (log *Log) Error(format string, args ...interface{}) {
	log.print(ErrorLevel, format, args...)
}

func (log *Log) print(level Level, format string, args ...interface{}) {
	if log.level > level {
		return
	}

	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}

	date := time.Now().UTC().Format(time.RFC3339Nano)

	msg := log.pool.Get().(*logMessage)
	msg.buf = make([]byte, 0, defaultBufSize)
	msg.buf = msg.BeginObj(msg.buf)
	msg.buf = msg.ObjKey(msg.buf, "timestamp")
	msg.buf = msg.ValueString(msg.buf, date)
	msg.buf = msg.Comma(msg.buf)

	msg.buf = msg.ObjKey(msg.buf, "msg")
	msg.buf = msg.ValueString(msg.buf, format)

	if log.err != nil || len(log.fields) > 0 {
		msg.buf = msg.Comma(msg.buf)
		msg.buf = msg.ObjKey(msg.buf, "fields")
		msg.buf = msg.BeginObj(msg.buf)
		if log.err != nil {
			msg.buf = msg.ObjKey(msg.buf, "error")
			msg.buf = msg.ValueBytes(msg.buf, []byte(log.err.Error()))
			if len(log.fields) > 0 {
				msg.buf = msg.Comma(msg.buf)
			}

			log.err = nil
		}

		msg.buf = msg.writeDetails(log.fields)
		msg.buf = msg.EndObj(msg.buf)
	}

	msg.buf = msg.EndObj(msg.buf)
	msg.buf = msg.NewLine(msg.buf)
	log.write(msg)
}

func (log *Log) write(msg *logMessage) {
	mu.Lock()
	log.writer.Write(msg.buf) // nolint
	mu.Unlock()
	log.pool.Put(msg)
}
