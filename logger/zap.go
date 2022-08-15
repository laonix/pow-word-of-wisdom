package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a contract to level-based logging.
type Logger interface {
	Debug(msg string, kvs ...any)
	Info(msg string, kvs ...any)
	Warn(msg string, kvs ...any)
	Error(err error, kvs ...any)
}

// Level is a logging Level.
type Level int

const (
	// LevelError is an error logging Level.
	LevelError Level = iota
	// LevelWarn is a warning logging Level.
	LevelWarn Level = iota
	// LevelInfo is an info logging Level.
	LevelInfo Level = iota
	// LevelDebug is a debug logging Level.
	LevelDebug Level = iota
)

// LevelOf returns a Level corresponding to an argument string.
func LevelOf(level string) Level {
	tmp := strings.ToLower(level)

	switch tmp {
	case "error":
		return LevelError
	case "warn":
		return LevelWarn
	case "debug":
		return LevelDebug
	default:
		return LevelInfo
	}
}

// ZapLogger is an implementation of Logger wrapping zap.SugaredLogger.
type ZapLogger struct {
	zap *zap.SugaredLogger
}

// NewZapLogger returns new NewZapLogger instance.
func NewZapLogger(level Level) *ZapLogger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zapLevel(level)
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.CallerKey = zapcore.OmitKey
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return &ZapLogger{
		zap: logger.Sugar(),
	}
}

// Info logs a message with some additional context.
func (z ZapLogger) Info(msg string, kvs ...interface{}) {
	z.zap.Infow(msg, kvs...)
}

// Error logs a message with some additional context.
func (z ZapLogger) Error(err error, kvs ...interface{}) {
	if caller, ok := err.(interface{ Caller() string }); ok {
		kvs = append(kvs, zap.String("caller", caller.Caller()))
	}
	if kver, ok := err.(interface{ KeyValues() []interface{} }); ok {
		kvs = append(kvs, kver.KeyValues()...)
	}
	z.zap.Errorw(err.Error(), kvs...)
}

// Warn logs a message with some additional context.
func (z ZapLogger) Warn(msg string, kvs ...interface{}) {
	z.zap.Warnw(msg, kvs...)
}

// Debug logs a message with some additional context.
func (z ZapLogger) Debug(msg string, kvs ...interface{}) {
	z.zap.Debugw(msg, kvs...)
}

func zapLevel(level Level) zap.AtomicLevel {
	al := zap.NewAtomicLevel()
	switch level {
	case LevelDebug:
		al.SetLevel(zap.DebugLevel)
	case LevelInfo:
		al.SetLevel(zap.InfoLevel)
	case LevelError:
		al.SetLevel(zap.ErrorLevel)
	case LevelWarn:
		al.SetLevel(zap.WarnLevel)
	}
	return al
}
