package cli

import "github.com/spf13/cobra"

const (
	GroupIDPlatform = "platform"
	GroupIDSettings = "settings"
)

var (
	PlatformGroup = &cobra.Group{
		ID:    GroupIDPlatform,
		Title: "Platform Commands",
	}
	SettingsGroup = &cobra.Group{
		ID:    GroupIDSettings,
		Title: "App Settings",
	}
)
