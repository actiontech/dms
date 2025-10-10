package log

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	gormLog "gorm.io/gorm/logger"
)

type gogmLogWrapper struct {
	logger log.Logger
	msgKey string
}

func NewGogmLogWrapper(logger log.Logger) *gogmLogWrapper {
	h := &gogmLogWrapper{
		msgKey: "msg", // default message key
		logger: log.With(logger, "module", "gogm"),
	}
	return h
}

// Debug logs a message at debug level.
func (h *gogmLogWrapper) Debug(a string) {
	_ = h.logger.Log(log.LevelDebug, h.msgKey, a)
}

// Debugf logs a message at debug level.
func (h *gogmLogWrapper) Debugf(format string, a ...interface{}) {
	_ = h.logger.Log(log.LevelDebug, h.msgKey, fmt.Sprintf(format, a...))
}

// Info logs a message at info level.
func (h *gogmLogWrapper) Info(a string) {
	_ = h.logger.Log(log.LevelInfo, h.msgKey, fmt.Sprint(a))
}

// Infof logs a message at info level.
func (h *gogmLogWrapper) Infof(format string, a ...interface{}) {
	_ = h.logger.Log(log.LevelInfo, h.msgKey, fmt.Sprintf(format, a...))
}

// Warn logs a message at warn level.
func (h *gogmLogWrapper) Warn(a string) {
	_ = h.logger.Log(log.LevelWarn, h.msgKey, fmt.Sprint(a))
}

// Warnf logs a message at warnf level.
func (h *gogmLogWrapper) Warnf(format string, a ...interface{}) {
	_ = h.logger.Log(log.LevelWarn, h.msgKey, fmt.Sprintf(format, a...))
}

// Error logs a message at error level.
func (h *gogmLogWrapper) Error(a string) {
	_ = h.logger.Log(log.LevelError, h.msgKey, fmt.Sprint(a))
}

// Errorf logs a message at error level.
func (h *gogmLogWrapper) Errorf(format string, a ...interface{}) {
	_ = h.logger.Log(log.LevelError, h.msgKey, fmt.Sprintf(format, a...))
}

// Fatal logs a message at Fatal level.
func (h *gogmLogWrapper) Fatal(a string) {
	_ = h.logger.Log(log.LevelFatal, h.msgKey, fmt.Sprint(a))
}

// Fatalf logs a message at Fatal level.
func (h *gogmLogWrapper) Fatalf(format string, a ...interface{}) {
	_ = h.logger.Log(log.LevelFatal, h.msgKey, fmt.Sprintf(format, a...))
}

type BoltLogger interface {
	LogClientMessage(context string, msg string, args ...interface{})
	LogServerMessage(context string, msg string, args ...interface{})
}

type BoltLogWrapper struct {
	logger *log.Helper
	msgKey string
}

func NewBoltLogWrapper(logger *log.Helper) *BoltLogWrapper {
	h := &BoltLogWrapper{
		msgKey: "msg", // default message key
		logger: logger,
	}
	return h
}

func (bl *BoltLogWrapper) LogClientMessage(id, msg string, args ...interface{}) {
	bl.logBoltMessage("C", id, msg, args)
}

func (bl *BoltLogWrapper) LogServerMessage(id, msg string, args ...interface{}) {
	bl.logBoltMessage("S", id, msg, args)
}

func (bl *BoltLogWrapper) logBoltMessage(src, id string, msg string, args []interface{}) {
	bl.logger.Log(log.LevelDebug, bl.msgKey, fmt.Sprintf("BOLT %s%s: %s", formatId(id), src, fmt.Sprintf(msg, args...)))
}

func formatId(id string) string {
	if id == "" {
		return ""
	}
	return fmt.Sprintf("[%s] ", id)
}

func NewUtilLogWrapper(logger log.Logger) Logger {
	return &UtilLogWrapper{logger: logger}
}

type UtilLogWrapper struct {
	logger log.Logger
}

func (l *UtilLogWrapper) Log(level Level, keyvals ...interface{}) error {
	var myLevel log.Level
	switch level {
	case LevelDebug:
		myLevel = log.LevelDebug
	case LevelInfo:
		myLevel = log.LevelInfo
	case LevelWarn:
		myLevel = log.LevelWarn
	case LevelError:
		myLevel = log.LevelError
	case LevelFatal:
		myLevel = log.LevelFatal
	case LevelInfoDilute:
		myLevel = log.LevelDebug
	default:
		myLevel = log.LevelDebug
	}
	return l.logger.Log(myLevel, keyvals...)
}

type gormLogWrapper struct {
	logger   log.Logger
	msgKey   string
	logLevel gormLog.LogLevel
}

func NewGormLogWrapper(logger log.Logger, level gormLog.LogLevel) *gormLogWrapper {
	h := &gormLogWrapper{
		msgKey:   "msg", // default message key
		logger:   log.With(logger, "module", "gorm"),
		logLevel: level,
	}
	return h
}

func (h *gormLogWrapper) LogMode(level gormLog.LogLevel) gormLog.Interface {
	h.logLevel = level
	return h
}

func (h *gormLogWrapper) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 只有在 Info 级别或更高时才输出 SQL trace 日志
	if h.logLevel >= gormLog.Info {
		elapsed := time.Since(begin)
		sql, rowsAffected := fc()
		_ = h.logger.Log(log.LevelDebug, h.msgKey, fmt.Sprintf("trace: begin:%v; elapsed:%v; sql: %v; rowsAffected: %v; err: %v", begin.Format(LogTimeLayout), elapsed, sql, rowsAffected, err))
	}
}

func (h *gormLogWrapper) Error(ctx context.Context, format string, a ...interface{}) {
	if h.logLevel >= gormLog.Error {
		_ = h.logger.Log(log.LevelError, h.msgKey, fmt.Sprintf(format, a...))
	}
}

func (h *gormLogWrapper) Warn(ctx context.Context, format string, a ...interface{}) {
	if h.logLevel >= gormLog.Warn {
		_ = h.logger.Log(log.LevelWarn, h.msgKey, fmt.Sprintf(format, a...))
	}
}

func (h *gormLogWrapper) Info(ctx context.Context, format string, a ...interface{}) {
	if h.logLevel >= gormLog.Info {
		_ = h.logger.Log(log.LevelInfo, h.msgKey, fmt.Sprintf(format, a...))
	}
}

type kWrapper struct {
	logger Logger
}

func NewKLogWrapper(logger Logger) *kWrapper {
	return &kWrapper{
		logger: logger,
	}
}

func (k *kWrapper) Log(level log.Level, keyvals ...interface{}) error {
	var l Level
	switch level {
	case log.LevelDebug:
		l = LevelDebug
	case log.LevelInfo:
		l = LevelInfo
	case log.LevelWarn:
		l = LevelWarn
	case log.LevelError:
		l = LevelError
	case log.LevelFatal:
		l = LevelFatal
	default:
		l = LevelDebug
	}
	return k.logger.Log(l, keyvals...)
}
