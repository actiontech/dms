package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

type myLogger struct {
	log  *log.Logger
	pool *sync.Pool
}

// NewMyLogger new a logger with writer.
func NewMyLogger(w io.Writer) *myLogger {
	return &myLogger{
		log: log.New(w, "", 0),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Log print the kv pairs log.
func (l *myLogger) Log(level Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if (len(keyvals) & 1) == 1 {
		keyvals = append(keyvals, "KEYVALS UNPAIRED")
	}
	buf := l.pool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.WriteString(level.String())
	for i := 0; i < len(keyvals); i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", keyvals[i], keyvals[i+1])
	}
	_ = l.log.Output(4, buf.String())
	buf.Reset()
	l.pool.Put(buf)
	return nil
}

var _ Logger = (*myLogger)(nil)
