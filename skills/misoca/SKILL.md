---
name: misoca
description: Misoca API v3（請求書・見積書・納品書・取引先）をmisoca CLIから操作する。請求書の確認・作成・入金/請求ステータス更新、見積書・納品書のPDF取得などMisocaに関する依頼で使用する。
---

# misoca CLI

Misoca（弥生の請求書作成サービス）の API v3 を操作する非公式CLIです。
`misoca <resource> <action>` の形式で、請求書・見積書・納品書・取引先・送り先・
品目・ユーザー情報を操作できます。

出力は既定で常にJSONです（`--format table` / `--format csv` で切り替え可能）。
JSON出力は `jq` と組み合わせて絞り込むことを想定しています。

## 前提: 認証確認

コマンドを実行する前に、必ず認証状態を確認してください。

```bash
misoca auth status
```

`ErrProfileNotFound`（終了コード3）が返る場合は未ログインです。ユーザーに
`misoca auth login` の実行を促してください（ブラウザでの対話操作が必要なため、
Claude Codeが代わりにログインを完了させることはできません）。

`--profile <名前>` で複数アカウント（個人・法人など）を切り替えられます。省略時は `default`。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 成功 |
| 1 | 一般的な失敗（ネットワーク・ファイルI/Oなど） |
| 2 | 使い方の誤り（不正なフラグ値・不正なID指定など） |
| 3 | 未認証（`misoca auth login` が必要） |
| 4 | Misoca APIがエラーを返した（標準エラー出力に理由が出力される） |

終了コードで分岐すれば、ユーザーへの案内文を出し分けられます。

## よく使うコマンド

```bash
# 未入金の請求書一覧（JSON）
misoca invoice list --payment-status unpaid

# 未入金の請求書一覧（人間向けの表）
misoca invoice list --payment-status unpaid --format table

# 件名にキーワードを含む請求書を検索
misoca invoice list --condition "9月分"

# 特定の請求書を取得
misoca invoice get 12345

# 請求書PDFの取得（保存先省略時は invoice-{id}.pdf）
misoca invoice pdf 12345 -o /tmp/invoice-12345.pdf

# 取引先一覧
misoca contact-group list --format table

# 送り先一覧（特定の取引先に紐づくもの）
misoca contact list --contact-group-id 10

# 品目一覧
misoca item list
```

### jqでの絞り込み例

```bash
# 未入金の請求書のIDと件名だけを抽出
misoca invoice list --payment-status unpaid | jq '.[] | {id, subject}'

# 合計金額の高い順に並べ替え
misoca invoice list | jq 'sort_by(-.body.total_amount_including_tax)'
```

## 請求書の作成

`invoice create` は `--file` でJSONファイル（または `-` で標準入力）から
リクエストボディを渡すのが基本です。頻出フィールドは個別フラグでも指定でき、
`--file` と併用した場合はフラグの値が優先されます。

```bash
cat <<'EOF' > /tmp/invoice.json
{
  "contact_id": 123,
  "subject": "9月分システム開発費",
  "issue_date": "2026/09/30",
  "payment_due_on": "2026/10/31",
  "body": {
    "notes": "いつもお世話になっております。"
  },
  "items": [
    {
      "name": "システム開発費",
      "quantity": 1,
      "unit_price": 500000,
      "unit_name": "式",
      "tax_type": "STANDARD_TAX_10"
    }
  ]
}
EOF

misoca invoice create --file /tmp/invoice.json
```

標準入力から渡す場合:

```bash
echo '{"contact_id":123,"subject":"9月分システム開発費"}' | misoca invoice create --file -
```

作成した請求書はデフォルトでは「未処理・未請求」の下書き状態です。確定して
取引先に送る際は、Misoca側の運用に合わせて `misoca invoice submit <id>` を
実行してください（Web UI側からの請求書送付を別途行っている場合は不要です）。

## ステータス更新

```bash
# 請求済にする / 未請求に戻す
misoca invoice submit 12345
misoca invoice unsubmit 12345

# 入金済にする（入金日省略時はお支払い期限が使われる） / 未入金に戻す
misoca invoice pay 12345 --paid-on 2026/10/31
misoca invoice unpay 12345

# ごみ箱に移動する / 復元する
misoca invoice trash 12345
misoca invoice untrash 12345
```

## 課金を伴う操作への注意

`misoca invoice postal-mail <id>`（請求書の郵送指示）はMisoca側で**課金が
発生する可能性がある**操作です。実行前に確認プロンプトが表示され、`y` の
入力が必要です。Claude Codeが自動実行する場合は、必ず事前にユーザーへ
実行内容と対象の請求書IDを提示し、明示的な承認を得てから `--yes` を付けて
実行してください。ユーザーの確認なしに `--yes` を使ってはいけません。

## ページネーション

一覧系コマンド（`invoice list` / `estimate list` / `delivery-slip list` /
`item list`）は既定で1ページ（最大100件）のみを返します。すべて取得したい
場合は `--all` を付けてください（Link ヘッダを自動で辿ります）。

```bash
misoca invoice list --all --per-page 100
```

## 見積書・納品書

```bash
misoca estimate list --format table
misoca estimate get 500
misoca estimate pdf 500 -o /tmp/estimate-500.pdf
misoca estimate distribute 500 --mail-subject "お見積りのご案内"

misoca delivery-slip list --format table
misoca delivery-slip pdf 800 -o /tmp/delivery-800.pdf
```

## 仕様変更の確認

Misoca APIは「現状有姿・予告なく変更される可能性がある」と明言されています。
挙動が想定と異なる場合は、以下でリポジトリ同梱の仕様スナップショットとの
差分を確認してください。

```bash
misoca spec check
```

差分がある場合、`api/swagger.json` を最新のライブ仕様で更新し、
`go run ./internal/codegen` でモデルを再生成する必要があります。
