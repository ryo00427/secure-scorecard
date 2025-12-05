# CLAUDE.md - プロジェクト固有の指示

## 重要: 公式ドキュメントを必ず参照する

**LLM の学習データは古い可能性がある。** 設定やAPI仕様は常に公式ドキュメントで最新情報を確認すること。

### 必ず確認すべき公式ドキュメント

| 技術          | URL                                                               |
| ------------- | ----------------------------------------------------------------- |
| Turborepo     | https://turbo.build/repo/docs                                     |
| pnpm          | https://pnpm.io/ja/                                               |
| Expo          | https://docs.expo.dev/                                            |
| React Native  | https://reactnative.dev/docs/getting-started                      |
| Echo (Go)     | https://echo.labstack.com/docs                                    |
| Terraform AWS | https://registry.terraform.io/providers/hashicorp/aws/latest/docs |

### Context7 MCP ツールの活用

`mcp__context7__resolve-library-id` と `mcp__context7__get-library-docs` を使って最新ドキュメントを取得できる。

```
例: Turborepo の最新ドキュメントを取得
1. resolve-library-id で "turborepo" を検索
2. get-library-docs で該当トピックのドキュメントを取得
```

## プロジェクト概要

家庭菜園管理アプリ - Turborepo モノレポ構成

## 過去の失敗から学んだ教訓

## コード実装の際には設計原則を遵守する事

以下の.mdを必ず参照する事
.kiro\steering\product.md
.kiro\steering\structure.md
.kiro\steering\tech.md

### 実装の際には日本語コメントを追加する

**要件**: コードを実装する際は、日本語で解説コメントを入れること

**コメントを入れる場所**:

```go
// パッケージレベルのドキュメント
// Package handler - タスク管理のHTTPハンドラ
//
// エンドポイント:
//   - GET /api/v1/tasks - 全タスク取得
//   - POST /api/v1/tasks - 新規作成
package handler

// 構造体の説明
// CreateTaskRequest はタスク作成リクエストの構造体です。
type CreateTaskRequest struct {
    Title string `json:"title"` // タスクのタイトル
}

// 関数/メソッドの説明
// CreateTask は新しいタスクを作成します。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - task: 作成するタスク
//
// 戻り値:
//   - error: 作成に失敗した場合のエラー
func (s *Service) CreateTask(ctx context.Context, task *model.Task) error {
    // ビジネスロジックの説明
    return s.repos.Task().Create(ctx, task)
}
```

**コメントの書き方**:

```
□ パッケージ: ファイルの目的と提供する機能
□ 構造体: 何を表すか、各フィールドの意味
□ 関数/メソッド: 処理内容、引数、戻り値の説明
□ 複雑なロジック: なぜその実装になっているか
□ セクション区切り: ==== で視覚的に分離
```

**重要**: コードが複雑になるほど、コメントは必須

### 0. IDEの診断エラーは実際のビルドで確認する

**問題**: ファイル編集後、IDEが「imported and not used」などのエラーを表示し続ける

**原因**:

- IDEのキャッシュ問題で、変更が即座に反映されない
- 特に複数ファイルを連続して編集する場合、IDEが追いつかない

**対策**:

```
IDEエラー確認フロー:
1. IDEで赤い波線が表示される
2. ⚠️ すぐに修正しようとしない
3. ✅ まず `pnpm build` または `go build` を実行
4. ✅ ビルドが成功すればIDEのキャッシュ問題
5. ✅ ビルドが失敗すれば実際の問題
```

**重要**: IDEの診断よりもコンパイラの結果を信頼する

### 0.1 関連する変更は同時に行う

**問題**: 複数ファイルにまたがる変更を段階的に行うと、途中で不整合が生じる

**原因**:

- 関数シグネチャを変更したが、呼び出し側を後で変更した
- 新しいパッケージをインポートしたが、使用箇所を後で追加した

**対策**:

```
複数ファイル変更時のベストプラクティス:
□ 変更の影響範囲を事前に把握する
□ 関連するファイルを同時に編集する
  例: middleware.SetupMiddleware(e, cfg)
  - middleware.go: 関数シグネチャ変更
  - main.go: 呼び出し側変更
  → これらを1つのメッセージで同時に実行
□ 中間状態でビルドエラーが出るのは正常
□ 最後にビルドして全体の整合性を確認
```

