# CLAUDE.md - プロジェクト固有の指示

## プロジェクト概要

家庭菜園管理アプリ - Turborepo モノレポ構成

## 過去の失敗から学んだ教訓

### 1. 構造変更時は全ファイルの整合性を確認する

**失敗例**: `packages/` → `apps/` + `packages/` に変更した際、以下を更新し忘れた：
- テストファイル（期待値が旧構造のまま）
- `package.json` の workspaces（`pnpm-workspace.yaml` だけ更新）

**対策**:
```
構造変更時のチェックリスト:
□ pnpm-workspace.yaml
□ package.json (workspaces)
□ turbo.json
□ テストファイルの期待値
□ README.md の構成図
□ .gitignore
```

### 2. 設定ファイルのバージョン形式を確認する

**失敗例**: Turbo v2 の `tasks` 形式で書いたが、テストは v1 の `pipeline` を期待していた

**対策**:
- Turborepo 2.x では `pipeline` → `tasks` に変更された
- テスト作成時は実際の設定ファイルの形式を確認してから期待値を書く

### 3. package.json スクリプトには実体を用意する

**失敗例**: `apps/backend/package.json` に Go スクリプトを書いたが、`go.mod` も `cmd/server/` も作らなかった

**対策**:
- ビルド/テストスクリプトを書いたら、最低限のスケルトンコードも作成する
- Go の場合: `go.mod` + `cmd/xxx/main.go`
- TypeScript の場合: `tsconfig.json` + `src/index.ts`

### 4. 二重管理される設定に注意

**失敗例**: pnpm は `pnpm-workspace.yaml` と `package.json` の workspaces 両方を見る

**対策**:
- pnpm モノレポでは両方を同期させる必要がある
- `pnpm-workspace.yaml` が優先されるが、npm/yarn 互換のため `package.json` も更新する

## モノレポ構成

```
secure-scorecard/
├── apps/                # デプロイ可能なアプリケーション
│   ├── backend/         # Go Echo バックエンド
│   └── mobile/          # React Native (Expo) モバイルアプリ
├── packages/            # 共有ライブラリ
│   └── shared/          # 共通型定義・ユーティリティ
├── infrastructure/      # Terraform (AWS)
└── tests/               # 統合テスト
```

## 技術スタック

- **バックエンド**: Go 1.23+, Echo v4
- **モバイル**: React Native 0.76+, Expo SDK 52
- **モノレポ**: Turborepo 2.x (tasks 形式), pnpm 9.x
- **IaC**: Terraform

## 開発時の注意点

1. **Turborepo 2.x**: `pipeline` ではなく `tasks` を使用
2. **Expo SDK 52**: `package.json` の `main` フィールドは不要
3. **pnpm workspaces**: `apps/*` と `packages/*` の両方を設定
