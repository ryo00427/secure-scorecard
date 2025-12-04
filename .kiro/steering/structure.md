# コード構成・命名規則

## モノレポ構造

```
secure-scorecard/
├── .github/
│   └── workflows/          # GitHub Actions CI/CD設定
├── .kiro/
│   ├── steering/           # プロジェクト全体の設計原則（本ファイル含む）
│   └── specs/              # 機能ごとの仕様（requirements.md, design.md, tasks.md）
├── infrastructure/
│   ├── terraform/
│   │   ├── modules/        # 再利用可能なTerraformモジュール
│   │   ├── environments/
│   │   │   ├── dev/
│   │   │   ├── stg/
│   │   │   └── prod/
│   │   └── terraform.tfvars
│   └── scripts/            # デプロイスクリプト
├── packages/
│   ├── backend/            # Go Echo バックエンド
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go          # エントリーポイント
│   │   ├── internal/
│   │   │   ├── handler/             # HTTP Handler層
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── crop_handler.go
│   │   │   │   ├── plot_handler.go
│   │   │   │   ├── task_handler.go
│   │   │   │   ├── analytics_handler.go
│   │   │   │   └── notification_handler.go
│   │   │   ├── service/             # Service層（ビジネスロジック）
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── crop_service.go
│   │   │   │   ├── plot_service.go
│   │   │   │   ├── task_service.go
│   │   │   │   ├── analytics_service.go
│   │   │   │   └── notification_service.go
│   │   │   ├── repository/          # Repository層（データアクセス）
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── crop_repository.go
│   │   │   │   ├── plot_repository.go
│   │   │   │   ├── task_repository.go
│   │   │   │   └── notification_repository.go
│   │   │   ├── model/               # データベースモデル（GORMモデル）
│   │   │   │   ├── user.go
│   │   │   │   ├── crop.go
│   │   │   │   ├── plot.go
│   │   │   │   ├── task.go
│   │   │   │   └── notification.go
│   │   │   ├── domain/              # ドメインモデル（ビジネスエンティティ）
│   │   │   │   ├── crop.go
│   │   │   │   ├── plot.go
│   │   │   │   └── task.go
│   │   │   ├── middleware/          # Echoミドルウェア
│   │   │   │   ├── auth.go          # JWT認証
│   │   │   │   ├── cors.go
│   │   │   │   └── logging.go
│   │   │   ├── util/                # ユーティリティ関数
│   │   │   │   ├── jwt.go
│   │   │   │   ├── password.go
│   │   │   │   └── validator.go
│   │   │   └── config/              # 設定管理
│   │   │       └── config.go
│   │   ├── migrations/              # SQLマイグレーションファイル
│   │   │   ├── 001_create_users.up.sql
│   │   │   ├── 001_create_users.down.sql
│   │   │   └── ...
│   │   ├── tests/                   # テストコード
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   └── repository/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   ├── mobile/                      # React Native (Expo) アプリ
│   │   ├── src/
│   │   │   ├── screens/             # 画面コンポーネント
│   │   │   │   ├── auth/
│   │   │   │   ├── dashboard/
│   │   │   │   ├── crop/
│   │   │   │   └── task/
│   │   │   ├── components/          # 再利用可能なコンポーネント
│   │   │   ├── navigation/          # React Navigation設定
│   │   │   ├── hooks/
│   │   │   ├── stores/
│   │   │   ├── types/
│   │   │   └── utils/
│   │   ├── assets/                  # 画像、フォント等
│   │   ├── app.json
│   │   ├── App.tsx
│   │   ├── tsconfig.json
│   │   └── package.json
│   └── shared/                      # 共通ライブラリ（TypeScript）
│       ├── src/
│       │   ├── types/               # 共通型定義
│       │   │   ├── api-response.ts
│       │   │   ├── crop.ts
│       │   │   └── task.ts
│       │   └── constants/           # 共通定数
│       │       ├── crop-types.ts
│       │       └── task-types.ts
│       ├── tsconfig.json
│       └── package.json
├── pnpm-workspace.yaml
├── turbo.json
├── package.json
├── .gitignore
└── README.md
```

