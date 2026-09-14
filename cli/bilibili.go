package cli

import "github.com/spf13/cobra"

func Bilibili() *cobra.Command {
	return &cobra.Command{
		Use:     "bilibili",
		Aliases: []string{"bili"},
		Short:   "Bilibili downloader",
		Long:    "Bilibili downloader is a command line tool to download videos from Bilibili.",
		GroupID: GroupIDPlatform,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}
