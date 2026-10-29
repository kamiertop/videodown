package logger

func (l *Logger) Errorf(format string, v ...any) {
	l.Logger.Error().Msgf(format, v...)
}
func (l *Logger) Warningf(format string, v ...any) {
	l.Logger.Warn().Msgf(format, v...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.Logger.Info().Msgf(format, v...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.Logger.Debug().Msgf(format, v...)
}
