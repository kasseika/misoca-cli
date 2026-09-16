// delivery_slip.go は `misoca delivery-slip` コマンド群を実装します。
package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kasseika/misoca-cli/internal/misoca"
	"github.com/kasseika/misoca-cli/internal/output"
)

func deliverySlipTable(slips []misoca.DeliverySlip) *output.Table {
	table := &output.Table{Header: []string{"ID", "件名", "納品書番号", "納品日"}}
	for _, s := range slips {
		id := ""
		if s.ID != nil {
			id = strconv.Itoa(*s.ID)
		}
		table.Rows = append(table.Rows, []string{id, derefString(s.Subject), derefString(s.DeliverySlipNumber), derefString(s.DeliveryDate)})
	}
	return table
}

func (a *App) newDeliverySlipCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "delivery-slip", Short: "納品書の操作"}

	var params misoca.ListDeliverySlipsParams
	var contactGroupID, page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "納品書一覧を取得します",
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
			slips, _, err := client.ListDeliverySlips(ctx, params)
			if err != nil {
				return err
			}
			return output.Write(a.Stdout, format, slips, deliverySlipTable(slips))
		},
	}
	listCmd.Flags().StringVar(&params.Type, "type", "", "active/archived/trashed/untrashed")
	listCmd.Flags().IntVar(&contactGroupID, "contact-group-id", 0, "取引先のID")
	listCmd.Flags().IntVar(&page, "page", 0, "ページ番号")
	listCmd.Flags().IntVar(&perPage, "per-page", 0, "1ページあたりの件数 (1〜100)")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <id>",
		Short: "納品書を取得します",
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
			slip, err := client.GetDeliverySlip(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, slip)
		},
	})

	var filePath, contactID, subject, issueDate, deliveryDate string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "納品書を作成します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			var req misoca.CreateDeliverySlipRequest
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
			if deliveryDate != "" {
				req.DeliveryDate = deliveryDate
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			created, err := client.CreateDeliverySlip(ctx, req)
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
	createCmd.Flags().StringVar(&deliveryDate, "delivery-date", "", "納品日 YYYY/MM/DD（--fileの値を上書きします）")
	cmd.AddCommand(createCmd)

	var outputPath string
	pdfCmd := &cobra.Command{
		Use:   "pdf <id>",
		Short: "納品書のPDFを取得します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = fmt.Sprintf("delivery-slip-%d.pdf", id)
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			data, err := client.GetDeliverySlipPDF(ctx, id)
			if err != nil {
				return err
			}
			return saveBinary(a.Stdout, outputPath, data)
		},
	}
	pdfCmd.Flags().StringVarP(&outputPath, "output", "o", "", "保存先パス（省略時は delivery-slip-{id}.pdf、\"-\"で標準出力）")
	cmd.AddCommand(pdfCmd)

	return cmd
}
