package bilibili

import (
	"time"

	_ "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

func (c *Core) Login() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "扫描二维码登录哔哩哔哩账号",
		RunE: func(cmd *cobra.Command, args []string) error {
			ticker := time.NewTicker(time.Second)
			for range ticker.C {

			}

			return nil
		},
	}
}
