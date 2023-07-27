package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	kLog "github.com/go-kratos/kratos/v2/log"
)

var _ kLog.Logger = (*stdLogger)(nil)

type stdLogger struct {
	logTimeLayout string
	log           *log.Logger
	pool          *sync.Pool
}

// NewStdLogger new a logger with writer.
func NewStdLogger(w io.Writer, timeLayout string) kLog.Logger {
	return &stdLogger{
		logTimeLayout: timeLayout,
		log:           log.New(w, "", 0),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Log print the kv pairs log.
func (l *stdLogger) Log(level kLog.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if (len(keyvals) & 1) == 1 {
		keyvals = append(keyvals, "KEYVALS UNPAIRED")
	}
	buf, _ := l.pool.Get().(*bytes.Buffer)
	buf.WriteString(time.Now().Format(l.logTimeLayout))
	buf.WriteString(" ")
	buf.WriteString(level.String())
	for i := 0; i < len(keyvals); i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", keyvals[i], keyvals[i+1])
	}
	_ = l.log.Output(4, buf.String()) //nolint:gomnd
	buf.Reset()
	l.pool.Put(buf)
	return nil
}

func (l *stdLogger) Close() error {
	return nil
}