## ドメイン境界

以下の6つのドメインで構成されます。各ドメインは独立したHandler、Service、Repositoryを持ちます。

### 1. User ドメイン

**責務**: ユーザー認証・プロフィール管理

**エンティティ**:
- `User`: ユーザー情報
- `TokenBlacklist`: JWT無効化リスト

**API Endpoints**:
- `POST /api/auth/register`: ユーザー登録
- `POST /api/auth/login`: ログイン
- `POST /api/auth/logout`: ログアウト
- `GET /api/users/me`: 自分のプロフィール取得
- `PUT /api/users/me`: プロフィール更新

### 2. Crop ドメイン

**責務**: 作物のライフサイクル管理

**エンティティ**:
- `Crop`: 作物基本情報
- `CropPhoto`: 作物写真
- `CropNote`: 作物メモ

**API Endpoints**:
- `GET /api/crops`: 作物一覧取得
- `POST /api/crops`: 作物作成
- `GET /api/crops/:id`: 作物詳細取得
- `PUT /api/crops/:id`: 作物更新
- `DELETE /api/crops/:id`: 作物削除（論理削除）
- `POST /api/crops/:id/photos`: 写真追加
- `POST /api/crops/:id/harvest`: 収穫記録
- `POST /api/crops/:id/notes`: メモ追加

### 3. Plot ドメイン

**責務**: 区画管理・空間配置

**エンティティ**:
- `Plot`: 区画情報
- `PlotAssignment`: 区画割り当て（Crop ↔ Plot の多対多）

**API Endpoints**:
- `GET /api/plots`: 区画一覧取得
- `POST /api/plots`: 区画作成
- `GET /api/plots/:id`: 区画詳細取得
- `PUT /api/plots/:id`: 区画更新
- `DELETE /api/plots/:id`: 区画削除
- `POST /api/plots/:id/assign`: 作物割り当て
- `DELETE /api/plots/:id/assign/:crop_id`: 割り当て解除

### 4. Task ドメイン

**責務**: 作業タスク・リマインダー管理

**エンティティ**:
- `Task`: タスク情報
- `RecurringTask`: 繰り返しタスク設定

**API Endpoints**:
- `GET /api/tasks`: タスク一覧取得
- `POST /api/tasks`: タスク作成
- `GET /api/tasks/:id`: タスク詳細取得
- `PUT /api/tasks/:id`: タスク更新
- `DELETE /api/tasks/:id`: タスク削除
- `POST /api/tasks/:id/complete`: タスク完了
- `POST /api/tasks/recurring`: 繰り返しタスク作成
- `GET /api/tasks/recurring`: 繰り返しタスク一覧
- `DELETE /api/tasks/recurring/:id`: 繰り返しタスク削除

### 5. Analytics ドメイン

**責務**: データ分析・統計計算

**エンティティ**:
- なし（他ドメインのデータを集計）

**API Endpoints**:
- `GET /api/analytics/crops/summary`: 作物栽培統計
- `GET /api/analytics/harvest`: 収穫量統計
- `GET /api/analytics/growth/:crop_id`: 成長曲線データ
- `GET /api/analytics/tasks/completion-rate`: タスク完了率

### 6. Notification ドメイン

**責務**: 通知配信・履歴管理

**エンティティ**:
- `Notification`: 通知履歴
- `NotificationSetting`: 通知設定

**API Endpoints**:
- `GET /api/notifications`: 通知履歴取得
- `PUT /api/notifications/:id/read`: 既読マーク
- `GET /api/notifications/settings`: 通知設定取得
- `PUT /api/notifications/settings`: 通知設定更新

## 命名規則

### バックエンド（Go）

