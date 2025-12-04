# Secure Scorecard

家庭菜園管理アプリ - Turborepo モノレポ

## 📦 モノレポ構成

```
secure-scorecard/
├── apps/                # 実行可能なアプリケーション
│   ├── backend/         # Go Echo バックエンド
│   └── mobile/          # React Native (Expo) モバイルアプリ
├── packages/            # 共有ライブラリ
│   └── shared/          # 共通型定義・ユーティリティ
├── infrastructure/      # Terraform (AWS)
├── .kiro/               # 仕様駆動開発
│   ├── steering/        # プロジェクト設計原則
│   └── specs/           # 機能仕様
├── package.json         # ルート設定
├── pnpm-workspace.yaml  # pnpm ワークスペース設定
└── turbo.json           # Turborepo パイプライン設定
```

## 🚀 セットアップ

### 必要な環境

- Node.js >= 18.0.0
- pnpm >= 9.0.0
- Go >= 1.23
- Terraform >= 1.x

### インストール

```bash
# 依存関係をインストール
pnpm install

# すべてのパッケージをビルド
pnpm build

# 開発サーバー起動（全パッケージ）
pnpm dev
```

## 📱 パッケージ別コマンド

### バックエンド (Go)

```bash
cd apps/backend
go run ./cmd/server          # 開発サーバー起動
go test ./...                # テスト実行
golangci-lint run            # リント実行
```

### モバイル (React Native + Expo)

```bash
cd apps/mobile
pnpm start                   # Expo 開発サーバー起動
pnpm android                 # Android エミュレータで起動
pnpm ios                     # iOS シミュレータで起動
pnpm test                    # テスト実行
```

### 共有ライブラリ (TypeScript)

```bash
cd packages/shared
pnpm build                   # TypeScript ビルド
pnpm dev                     # Watch モードでビルド
pnpm test                    # テスト実行
```

## 🏗️ Turborepo コマンド

```bash
# すべてのパッケージをビルド
pnpm build

# すべてのパッケージでテスト実行
pnpm test

# すべてのパッケージでリント実行
pnpm lint

# コードフォーマット
pnpm format
```

## 🎯 開発ワークフロー

1. **ブランチ作成**: `git checkout -b feature/xxx` または `task/x.x-xxx`
2. **実装**: TDD (Test-Driven Development) で実装
3. **テスト**: `pnpm test` ですべてのテストをパス
4. **コミット**: 意味のある単位でコミット
5. **PR作成**: `feature/xxx` → `main` へプルリクエスト

## 📚 技術スタック

- **バックエンド**: Go 1.23+, Echo v4, GORM v2, PostgreSQL 16
- **モバイル**: React Native 0.76+, Expo SDK 52, TypeScript 5.7
- **状態管理**: Zustand
- **モノレポ**: Turborepo 2.x, pnpm 9.x
- **インフラ**: AWS (ECS Fargate, RDS, S3, CloudFront)
- **IaC**: Terraform

## 🔧 仕様駆動開発 (Kiro)

プロジェクトは `.kiro/` ディレクトリで管理されています:

- `.kiro/steering/`: 設計原則（product.md, tech.md, structure.md）
- `.kiro/specs/`: 機能仕様（requirements, design, tasks）

詳細は [CLAUDE.md](./CLAUDE.md) を参照してください。

## 📄 ライセンス

MIT