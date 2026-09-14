package cli

import (
	"github.com/kamiertop/videodown/cli/douyin"
	"github.com/spf13/cobra"
)

func Douyin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "douyin",
		Short: "Douyin downloader",
		Long:  "Douyin downloader is a command line tool to download videos from Douyin.",
		Run: func(cmd *cobra.Command, args []string) {

		},
		GroupID: GroupIDPlatform,
	}
	cmd.AddCommand(douyin.Login())

	return cmd
}