### 0.2 環境構築は「動作確認」まで完了させる（最重要）

**失敗例**: 設定ファイルを作成して「環境構築完了」としたが、実際に動かすとエラーが多発

**原因**:

- 設定ファイルを書いただけで終わり、実際の動作確認をしていない
- 「ファイルがある」≠「動く」という認識の欠如
- 部分的な確認（例: `pnpm install` 成功）で全体が動くと思い込む

**対策**: 環境構築完了の定義を明確にする

```
環境構築完了チェックリスト:
□ pnpm install が成功する
□ pnpm build が全パッケージで成功する
□ pnpm lint がエラーなく完了する
□ pnpm type-check がエラーなく完了する
□ pnpm test が実行できる（テストがあれば）
□ 各アプリが起動できる:
  □ apps/mobile: pnpm start で Expo が起動
  □ apps/backend: go run が成功
□ 共有パッケージがインポートできる
```

**重要**: 上記すべてが通るまで「環境構築完了」と言わない

### 0.1 仕様ドキュメントと実装の整合性を保つ

**失敗例**: tasks.md に `packages/backend` と書いたが、実際は `apps/backend` で実装

**原因**:

- 設計フェーズと実装フェーズで構成を変更した
- 変更後にドキュメントを更新しなかった

**対策**:

- 構成を変更したら、関連ドキュメントも即座に更新
- テストファイルの期待値も実装に合わせる
- 定期的に `仕様 vs 実装` の差分をチェック

```
構成変更時のドキュメント更新チェックリスト:
□ .kiro/specs/*/tasks.md
□ .kiro/specs/*/design.md
□ README.md
□ CLAUDE.md
□ tests/ 内のテスト期待値
□
```

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

### 5. 設定ファイルは「動く最小構成」から始める

**失敗例**: Expo プロジェクトを手動で設定したら、エントリーポイント・依存関係・アセットが欠落

**原因**:

- 設定を「部分的に」追加していくと、依存関係の全体像を見失う
- 公式テンプレートが持つ暗黙の前提条件を見落とす

**対策**:

```bash
# 悪い例: 空のフォルダに手動で設定を追加
mkdir my-app && cd my-app
# package.json, app.json, babel.config.js を手動作成...

# 良い例: 公式テンプレートから始める
npx create-expo-app my-app --template blank-typescript
# 必要に応じて設定を変更
```

### 7. pnpm モノレポでは Expo カスタムエントリーポイントが必要

**失敗例**: `Unable to resolve "../../App" from "node_modules\expo\AppEntry.js"`

**原因**:

- `expo/AppEntry.js` は `../../App` を相対パスでインポートする
- pnpm ホイスティングにより `expo` がルートの `node_modules` に配置される
- 結果、`../../App` が正しいパスを指さない

**対策**: カスタムエントリーポイントを作成

```javascript
// apps/mobile/index.js
import { registerRootComponent } from 'expo';
import App from './App';

registerRootComponent(App);
```

```json
// apps/mobile/package.json
{
  "main": "./index.js"
}
```

**チェックリスト（Expo プロジェクト）**:

```
□ package.json: "main": "./index.js" (モノレポの場合)
□ app.json: expo.name, expo.slug, expo.version
□ assets/: icon.png, splash-icon.png, adaptive-icon.png, favicon.png
□ babel.config.js: babel-preset-expo
□ .npmrc: shamefully-hoist=true (pnpm の場合)
□ 依存関係: expo, react, react-native, react-native-web, react-dom
```

### 6. pnpm + Expo は hoisting 設定が必要

**失敗例**: `Cannot find module 'babel-preset-expo'` エラー

**原因**:

- pnpm はデフォルトで厳格な依存関係解決（phantom dependencies を防ぐ）
- Expo/React Native は依存関係の hoisting を前提としている

**対策**: `.npmrc` を作成

```
shamefully-hoist=true
node-linker=hoisted
```

### 8. N+1問題を防ぐ（データベース操作）

**失敗例**: DeleteGarden メソッドでループ内で個別削除を実行（1 + N + 1 クエリ）

```go
// 🔴 悪い例: N+1問題
plants, _ := repo.GetByGardenID(ctx, gardenID)  // 1回
for _, plant := range plants {
    repo.Delete(ctx, plant.ID)  // N回（plant の数だけクエリ）
}
repo.DeleteGarden(ctx, gardenID)  // 1回
// 合計: 1 + N + 1 クエリ
```

