// invoice.go は `misoca invoice` コマンド群を実装します。
package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kasseika/misoca-cli/internal/misoca"
	"github.com/kasseika/misoca-cli/internal/output"
)

// invoiceStatusLabel は請求ステータス（0/1の整数）を日本語ラベルに変換します。
func invoiceStatusLabel(status *int) string {
	if status == nil {
		return ""
	}
	if *status == 1 {
		return "請求済"
	}
	return "未請求"
}

// paymentStatusLabel は入金ステータス（0/1の整数）を日本語ラベルに変換します。
func paymentStatusLabel(status *int) string {
	if status == nil {
		return ""
	}
	if *status == 1 {
		return "入金済"
	}
	return "未入金"
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func invoiceTable(invoices []misoca.Invoice) *output.Table {
	table := &output.Table{Header: []string{"ID", "件名", "請求ステータス", "入金ステータス", "支払期限"}}
	for _, inv := range invoices {
		id := ""
		if inv.ID != nil {
			id = strconv.Itoa(*inv.ID)
		}
		table.Rows = append(table.Rows, []string{
			id,
			derefString(inv.Subject),
			invoiceStatusLabel(inv.InvoiceStatus),
			paymentStatusLabel(inv.PaymentStatus),
			derefString(inv.PaymentDueOn),
		})
	}
	return table
}

func (a *App) newInvoiceCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoice",
		Short: "請求書の操作",
	}

	cmd.AddCommand(a.newInvoiceListCmd(flags))
	cmd.AddCommand(a.newInvoiceGetCmd(flags))
	cmd.AddCommand(a.newInvoiceCreateCmd(flags))
	cmd.AddCommand(a.newInvoicePdfCmd(flags))
	cmd.AddCommand(a.newInvoiceActionCmd(flags, "submit", "請求書を請求済にします", func(c *misoca.Client) invoiceIDAction { return c.SubmitInvoice }))
	cmd.AddCommand(a.newInvoiceActionCmd(flags, "unsubmit", "請求書を未請求に戻します", func(c *misoca.Client) invoiceIDAction { return c.UnsubmitInvoice }))
	cmd.AddCommand(a.newInvoicePayCmd(flags))
	cmd.AddCommand(a.newInvoiceActionCmd(flags, "unpay", "請求書を未入金に戻します", func(c *misoca.Client) invoiceIDAction { return c.UnpayInvoice }))
	cmd.AddCommand(a.newInvoiceActionCmd(flags, "trash", "請求書をごみ箱に移動します", func(c *misoca.Client) invoiceIDAction { return c.TrashInvoice }))
	cmd.AddCommand(a.newInvoiceActionCmd(flags, "untrash", "ごみ箱の請求書を復元します", func(c *misoca.Client) invoiceIDAction { return c.UntrashInvoice }))
	cmd.AddCommand(a.newInvoicePostalMailCmd(flags))

	return cmd
}

func (a *App) newInvoiceListCmd(flags *globalFlags) *cobra.Command {
	var params misoca.ListInvoicesParams
	var contactGroupID, page, perPage int
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "請求書一覧を取得します",
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

			var result []misoca.Invoice
			for {
				invoices, linkHeader, err := client.ListInvoices(ctx, params)
				if err != nil {
					return err
				}
				result = append(result, invoices...)
				if !all {
					break
				}
				next, ok := nextPage(linkHeader)
				if !ok {
					break
				}
				params.Page = &next
			}

			return output.Write(a.Stdout, format, result, invoiceTable(result))
		},
	}

	cmd.Flags().StringVar(&params.Type, "type", "", "active/archived/trashed/untrashed")
	cmd.Flags().StringVar(&params.From, "from", "", "請求日の開始日 (YYYY/MM/DD)")
	cmd.Flags().StringVar(&params.To, "to", "", "請求日の終了日 (YYYY/MM/DD)")
	cmd.Flags().StringVar(&params.DueDateFrom, "due-date-from", "", "お支払い期限の開始日 (YYYY/MM/DD)")
	cmd.Flags().StringVar(&params.DueDateTo, "due-date-to", "", "お支払い期限の終了日 (YYYY/MM/DD)")
	cmd.Flags().StringVar(&params.UpdatedAtFrom, "updated-at-from", "", "更新日の開始日 (YYYY/MM/DD)")
	cmd.Flags().StringVar(&params.UpdatedAtTo, "updated-at-to", "", "更新日の終了日 (YYYY/MM/DD)")
	cmd.Flags().StringVar(&params.PaymentStatus, "payment-status", "", "paid/unpaid")
	cmd.Flags().StringVar(&params.InvoiceStatus, "invoice-status", "", "submitted/unsubmitted")
	cmd.Flags().IntVar(&contactGroupID, "contact-group-id", 0, "取引先のID")
	cmd.Flags().StringVar(&params.Condition, "condition", "", "請求書番号・取引先名・件名・社内メモの検索キーワード")
	cmd.Flags().StringVar(&params.Order, "order", "", "asc/desc")
	cmd.Flags().StringVar(&params.OrderBy, "order-by", "", "created_at/updated_at/issue_date/payment_due_on")
	cmd.Flags().IntVar(&page, "page", 0, "ページ番号")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "1ページあたりの件数 (1〜100)")
	cmd.Flags().BoolVar(&all, "all", false, "Linkヘッダを辿ってすべてのページを取得します")

	return cmd
}

