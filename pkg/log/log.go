package log

import (
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/mattn/go-isatty"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultHclogAdapter = &hclogAdapter{}
	gitLogger           *zap.SugaredLogger
)

var (
	DefaultLogger *zap.SugaredLogger
)

func Log() *zap.SugaredLogger {
	return DefaultLogger
}

func init() {
	// Ensure that a logger is always present
	logger := zap.Must(zap.NewProduction())
	DefaultLogger = logger.Sugar()
}

type hclogAdapter struct {
	levelHclog hclog.Level
	logger     *zap.SugaredLogger
	name       string
}

func (l *hclogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	lvl := mapHclogToLevel(level)
	l.logger.Logw(lvl, msg, args...)
}

func (l *hclogAdapter) Trace(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}

func (l *hclogAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}

func (l *hclogAdapter) Info(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

func (l *hclogAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warnf(msg, args...)
}

func (l *hclogAdapter) Error(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
}

func (l *hclogAdapter) IsTrace() bool {
	return false
}

func (l *hclogAdapter) IsDebug() bool {
	return l.logger.Level() == zapcore.DebugLevel
}

func (l *hclogAdapter) IsInfo() bool {
	return l.logger.Level() == zapcore.InfoLevel
}

func (l *hclogAdapter) IsWarn() bool {
	return l.logger.Level() == zapcore.WarnLevel
}

func (l *hclogAdapter) IsError() bool {
	return l.logger.Level() == zapcore.ErrorLevel
}

func (l *hclogAdapter) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l *hclogAdapter) With(args ...interface{}) hclog.Logger {
	logger := l.logger.With(args...)
	return &hclogAdapter{logger: logger}
}

func (l *hclogAdapter) Name() string {
	return l.name
}

func (l *hclogAdapter) Named(name string) hclog.Logger {
	logger := l.logger.Named(l.name + "." + name)
	return &hclogAdapter{logger: logger, name: name}
}

func (l *hclogAdapter) ResetNamed(name string) hclog.Logger {
	logger := l.logger.Named(name)
	return &hclogAdapter{logger: logger, name: name}
}

func (l *hclogAdapter) SetLevel(level hclog.Level) {}

func (l *hclogAdapter) GetLevel() hclog.Level {
	return l.levelHclog
}

func (l *hclogAdapter) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(os.Stderr, "", 0)
}

func (l *hclogAdapter) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}

func mapHclogToLevel(in hclog.Level) zapcore.Level {
	switch in {
	case hclog.Debug:
		return zap.DebugLevel
	case hclog.Error:
		return zap.ErrorLevel
	case hclog.Info:
		return zap.InfoLevel
	case hclog.NoLevel:
		return zap.WarnLevel
	case hclog.Off:
		return zap.DebugLevel
	case hclog.Warn:
		return zap.WarnLevel
	case hclog.Trace:
		return zap.DebugLevel
	}

	return zap.InfoLevel
}

func InitLog(format config.ConfigurationLogFormat, level config.ConfigurationLogLevel, levelGit config.ConfigurationGitLogLevel) {
	loggerCfg := newConfig(format)
	loggerCfg.Level = zap.NewAtomicLevelAt(logStringToLevel(string(level)))

	logger := zap.Must(loggerCfg.Build())
	DefaultLogger = logger.Sugar()

	defaultHclogAdapter.levelHclog = hclog.Info
	defaultHclogAdapter.logger = DefaultLogger
	defaultHclogAdapter.name = "plugin"
	lvlGit := logStringToLevel(string(levelGit))
	gitLogger = DefaultLogger.
		WithOptions(zap.IncreaseLevel(lvlGit))
}

func DefaultHclogAdapter() hclog.Logger {
	return defaultHclogAdapter
}

func GitLogger() *zap.SugaredLogger {
	if gitLogger == nil {
		panic("git logger not initialized")
	}

	return gitLogger
}

func logStringToLevel(level string) zapcore.Level {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	return lvl
}

func newConfig(format config.ConfigurationLogFormat) zap.Config {
	zapCfg := zap.NewProductionConfig()
	var encoderFormat string
	if format == "auto" {
		if isatty.IsTerminal(os.Stderr.Fd()) {
			encoderFormat = "console"
		} else {
			encoderFormat = "json"
		}
	}

	if encoderFormat != "console" && encoderFormat != "json" {
		encoderFormat = "json"
	}

	if encoderFormat == "console" {
		zapCfg.DisableCaller = true
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if encoderFormat == "json" {
		zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		zapCfg.EncoderConfig.TimeKey = "time"
	}

	zapCfg.Encoding = encoderFormat
	return zapCfg
}

func FieldDryRun(v bool) zap.Field {
	const key = "saturn-bot.dryRun"
	return zap.Bool(key, v)
}

func FieldRepo(name string) zap.Field {
	const key = "saturn-bot.repository"
	return zap.String(key, name)
}

func FieldTask(name string) zap.Field {
	const key = "saturn-bot.task"
	return zap.String(key, name)
}