**原因**:

- 単一削除メソッドをループで再利用した（実装が簡単だった）
- パフォーマンスへの配慮不足（「動けばいい」で実装）
- 少量データでは問題が顕在化しない（3-5個なら気づかない）
- トランザクション内なのでエラーが出ない

**対策**: バッチ操作を優先する

```go
// ✅ 良い例: バッチ削除
repo.DeleteByGardenID(ctx, gardenID)  // 1回（WHERE条件で一括削除）
repo.DeleteGarden(ctx, gardenID)      // 1回
// 合計: 2 クエリ
```

**N+1問題チェックリスト**:

```
実装時:
□ ループ内でDB操作をしていないか？
  - repository メソッド呼び出し
  - SELECT/UPDATE/DELETE クエリ
□ 取得したリストを1件ずつ処理していないか？
□ バッチ操作が可能か？
  - WHERE IN (ids...)
  - WHERE column = value （複数行に適用）
  - GORM の Create/Update/Delete with slice

コードレビュー時:
□ for/range ループ内に repository 呼び出しがないか
□ トランザクション内のクエリ数を数える
□ 「N件のデータがあったら何クエリ？」を想像する
```

**GORM のバッチ操作パターン**:

```go
// 複数IDで削除
db.Delete(&Model{}, []uint{1, 2, 3, 4, 5})

// WHERE条件で一括削除
db.Where("parent_id = ?", parentID).Delete(&Model{})

// 複数件を一括作成
db.Create(&[]Model{...})

// 複数件を一括更新
db.Model(&Model{}).Where("status = ?", "old").Update("status", "new")
```

**重要**: ループでDB操作を見たら N+1 問題を疑う

### 9. Repository Manager に新しいリポジトリを追加する際の3ステップ

**失敗例**: `transaction.go` に新しいリポジトリのフィールドと初期化を追加したが、アクセサメソッドを忘れた

```go
// 🔴 悪い例: 2ステップだけ実行（3つ目を忘れた）
type repositoryManager struct {
    db   *gorm.DB
    task *taskRepository  // ✅ ステップ1: フィールド追加
}

func NewRepositoryManager(db *gorm.DB) Repositories {
    return &repositoryManager{
        db:   db,
        task: &taskRepository{db: db},  // ✅ ステップ2: 初期化
    }
}
// ❌ ステップ3 を忘れた: Task() TaskRepository メソッド
```

**原因**:

- 構造体フィールドと初期化だけで「完了」と思い込んだ
- インターフェースを満たすためのメソッドを見落とした
- ビルドするまでエラーに気づかない

**対策**: 3ステップを必ず実行する

```go
// ✅ 良い例: 3ステップすべて実行

// ステップ1: 構造体にフィールド追加
type repositoryManager struct {
    db   *gorm.DB
    task *taskRepository
}

// ステップ2: コンストラクタで初期化
func NewRepositoryManager(db *gorm.DB) Repositories {
    return &repositoryManager{
        db:   db,
        task: &taskRepository{db: db},
    }
}

// ステップ3: アクセサメソッド追加（これを忘れがち！）
func (m *repositoryManager) Task() TaskRepository {
    return m.task
}
```

**新しいリポジトリ追加チェックリスト**:

```
Repository Manager (transaction.go):
□ ステップ1: repositoryManager 構造体にフィールド追加
□ ステップ2: NewRepositoryManager() で初期化
□ ステップ3: アクセサメソッド追加 (例: Task() TaskRepository)

インターフェース (interfaces.go):
□ 新しい Repository インターフェース定義
□ Repositories インターフェースにメソッド追加

モック (mock_repository.go):
□ Mock構造体の定義
□ MockRepositories に追加
□ NewMockRepositories() で初期化
□ アクセサメソッド追加
□ GetMock*Repository() ヘルパー追加
```

**重要**: インターフェースを満たすには、フィールド追加だけでなくメソッド追加が必須

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
2. **Expo managed workflow**: モノレポではカスタム `index.js` を作成し `"main": "./index.js"` を設定
3. **pnpm workspaces**: `apps/*` と `packages/*` の両方を設定
4. **pnpm + Expo**: `.npmrc` に `shamefully-hoist=true` が必要
