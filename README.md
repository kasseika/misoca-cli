# misoca-cli

[Misoca API v3](https://doc.misoca.jp/) を操作するための非公式 CLI です。

Claude Code などの CLI 経由でツールを扱うエージェントから、請求書・見積書・納品書・取引先を操作できるようにすることを目的としています。

## インストール

### go install（開発者向け）

```bash
go install github.com/mtane0412/misoca-cli/cmd/misoca@latest
```

### GitHub Releasesのバイナリ（Goをインストールしていない方向け）

[Releases](https://github.com/kasseika/misoca-cli/releases) から
お使いのOS・CPUアーキテクチャに合ったアーカイブ（例: macOSのApple
Siliconなら `misoca_<version>_darwin_arm64.tar.gz`）をダウンロードし、
展開した `misoca` バイナリをPATHの通ったディレクトリ（例: `/usr/local/bin`）
に配置してください。

```bash
tar xzf misoca_<version>_darwin_arm64.tar.gz
sudo mv misoca /usr/local/bin/
misoca --help
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

## チートシート

```bash
# 未入金の請求書を表で確認
misoca invoice list --payment-status unpaid --format table

# 特定の請求書のPDFをダウンロード
misoca invoice pdf 12345 -o ~/Desktop/invoice-12345.pdf

# 請求書を入金済にする（入金日を指定）
misoca invoice pay 12345 --paid-on 2026/10/31

# 取引先の一覧
misoca contact-group list --format table

# 見積書のPDFをダウンロード
misoca estimate pdf 500 -o ~/Desktop/estimate-500.pdf
```

出力形式は既定でJSONです。ターミナルで人が読む場合は `--format table` を、
Excel等に貼り付ける場合は `--format csv` を付けてください。

複数のMisocaアカウント（個人・法人など）を使い分ける場合は、ログイン時・
利用時の両方で `--profile <名前>` を指定してください（省略時は `default`）。

```bash
misoca auth login --profile work
misoca invoice list --profile work --format table
```

## Claude Codeとの連携

`skills/misoca/SKILL.md` に、Claude CodeがこのCLIを操作するためのガイドが
あります。このリポジトリをClaude Codeのプロジェクトに含めるか、
`~/.claude/skills/` 配下に `skills/misoca` をコピーしてください。

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
