package logger

func (l *Logger) Warnf(format string, v ...any) {
	l.Logger.Warn().Msgf(format, v...)
}
