package utils

import (
	"testing"

	"github.com/kamiertop/videodown/logger"
)

func TestSettings_GetTheme(t *testing.T) {
	settings := NewSettingsWithMemory(logger.New())
	theme, err := settings.GetTheme()
	if err != nil {
		t.Error(err)
	}
	Equal(t, "light", theme)

	if err := settings.SetTheme("dark"); err != nil {
		t.Error(err)
	}
	theme, err = settings.GetTheme()
	if err != nil {
		t.Error(err)
	}
	Equal(t, "dark", theme)
}
