package logger

import (
	"log/slog"
	"os"
	"strings"
)

/**
 * New 创建 JSON 格式的结构化日志记录器。
 *
 * @param levelName 日志级别名称
 * @return *slog.Logger 结构化日志记录器
 */
func New(levelName string) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(levelName) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
