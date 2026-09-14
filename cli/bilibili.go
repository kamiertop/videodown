package cli

import (
	"github.com/kamiertop/videodown/cli/bilibili"
	"github.com/spf13/cobra"
)

func Bilibili() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bilibili",
		Aliases: []string{"bili"},
		Short:   "Bilibili downloader",
		Long:    "Bilibili downloader is a command line tool to download videos from Bilibili.",
		GroupID: GroupIDPlatform,
	}
	var c bilibili.Core

	cmd.AddCommand(c.Login())

	return cmd
}
