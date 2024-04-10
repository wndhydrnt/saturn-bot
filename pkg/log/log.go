package log

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

var (
	ctx                 = context.Background()
	defaultHclogAdapter = &hclogAdapter{}
	gitLogger           *slog.Logger
)

type hclogAdapter struct {
	level      *slog.LevelVar
	levelHclog hclog.Level
	logger     *slog.Logger
	name       string
}

func (l *hclogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	slogLvl := mapHclogToSlogLevel(level)
	l.logger.Log(context.Background(), slogLvl, msg, args...)
}

func (l *hclogAdapter) Trace(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *hclogAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *hclogAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *hclogAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

func (l *hclogAdapter) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *hclogAdapter) IsTrace() bool {
	return false
}

func (l *hclogAdapter) IsDebug() bool {
	return l.logger.Enabled(ctx, slog.LevelDebug)
}

func (l *hclogAdapter) IsInfo() bool {
	return l.logger.Enabled(ctx, slog.LevelInfo)
}

func (l *hclogAdapter) IsWarn() bool {
	return l.logger.Enabled(ctx, slog.LevelWarn)
}

func (l *hclogAdapter) IsError() bool {
	return l.logger.Enabled(ctx, slog.LevelError)
}

func (l *hclogAdapter) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l *hclogAdapter) With(args ...interface{}) hclog.Logger {
	logger := slog.With(args...)
	return &hclogAdapter{logger: logger}
}

func (l *hclogAdapter) Name() string {
	return l.name
}

func (l *hclogAdapter) Named(name string) hclog.Logger {
	logger := l.logger.WithGroup(l.name + "." + name)
	return &hclogAdapter{logger: logger, name: name}
}

func (l *hclogAdapter) ResetNamed(name string) hclog.Logger {
	logger := l.logger.WithGroup(name)
	return &hclogAdapter{logger: logger, name: name}
}

func (l *hclogAdapter) SetLevel(level hclog.Level) {
	slogLevel := mapHclogToSlogLevel(level)
	l.level.Set(slogLevel)
	l.levelHclog = level
}

func (l *hclogAdapter) GetLevel() hclog.Level {
	return l.levelHclog
}

func (l *hclogAdapter) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(os.Stderr, "", 0)
}

func (l *hclogAdapter) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}

func mapHclogToSlogLevel(in hclog.Level) slog.Level {
	switch in {
	case hclog.Debug:
		return slog.LevelDebug
	case hclog.Error:
		return slog.LevelError
	case hclog.Info:
		return slog.LevelInfo
	case hclog.NoLevel:
		return slog.LevelError
	case hclog.Off:
		return slog.LevelError
	case hclog.Warn:
		return slog.LevelWarn
	case hclog.Trace:
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

func InitLog(format string, level string, levelGit string) {
	lvl := logStringToLevel(level)
	w := os.Stderr
	isTty := isatty.IsTerminal(w.Fd())
	handler := newHandler(format, isTty, lvl, w)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	defaultHclogAdapter.level = lvl
	defaultHclogAdapter.levelHclog = hclog.Info
	defaultHclogAdapter.logger = logger
	defaultHclogAdapter.name = "plugin"
	lvlGit := logStringToLevel(levelGit)
	handlerGit := newHandler(format, isTty, lvlGit, w)
	gitLogger = slog.New(handlerGit)
}

func DefaultHclogAdapter() hclog.Logger {
	return defaultHclogAdapter
}

func GitLogger() *slog.Logger {
	if gitLogger == nil {
		panic("git logger not initialized")
	}

	return gitLogger
}

func logStringToLevel(level string) *slog.LevelVar {
	lvl := new(slog.LevelVar)
	switch strings.ToLower(level) {
	case "debug":
		lvl.Set(slog.LevelDebug)
	case "error":
		lvl.Set(slog.LevelError)
	case "warn":
		lvl.Set(slog.LevelWarn)
	case "info":
		lvl.Set(slog.LevelInfo)
	default:
		slog.Warn("Unknown log level, falling back to info", "unknown", level)
		lvl.Set(slog.LevelInfo)
	}

	return lvl
}

func newHandler(format string, isTty bool, level *slog.LevelVar, w io.Writer) slog.Handler {
	switch format {
	case "console":
		return tint.NewHandler(w, &tint.Options{
			NoColor: !isTty,
			Level:   level,
		})
	case "json":
		return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	default:
		if isTty {
			return tint.NewHandler(w, &tint.Options{
				Level: level,
			})
		} else {
			return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
		}
	}
}