func (a *App) newInvoiceGetCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "請求書を取得します",
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

			invoice, err := client.GetInvoice(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, invoice)
		},
	}
}

func (a *App) newInvoiceCreateCmd(flags *globalFlags) *cobra.Command {
	var filePath, contactID, subject, issueDate, paymentDueOn string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "請求書を作成します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}

			var req misoca.CreateInvoiceRequest
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
			if paymentDueOn != "" {
				req.PaymentDueOn = paymentDueOn
			}

			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()

			created, err := client.CreateInvoice(ctx, req)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, created)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "リクエストボディのJSONファイル（\"-\"で標準入力）")
	cmd.Flags().StringVar(&contactID, "contact-id", "", "送り先ID（--fileの値を上書きします）")
	cmd.Flags().StringVar(&subject, "subject", "", "件名（--fileの値を上書きします）")
	cmd.Flags().StringVar(&issueDate, "issue-date", "", "請求日 YYYY/MM/DD（--fileの値を上書きします）")
	cmd.Flags().StringVar(&paymentDueOn, "payment-due-on", "", "お支払い期限 YYYY/MM/DD（--fileの値を上書きします）")

	return cmd
}

func (a *App) newInvoicePdfCmd(flags *globalFlags) *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "pdf <id>",
		Short: "請求書のPDFを取得します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = fmt.Sprintf("invoice-%d.pdf", id)
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()

			data, err := client.GetInvoicePDF(ctx, id)
			if err != nil {
				return err
			}
			return saveBinary(a.Stdout, outputPath, data)
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "保存先パス（省略時は invoice-{id}.pdf、\"-\"で標準出力）")
	return cmd
}

// invoiceIDAction はIDのみを引数に取る請求書操作（submit/unsubmit/unpay/trash/untrash）の型です。
type invoiceIDAction func(ctx context.Context, id int) (*misoca.Invoice, error)

func (a *App) newInvoiceActionCmd(flags *globalFlags, use, short string, actionFor func(*misoca.Client) invoiceIDAction) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id>",
		Short: short,
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

			invoice, err := actionFor(client)(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, invoice)
		},
	}
}

func (a *App) newInvoicePayCmd(flags *globalFlags) *cobra.Command {
	var paidOn string
	cmd := &cobra.Command{
		Use:   "pay <id>",
		Short: "請求書を入金済にします",
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

			invoice, err := client.PayInvoice(ctx, id, misoca.MarkInvoicePaidRequest{PaidOn: paidOn})
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, invoice)
		},
	}
	cmd.Flags().StringVar(&paidOn, "paid-on", "", "入金日 YYYY/MM/DD（省略時はお支払い期限が使われます）")
	return cmd
}

func (a *App) newInvoicePostalMailCmd(flags *globalFlags) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "postal-mail <id>",
		Short: "請求書の郵送を指示します（課金が発生する可能性があります）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			if !confirm(a.Stdin, a.Stderr, yes, fmt.Sprintf(
				"請求書(ID=%d)の郵送を指示します。Misoca側で課金が発生する可能性があります。よろしいですか？", id)) {
				return usageErrorf("郵送指示を中断しました")
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

			result, err := client.SendInvoiceByPostalMail(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, result)
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "確認プロンプトを省略します")
	return cmd
}
