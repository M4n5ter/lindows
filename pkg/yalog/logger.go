package yalog

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	text *slog.Logger
	json *slog.Logger
}

func (l *Logger) Debug(msg string, args ...any) {
	r := newRecord(slog.LevelDebug, msg)
	r.Add(args...)
	handle(l, r, slog.LevelDebug)
}

func (l *Logger) Info(msg string, args ...any) {
	r := newRecord(slog.LevelInfo, msg)
	r.Add(args...)
	handle(l, r, slog.LevelInfo)
}

func (l *Logger) Warn(msg string, args ...any) {
	r := newRecord(slog.LevelWarn, msg)
	r.Add(args...)
	handle(l, r, slog.LevelWarn)
}

func (l *Logger) Error(msg string, args ...any) {
	r := newRecord(slog.LevelError, msg)
	r.Add(args...)
	handle(l, r, slog.LevelError)
}

func (l *Logger) Panic(msg string, args ...any) {
	r := newRecord(slog.LevelError, msg)
	r.Add(args...)
	handle(l, r, slog.LevelError)
	os.Exit(1)
}

func (l *Logger) Debugf(format string, args ...any) {
	r := newRecord(slog.LevelDebug, format, args...)
	handle(l, r, slog.LevelDebug)
}

func (l *Logger) Infof(format string, args ...any) {
	r := newRecord(slog.LevelInfo, format, args...)
	handle(l, r, slog.LevelInfo)
}

func (l *Logger) Warnf(format string, args ...any) {
	r := newRecord(slog.LevelWarn, format, args...)
	handle(l, r, slog.LevelWarn)
}

func (l *Logger) Errorf(format string, args ...any) {
	r := newRecord(slog.LevelError, format, args...)
	handle(l, r, slog.LevelError)
}

func (l *Logger) Panicf(format string, args ...any) {
	r := newRecord(slog.LevelError, format, args...)
	handle(l, r, slog.LevelError)
	os.Exit(1)
}

// With works like `slog.Logger.With`.
//
// It returns a new `Logger` whatever the inner Logger is enabled or not, if both text and json are nil, it returns the original `Logger`.
func (l *Logger) With(args ...any) *Logger {
	if l.text == nil && l.json == nil {
		return l
	}

	var text, json *slog.Logger
	if l.text != nil {
		text = l.text.With(args...)
	}
	if l.json != nil {
		json = l.json.With(args...)
	}
	return &Logger{text: text, json: json}
}

// WithGroup works like `slog.Logger.WithGroup`
//
// It returns a new `Logger` whatever the inner Logger is enabled or not, if both text and json are nil, it returns the original `Logger`.
func (l *Logger) WithGroup(name string) *Logger {
	if l.text == nil && l.json == nil {
		return l
	}

	var text, json *slog.Logger
	if l.text != nil {
		text = l.text.WithGroup(name)
	}
	if l.json != nil {
		json = l.json.WithGroup(name)
	}
	return &Logger{text: text, json: json}
}

// Log works like slog.Logger.Log
func (l *Logger) Log(ctx context.Context, level Level, msg string, args ...any) {
	lv := slog.Level(level)
	if ctx == nil {
		ctx = context.Background()
	}

	r := newRecord(lv, msg)
	r.Add(args...)
	if textEnabled && l.text.Enabled(ctx, lv) {
		_ = l.text.Handler().Handle(ctx, r)
	}

	if jsonEnabled && l.json.Enabled(ctx, lv) {
		_ = l.json.Handler().Handle(ctx, r)
	}
}

// LogAttrs works like slog.Logger.LogAttrs
func (l *Logger) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	lv := slog.Level(level)
	if ctx == nil {
		ctx = context.Background()
	}

	r := newRecord(lv, msg)
	r.AddAttrs(attrs...)
	if textEnabled && l.text.Enabled(ctx, lv) {
		_ = l.text.Handler().Handle(ctx, r)
	}

	if jsonEnabled && l.json.Enabled(ctx, lv) {
		_ = l.json.Handler().Handle(ctx, r)
	}
}
