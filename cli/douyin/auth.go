package douyin

import (
	"errors"

	"github.com/spf13/cobra"
)

func Login() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login <cookie>",
		Short:   "使用抖音Cookie登录账号",
		Example: `  videodown-cli douyin login "enter_pc_once=1; bd_ticket_guard_client_web_domain=2; "`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("缺少必须的参数：请在浏览器中登录抖音查询cookie")
			}
			if len(args) > 1 {
				return errors.New("参数过多：考虑使用双引号包裹")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cookie := args[0]
			_ = cookie

			return nil
		},
	}

	return cmd
}
