// estimate.go は `misoca estimate` コマンド群を実装します。
package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mtane0412/misoca-cli/internal/misoca"
	"github.com/mtane0412/misoca-cli/internal/output"
)

func estimateTable(estimates []misoca.Estimate) *output.Table {
	table := &output.Table{Header: []string{"ID", "件名", "見積書番号", "有効期限"}}
	for _, e := range estimates {
		id := ""
		if e.ID != nil {
			id = strconv.Itoa(*e.ID)
		}
		table.Rows = append(table.Rows, []string{id, derefString(e.Subject), derefString(e.EstimateNumber), derefString(e.ExpireDate)})
	}
	return table
}

func (a *App) newEstimateCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "estimate", Short: "見積書の操作"}

	var params misoca.ListEstimatesParams
	var contactGroupID, page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "見積書一覧を取得します",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("contact-group-id") {
				params.ContactGroupID = &contactGroupID
			}
			if cmd.Flags().Changed("page") {
				params.Page = &page
			}
			if cmd.Flags().Changed("per-page") {
				params.PerPage = &perPage
			}
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
			estimates, _, err := client.ListEstimates(ctx, params)
			if err != nil {
				return err
			}
			return output.Write(a.Stdout, format, estimates, estimateTable(estimates))
		},
	}
	listCmd.Flags().StringVar(&params.Type, "type", "", "active/archived/trashed/untrashed")
	listCmd.Flags().IntVar(&contactGroupID, "contact-group-id", 0, "取引先のID")
	listCmd.Flags().IntVar(&page, "page", 0, "ページ番号")
	listCmd.Flags().IntVar(&perPage, "per-page", 0, "1ページあたりの件数 (1〜100)")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <id>",
		Short: "見積書を取得します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
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
			estimate, err := client.GetEstimate(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, estimate)
		},
	})

	var filePath, contactID, subject, issueDate, expireDate string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "見積書を作成します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			var req misoca.CreateEstimateRequest
			if filePath != "" {
				if err := readJSONBody(a.Stdin, filePath, &req); err != nil {
					return err
				}
			}
			if contactID != "" {
				id, err := parseID(contactID)
				if err != nil {
					return err
				}
				req.ContactID = id
			}
			if subject != "" {
				req.Subject = subject
			}
			if issueDate != "" {
				req.IssueDate = issueDate
			}
			if expireDate != "" {
				req.ExpireDate = expireDate
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			created, err := client.CreateEstimate(ctx, req)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, created)
		},
	}
	createCmd.Flags().StringVar(&filePath, "file", "", "リクエストボディのJSONファイル（\"-\"で標準入力）")
	createCmd.Flags().StringVar(&contactID, "contact-id", "", "送り先ID（--fileの値を上書きします）")
	createCmd.Flags().StringVar(&subject, "subject", "", "件名（--fileの値を上書きします）")
	createCmd.Flags().StringVar(&issueDate, "issue-date", "", "発行日 YYYY/MM/DD（--fileの値を上書きします）")
	createCmd.Flags().StringVar(&expireDate, "expire-date", "", "有効期限 YYYY/MM/DD（--fileの値を上書きします）")
	cmd.AddCommand(createCmd)

	cmd.AddCommand(a.newEstimateBinaryCmd(flags, "pdf", "見積書のPDFを取得します", func(c *misoca.Client) binaryFetcher { return c.GetEstimatePDF }, "estimate-%d.pdf"))
	cmd.AddCommand(a.newEstimateBinaryCmd(flags, "logo", "見積書のロゴ画像を取得します", func(c *misoca.Client) binaryFetcher { return c.GetEstimateLogo }, "estimate-%d-logo"))
	cmd.AddCommand(a.newEstimateBinaryCmd(flags, "stamp", "見積書の印影画像を取得します", func(c *misoca.Client) binaryFetcher { return c.GetEstimateStamp }, "estimate-%d-stamp"))

	var mailSubject, mailBody, mailMessage string
	var includingSelfToCc bool
	distributeCmd := &cobra.Command{
		Use:   "distribute <id>",
		Short: "見積書をメール送信します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
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
			dist, err := client.DistributeEstimate(ctx, id, misoca.DistributeEstimateRequest{
				MailSubject:       mailSubject,
				MailBody:          mailBody,
				MailMessage:       mailMessage,
				IncludingSelfToCc: includingSelfToCc,
			})
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, dist)
		},
	}
	distributeCmd.Flags().StringVar(&mailSubject, "mail-subject", "", "メール件名")
	distributeCmd.Flags().StringVar(&mailBody, "mail-body", "", "メール本文")
	distributeCmd.Flags().StringVar(&mailMessage, "mail-message", "", "取引先へのメッセージ（本文指定時は無視されます）")
	distributeCmd.Flags().BoolVar(&includingSelfToCc, "cc-self", false, "自分自身にCCを送る")
	cmd.AddCommand(distributeCmd)

	return cmd
}

// binaryFetcher はIDからバイナリデータを取得するクライアントメソッドの型です。
type binaryFetcher func(ctx context.Context, id int) ([]byte, error)

func (a *App) newEstimateBinaryCmd(flags *globalFlags, use, short string, fetcherFor func(*misoca.Client) binaryFetcher, defaultNamePattern string) *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   use + " <id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = fmt.Sprintf(defaultNamePattern, id)
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			data, err := fetcherFor(client)(ctx, id)
			if err != nil {
				return err
			}
			return saveBinary(a.Stdout, outputPath, data)
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "保存先パス（\"-\"で標準出力）")
	return cmd
}
