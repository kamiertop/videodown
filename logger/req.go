package logger

func (l *Logger) Warnf(format string, v ...any) {
	l.Logger.Warn().Msgf(format, v...)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.Logger.Fatal().Msgf(format, v...)
}