| 対象 | 規則 | 例 |
|------|------|---|
| **パッケージ名** | 小文字単語（単数形） | `service`, `repository`, `handler` |
| **ファイル名** | snake_case | `crop_service.go`, `user_repository.go` |
| **構造体** | PascalCase | `CropService`, `UserRepository`, `TaskHandler` |
| **インターフェース** | PascalCase | `CropRepository`, `NotificationService` |
| **関数（公開）** | PascalCase | `GetCropByID`, `CreateTask`, `SendNotification` |
| **関数（非公開）** | camelCase | `validateCropData`, `hashPassword` |
| **変数** | camelCase | `cropID`, `userEmail`, `taskList` |
| **定数** | PascalCase または UPPER_SNAKE_CASE | `MaxPhotoSize` または `MAX_PHOTO_SIZE` |
| **メソッドレシーバ** | 1-2文字の小文字 | `s *CropService`, `r *UserRepository` |

### フロントエンド（TypeScript/JavaScript）

| 対象 | 規則 | 例 |
|------|------|---|
| **ファイル名** | kebab-case | `crop-service.ts`, `use-auth.ts` |
| **コンポーネント** | PascalCase | `CropCard.tsx`, `DashboardLayout.tsx` |
| **関数** | camelCase | `getCropById`, `handleSubmit`, `formatDate` |
| **変数** | camelCase | `cropList`, `isLoading`, `userEmail` |
| **型・インターフェース** | PascalCase | `CropData`, `ApiResponse`, `TaskStatus` |
| **Enum** | PascalCase（キーもPascalCase） | `CropType.MiniTomato`, `TaskStatus.Completed` |
| **定数** | UPPER_SNAKE_CASE | `API_BASE_URL`, `MAX_PHOTO_SIZE` |
| **React Hooks** | camelCase（`use`プレフィックス） | `useAuth`, `useCrops`, `useLocalStorage` |
| **イベントハンドラー** | camelCase（`handle`プレフィックス） | `handleClick`, `handleSubmit`, `handleChange` |

### データベース（PostgreSQL）

| 対象 | 規則 | 例 |
|------|------|---|
| **テーブル名** | snake_case（複数形） | `users`, `crops`, `crop_photos` |
| **カラム名** | snake_case | `user_id`, `created_at`, `crop_type` |
| **インデックス名** | `idx_<table>_<column(s)>` | `idx_crops_user_id`, `idx_tasks_due_date` |
| **外部キー名** | `fk_<table>_<column>` | `fk_crops_user_id`, `fk_tasks_crop_id` |

### API Endpoint

| 規則 | 例 |
|------|---|
| **リソース名** | 複数形、kebab-case | `/api/crops`, `/api/recurring-tasks` |
| **IDパラメータ** | `:id` | `/api/crops/:id`, `/api/plots/:id` |
| **アクション** | 動詞を追加 | `/api/tasks/:id/complete`, `/api/crops/:id/harvest` |

## レイヤー間のデータフロー

### リクエスト処理フロー

```
[Client]
  ↓ HTTP Request
[Handler Layer]
  - リクエストバリデーション
  - JWT認証チェック（middleware）
  ↓
[Service Layer]
  - ビジネスロジック実行
  - トランザクション制御
  ↓
[Repository Layer]
  - データベースクエリ実行
  - データマッピング
  ↓
[Database]
  - PostgreSQL操作
  ↓
[Repository Layer]
  - ドメインモデルに変換
  ↓
[Service Layer]
  - ビジネスルール適用
  ↓
[Handler Layer]
  - レスポンス形成
  ↓ HTTP Response
[Client]
```

### データ型の変換

| Layer | データ型 | 責務 |
|-------|---------|------|
| **Handler** | Request/Response DTO | JSON ↔ 構造体変換 |
| **Service** | Domain Model | ビジネスロジック適用 |
| **Repository** | GORM Model | Database ↔ ドメインモデル変換 |
| **Database** | SQL Rows | データ永続化 |

**例**（作物作成の場合）:

