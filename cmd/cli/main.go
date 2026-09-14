package main

import (
	"fmt"
	"os"

	"github.com/kamiertop/videodown/cli"
	"github.com/spf13/cobra"
)

func main() {
	root := cobra.Command{
		Use:   "videodown-cli",
		Short: "videodown 命令行工具",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	root.AddGroup(cli.PlatformGroup, cli.SettingsGroup)

	root.AddCommand(cli.Douyin())
	root.AddCommand(cli.Bilibili())

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
