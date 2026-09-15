// user.go は `misoca user` コマンド群を実装します。
package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) newUserCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "ユーザー情報の取得",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "me",
		Short: "認証中のユーザー自身の情報を取得します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()

			user, err := client.GetMe(ctx)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, user)
		},
	})

	return cmd
}
