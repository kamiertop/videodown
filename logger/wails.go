package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	zerolog.Logger
}

const (
	prodEnvKey   = "mode"
	prodEnvValue = "prod"
)

func New() *Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	mode := os.Getenv(prodEnvKey)
	if mode == prodEnvValue {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	return &Logger{
		zerolog.New(logWriter(mode)).With().Timestamp().Logger(),
	}
}

func (l *Logger) WithName(name string) *Logger {
	return &Logger{
		l.Logger.With().Str("name", name).Logger(),
	}
}

func logWriter(mode string) io.Writer {
	if mode == prodEnvValue {
		return &lumberjack.Logger{
			Filename:   "videodown.log",
			MaxSize:    10, // MB
			MaxAge:     0,
			MaxBackups: 100,
			Compress:   true,
		}
	}

	return os.Stdout
}

func (l *Logger) WithCaller(skipFrameCount int) *Logger {
	return &Logger{
		l.Logger.With().CallerWithSkipFrameCount(skipFrameCount).Logger(),
	}
}

func (l *Logger) Print(message string) {
	l.Logger.Print(message)
}

func (l *Logger) Trace(message string) {
	l.Logger.Trace().Msg(message)
}

func (l *Logger) Debug(message string) {
	l.Logger.Debug().Msg(message)
}

func (l *Logger) Info(message string) {
	l.Logger.Info().Msg(message)
}

func (l *Logger) Warning(message string) {
	l.Logger.Warn().Msg(message)
}

func (l *Logger) Error(message string) {
	l.Logger.Error().Msg(message)
}

func (l *Logger) Fatal(message string) {
	l.Logger.Fatal().Msg(message)
}
