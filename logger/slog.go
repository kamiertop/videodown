package logger

import (
	"context"
	"log/slog"
	"strings"

	"github.com/rs/zerolog"
)

// Slog returns a slog.Logger that writes through this package's zerolog
// configuration. The level is supplied by the caller so libraries such as
// Wails can filter their own verbose logs before they reach zerolog.
func (l *Logger) Slog(level slog.Level) *slog.Logger {
	return slog.New(&slogHandler{logger: l.Logger, level: level})
}

type slogHandler struct {
	logger zerolog.Logger
	level  slog.Level
	groups []string
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *slogHandler) Handle(_ context.Context, record slog.Record) error {
	event := h.logger.WithLevel(slogLevel(record.Level))
	record.Attrs(func(attr slog.Attr) bool {
		appendSlogAttr(event, h.groups, attr)
		return true
	})
	event.Msg(record.Message)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	logger := h.logger.With()
	for _, attr := range attrs {
		if attr.Equal(slog.Attr{}) {
			continue
		}
		logger = appendSlogContextAttr(logger, h.groups, attr)
	}
	return &slogHandler{logger: logger.Logger(), level: h.level, groups: h.groups}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	groups := append(append([]string(nil), h.groups...), name)
	return &slogHandler{logger: h.logger, level: h.level, groups: groups}
}

func slogLevel(level slog.Level) zerolog.Level {
	switch {
	case level <= slog.LevelDebug:
		return zerolog.DebugLevel
	case level < slog.LevelWarn:
		return zerolog.InfoLevel
	case level < slog.LevelError:
		return zerolog.WarnLevel
	default:
		return zerolog.ErrorLevel
	}
}

func appendSlogAttr(event *zerolog.Event, groups []string, attr slog.Attr) {
	if attr.Equal(slog.Attr{}) {
		return
	}
	event.Any(slogAttrKey(groups, attr.Key), slogValue(attr.Value))
}

func appendSlogContextAttr(ctx zerolog.Context, groups []string, attr slog.Attr) zerolog.Context {
	if attr.Value.Kind() == slog.KindGroup {
		values := make(map[string]any, len(attr.Value.Group()))
		for _, groupAttr := range attr.Value.Group() {
			values[groupAttr.Key] = slogValue(groupAttr.Value)
		}
		return ctx.Interface(slogAttrKey(groups, attr.Key), values)
	}
	return ctx.Interface(slogAttrKey(groups, attr.Key), slogValue(attr.Value))
}

func slogAttrKey(groups []string, key string) string {
	if len(groups) == 0 {
		return key
	}
	return strings.Join(append(append([]string(nil), groups...), key), ".")
}

func slogValue(value slog.Value) any {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindInt64:
		return value.Int64()
	case slog.KindUint64:
		return value.Uint64()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindBool:
		return value.Bool()
	case slog.KindTime:
		return value.Time()
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindAny:
		return value.Any()
	case slog.KindGroup:
		values := make(map[string]any, len(value.Group()))
		for _, attr := range value.Group() {
			values[attr.Key] = slogValue(attr.Value)
		}
		return values
	default:
		return nil
	}
}
