package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func New() *Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return &Logger{
		zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(3).Logger(),
	}
}

func (l *Logger) WithName(name string) *Logger {
	return &Logger{
		l.Logger.With().Str("name", name).Logger(),
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
