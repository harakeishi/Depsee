package logger

import (
	"io"
	"log/slog"
	"os"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type Config struct {
	Level  LogLevel
	Format string // "json" or "text"
	Output io.Writer
}

var defaultLogger *slog.Logger

// Init はログ設定を初期化する
func Init(config Config) {
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(config.Output, opts)
	} else {
		handler = slog.NewTextHandler(config.Output, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// GetLogger はデフォルトロガーを返す
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		// デフォルト設定で初期化
		Init(Config{
			Level:  LevelInfo,
			Format: "text",
			Output: os.Stderr,
		})
	}
	return defaultLogger
}

// Debug はデバッグレベルのログを出力
func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

// Info は情報レベルのログを出力
func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

// Warn は警告レベルのログを出力
func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

// Error はエラーレベルのログを出力
func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

// With は追加のコンテキストを持つロガーを返す
func With(args ...any) *slog.Logger {
	return GetLogger().With(args...)
}
