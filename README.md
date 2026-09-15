# misoca-cli

[Misoca API v3](https://doc.misoca.jp/) を操作するための非公式 CLI です。

Claude Code などの CLI 経由でツールを扱うエージェントから、請求書・見積書・納品書・取引先を操作できるようにすることを目的としています。

## インストール

```bash
go install github.com/mtane0412/misoca-cli/cmd/misoca@latest
```

## アプリケーションの登録

Misoca API を利用するには、事前に Misoca にログインした状態で
[アプリケーション管理ページ](https://app.misoca.jp/oauth2/applications) から
アプリケーションを登録する必要があります。

1. 「新しいアプリケーション」を押下します。
2. 名称に任意の名前（例: `misoca-cli`）を入力します。
3. コールバックURLに `http://localhost:8765/callback` を入力します。
4. 登録後に表示される「アプリケーションID」と「シークレット」を控えます。

## 使い方

```bash
# ログイン（初回のみ。クライアントID/シークレットを対話的に入力）
misoca auth login

# ログイン状態の確認
misoca auth status

# 自分のユーザー情報
misoca user me

# 請求書一覧（JSON）
misoca invoice list --per-page 10

# 請求書一覧（表形式）
misoca invoice list --payment-status unpaid --format table
```

詳しくは `misoca --help` を参照してください。

## 開発

```bash
go build ./...
go vet ./...
go test ./...
```

Swagger 仕様のスナップショットは `api/swagger.json` にあります。モデル定義は
`internal/codegen` によりこのファイルから生成されます。

```bash
go run ./internal/codegen
```

## ライセンス

MIT