```go
// Handler Layer: Request DTO
type CreateCropRequest struct {
    Name     string `json:"name" validate:"required"`
    CropType string `json:"crop_type" validate:"required"`
}

// Service Layer: Domain Model
type Crop struct {
    ID          string
    Name        string
    CropType    CropType
    UserID      string
    Status      CropStatus
    CreatedAt   time.Time
}

// Repository Layer: GORM Model
type CropModel struct {
    ID        uuid.UUID      `gorm:"type:uuid;primary_key"`
    Name      string         `gorm:"not null"`
    CropType  string         `gorm:"not null"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null"`
    Status    string         `gorm:"not null"`
    CreatedAt time.Time      `gorm:"not null"`
    UpdatedAt time.Time      `gorm:"not null"`
    DeletedAt *time.Time     `gorm:"index"`
}
```

## 依存関係の方向

Clean Architectureに従い、依存関係は以下の方向に限定されます：

```
Handler → Service → Repository → Database
   ↓         ↓          ↓
  なし    Domain     Model
```

**禁止事項**:
- Handler が Repository を直接呼び出すこと
- Repository が Service を呼び出すこと
- 下位層が上位層に依存すること

## テストファイル配置

- **単体テスト**: 同一ディレクトリに `_test.go` / `.test.ts` サフィックス
  - 例: `crop_service.go` → `crop_service_test.go`
  - 例: `use-auth.ts` → `use-auth.test.ts`
- **統合テスト**: `tests/` ディレクトリ配下
- **E2Eテスト**: `tests/e2e/` ディレクトリ

## 環境変数管理

### バックエンド（Go）

`.env` ファイル + viper による読み込み

```bash
# .env.example
DATABASE_URL=postgresql://user:password@localhost:5432/mygardendb
JWT_SECRET=your-secret-key-here
AWS_REGION=ap-northeast-1
S3_BUCKET_NAME=mygarden-photos
```

### フロントエンド（React Native）

`.env` ファイル（Expoの環境変数）

```bash
# .env.example
API_BASE_URL=http://localhost:8080/api
ENABLE_ANALYTICS=true
```

**重要**: `.env` ファイルは `.gitignore` に含め、`.env.example` のみコミットする。

## Git ブランチ戦略

- **`main`**: 本番環境（`prod`）デプロイブランチ
- **`develop`**: 開発統合ブランチ
- **`feature/<feature-name>`**: 機能開発ブランチ
- **`fix/<bug-name>`**: バグ修正ブランチ
- **`hotfix/<issue-name>`**: 緊急修正ブランチ

### コミットメッセージ

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**:
- `feat`: 新機能
- `fix`: バグ修正
- `refactor`: リファクタリング
- `test`: テスト追加・修正
- `docs`: ドキュメント更新
- `chore`: ビルド・設定変更

**例**:
```
feat(crop): Add harvest recording endpoint

Implement POST /api/crops/:id/harvest with validation
and notification trigger.

Closes #123
```

## ドキュメント管理

- **API仕様**: OpenAPI 3.0（Swagger）形式で `docs/api/openapi.yaml` に記述
- **アーキテクチャ図**: Mermaid形式で `.kiro/specs/*/design.md` に埋め込み
- **開発者ガイド**: `docs/developer-guide.md`
- **デプロイ手順**: `docs/deployment.md`

## コードレビュー基準

### 必須チェック項目

- [ ] ビジネスロジックが Service 層に集約されているか
- [ ] エラーハンドリングが適切に実装されているか
- [ ] テストが追加されているか（カバレッジ80%以上）
- [ ] 命名規則に従っているか
- [ ] 不要なコメントが残っていないか
- [ ] 論理削除パターンが守られているか
- [ ] JWT認証が必要なエンドポイントで適用されているか

### パフォーマンスチェック

- [ ] N+1問題が発生していないか
- [ ] 不要なgoroutine/非同期処理が含まれていないか
- [ ] データベースインデックスが適切に設定されているか
- [ ] 画像サイズが最適化されているか（< 5MB）

## 実装優先順位

仕様書の tasks.md に記載された順序で実装を進めます：

1. **基盤構築**: モノレポ初期化、DB設計、認証機能
2. **コアドメイン**: User → Crop → Plot → Task の順
3. **周辺機能**: Analytics、Notification
4. **フロントエンド**: Web → Mobile の順
5. **最適化**: パフォーマンス改善、テスト拡充

各ドメインの実装は「CRUD完成 → 写真・メモ機能 → 通知連携」の順で進めます。
