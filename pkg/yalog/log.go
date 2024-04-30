package yalog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	logger   Logger
	levelVar slog.LevelVar

	textEnabled bool
	jsonEnabled bool
)

type Level slog.Level

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

func init() {
	SetTextLogger(os.Stdout, true)
}

func setDefaultSlogHandlerOptions(l *slog.HandlerOptions) {
	l.AddSource = true
	l.Level = &levelVar
}

// EnableTextLogger enables text logger.
func EnableTextLogger() {
	textEnabled = true
}

// EnableJSONLogger enables json logger.
func EnableJSONLogger() {
	jsonEnabled = true
}

// DisableTextLogger disables text logger.
func DisableTextLogger() {
	if !jsonEnabled {
		return
	}
	textEnabled = false
}

// DisableJSONLogger disables json logger.
func DisableJSONLogger() {
	if !textEnabled {
		return
	}
	jsonEnabled = false
}

// Default returns the default logger.
func Default() *Logger {
	return &logger
}

// AddSource adds source to slog handler options.
func AddSource(options *slog.HandlerOptions) {
	options.AddSource = true
	options.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		// Remove the directory from the source's filename.
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
}

// SetTextLogger sets and enables text logger.
func SetTextLogger(writer io.Writer, addSource bool) {
	slogHandlerOptions := &slog.HandlerOptions{}
	setDefaultSlogHandlerOptions(slogHandlerOptions)
	if addSource {
		AddSource(slogHandlerOptions)
	}

	logger.text = slog.New(slog.NewTextHandler(writer, slogHandlerOptions))
	textEnabled = true
}

// SetJSONLogger sets and enables json logger.
func SetJSONLogger(writer io.Writer, addSource bool) {
	slogHandlerOptions := &slog.HandlerOptions{}
	setDefaultSlogHandlerOptions(slogHandlerOptions)
	if addSource {
		AddSource(slogHandlerOptions)
	}

	logger.json = slog.New(slog.NewJSONHandler(writer, slogHandlerOptions))
	jsonEnabled = true
}

// SetLevelDebug sets the default logger's level to debug.
func SetLevelDebug() {
	levelVar.Set(slog.LevelDebug)
}

// SetLevelInfo sets the default logger's level to info.
func SetLevelInfo() {
	levelVar.Set(slog.LevelInfo)
}

// SetLevelWarn sets the default logger's level to warn.
func SetLevelWarn() {
	levelVar.Set(slog.LevelWarn)
}

// SetLevelError sets the default logger's level to error.
func SetLevelError() {
	levelVar.Set(slog.LevelError)
}

// Debug logs a debug message.
//
//	yalog.Debug("hello world")
//	yalog.Debug("hello world", "age", 18, "name", "foo")
func Debug(msg string, args ...any) {
	r := newRecord(slog.LevelDebug, msg)
	r.Add(args...)
	handle(nil, r, slog.LevelDebug)
}

// Info logs an info message.
//
//	yalog.Info("hello world")
//	yalog.Info("hello world", "age", 18, "name", "foo")
func Info(msg any, args ...any) {
	r := newRecord(slog.LevelInfo, fmt.Sprintf("%v", msg))
	r.Add(args...)
	handle(nil, r, slog.LevelInfo)
}

// Warn logs a warn message.
//
// In most cases, you should use `Error` instead of `Warn` because people always ignore warnings.
//
//	yalog.Warn("hello world")
//	yalog.Warn("hello world", "age", 18, "name", "foo")
func Warn(msg any, args ...any) {
	r := newRecord(slog.LevelWarn, fmt.Sprintf("%v", msg))
	r.Add(args...)
	handle(nil, r, slog.LevelWarn)
}

// Error logs an error message.
//
//	yalog.Error("hello world")
//	yalog.Error("hello world", "age", 18, "name", "foo")
func Error(msg any, args ...any) {
	r := newRecord(slog.LevelError, fmt.Sprintf("%v", msg))
	r.Add(args...)
	handle(nil, r, slog.LevelError)
}

// Panic logs an error message and exit the current program with `1` error code.
//
//	yalog.Fatal("hello world")
//	yalog.Fatal("hello world", "age", 18, "name", "foo")
func Fatal(msg any, args ...any) {
	r := newRecord(slog.LevelError, fmt.Sprintf("%v", msg))
	r.Add(args...)
	handle(nil, r, slog.LevelError)
	os.Exit(1)
}

// Debugf logs and formats a debug message. Can't take attributes.
//
//	yalog.Debugf("hello world")
//	yalog.Debugf("hello %s", "world")
func Debugf(format string, args ...any) {
	r := newRecord(slog.LevelDebug, format, args...)
	handle(nil, r, slog.LevelDebug)
}

// Infof logs and formats an info message. Can't take attributes.
//
//	yalog.Infof("hello world")
//	yalog.Infof("hello %s", "world")
func Infof(format string, args ...any) {
	r := newRecord(slog.LevelInfo, format, args...)
	handle(nil, r, slog.LevelInfo)
}

// Warnf logs and formats a warn message. Can't take attributes.
//
//	yalog.Warnf("hello world")
//	yalog.Warnf("hello %s", "world")
func Warnf(format string, args ...any) {
	r := newRecord(slog.LevelWarn, format, args...)
	handle(nil, r, slog.LevelWarn)
}

// Errorf logs and formats an error message. Can't take attributes.
//
//	yalog.Errorf("hello world")
//	yalog.Errorf("hello %s", "world")
func Errorf(format string, args ...any) {
	r := newRecord(slog.LevelError, format, args...)
	handle(nil, r, slog.LevelError)
}

// Fatalf logs and formats an error message and exit the current program with `1` error code. Can't take attributes.
//
//	yalog.Fatalf("hello world")
//	yalog.Fatalf("hello %s", "world")
func Fatalf(format string, args ...any) {
	r := newRecord(slog.LevelError, format, args...)
	handle(nil, r, slog.LevelError)
	os.Exit(1)
}

func newRecord(level slog.Level, format string, args ...any) slog.Record {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [runtime.Callers, this function, this function's caller]
	if args == nil {
		return slog.NewRecord(time.Now(), level, format, pcs[0])
	}
	return slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
}

func handle(l *Logger, r slog.Record, level slog.Level) {
	if l == nil {
		if textEnabled && logger.text.Enabled(context.TODO(), level) {
			_ = logger.text.Handler().Handle(context.Background(), r)
		}

		if jsonEnabled && logger.json.Enabled(context.TODO(), level) {
			_ = logger.json.Handler().Handle(context.Background(), r)
		}
	} else {
		if textEnabled && l.text.Enabled(context.TODO(), level) {
			_ = l.text.Handler().Handle(context.Background(), r)
		}

		if jsonEnabled && l.json.Enabled(context.TODO(), level) {
			_ = l.json.Handler().Handle(context.Background(), r)
		}
	}
}
