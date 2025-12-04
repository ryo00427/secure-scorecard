# 技術設計原則

## アーキテクチャ原則

### 全体アーキテクチャ

**Clean Architecture + Repository Pattern** を採用し、ビジネスロジックの独立性と保守性を確保します。

```
[Presentation Layer]  ← HTTP/WebSocket
    ↓
[Handler Layer]       ← ルーティング、リクエスト/レスポンス変換
    ↓
[Service Layer]       ← ビジネスロジック（ドメイン層）
    ↓
[Repository Layer]    ← データアクセス抽象化
    ↓
[Database Layer]      ← PostgreSQL、AWS S3
```

### レイヤー責務

1. **Handler Layer**
   - HTTPリクエストのバリデーション
   - JWT認証チェック（middleware）
   - レスポンス形成（JSON/エラーハンドリング）
   - ビジネスロジックは含まない

2. **Service Layer**
   - ビジネスルールの実装
   - トランザクション境界の定義
   - 複数Repositoryの協調
   - ドメインイベントの発行

3. **Repository Layer**
   - データベース操作の抽象化
   - クエリの最適化
   - データマッピング（DB ↔ Domain Model）

4. **Database Layer**
   - PostgreSQL: トランザクショナルデータ
   - S3: 画像ファイル、静的アセット

## 技術スタック

### バックエンド

| カテゴリ | 技術 | バージョン | 選定理由 |
|---------|------|----------|---------|
| **言語** | Go | 1.23+ | 高速、型安全、goroutineによる並行処理 |
| **Webフレームワーク** | Echo | v4 | 軽量、高速、ミドルウェア豊富 |
| **ORM** | GORM | v2 | 型安全、マイグレーション対応、フック機能 |
| **JWT** | golang-jwt | v5 | 標準的なJWT実装 |
| **バリデーション** | go-playground/validator | v10 | 構造体タグベースのバリデーション |
| **環境変数** | viper | - | 設定管理の標準 |
| **ログ** | zerolog | - | 構造化ログ、高速 |

#### Go コーディング規約

- **命名規則**:
  - パッケージ名: 小文字単語（例: `service`, `repository`）
  - 構造体: PascalCase（例: `CropService`, `UserRepository`）
  - 関数: PascalCase（公開）、camelCase（非公開）
  - 変数: camelCase（例: `cropID`, `userEmail`）
- **エラーハンドリング**:
  - エラーは必ず処理（`if err != nil` チェック）
  - カスタムエラー型で詳細情報を保持
  - ログには必ずエラースタックを含める
- **並行処理**:
  - goroutineは必ずコンテキストで制御
  - チャネルのクローズは送信側が責任を持つ
  - `sync.WaitGroup` でgoroutineの完了を待機
- **Context使用**:
  - Cancel関数は必ず `defer cancel()` で呼び出す（リソースリーク防止）
  - Cancel関数の受け渡しは禁止（下流への影響不明のためキャンセルすると予期しない動作）
  - `select` 文で `<-ctx.Done()` を監視し、キャンセル時に処理中断
  - 親コンテキストがキャンセルされたら、すべての子goroutineを終了
  - タイムアウト設定: HTTP Handler では 30秒、DB操作は 5秒
- **Repository Pattern検証**:
  - データベース（PostgreSQL → MySQL）を交換可能にできれば、Repository Patternが正しく実装されている証明

### フロントエンド（モバイル）

#### React Native + Expo

| カテゴリ | 技術 | バージョン | 選定理由 |
|---------|------|----------|---------|
| **フレームワーク** | React Native | 0.76.x | クロスプラットフォーム |
| **管理ツール** | Expo | SDK 52 | 開発効率、OTA更新、Push通知 |
| **ナビゲーション** | React Navigation | v6 | ネイティブスタック、タブナビゲーション |
| **UI** | React Native Paper | - | Material Design準拠 |
| **状態管理** | Zustand | - | 軽量、TypeScript親和性 |
| **APIクライアント** | Axios | - | HTTPリクエスト、インターセプター |
| **画像** | Expo Image Picker | - | カメラ・ギャラリーアクセス |
| **通知** | Expo Notifications | - | ローカル・リモート通知 |

#### TypeScript/JavaScript コーディング規約

- **命名規則**:
  - ファイル名: kebab-case（例: `crop-service.ts`, `use-auth.ts`）
  - コンポーネント: PascalCase（例: `CropCard.tsx`, `DashboardLayout.tsx`）
  - 関数・変数: camelCase（例: `getCropById`, `cropList`）
  - 型・インターフェース: PascalCase（例: `CropData`, `ApiResponse`）
  - 定数: UPPER_SNAKE_CASE（例: `API_BASE_URL`, `MAX_PHOTO_SIZE`）
- **型安全性**:
  - `any` 型の使用禁止（やむを得ない場合は `unknown` を使用）
  - すべての関数に戻り値の型を明示
  - Props型は `interface` で定義
- **コンポーネント設計**:
  - 単一責任原則: 1コンポーネント = 1つの責務
  - Composition over Inheritance
  - Propsは読み取り専用（イミュータブル）

### React Native アーキテクチャ原則

#### New Architecture 採用（2025年標準）

React Native 0.76+ では **New Architecture** がデフォルトで有効化されます。

**主要コンポーネント**:

- **JSI（JavaScript Interface）**:
  - JavaScriptとネイティブコード間のブリッジレス通信
  - 従来のブリッジより低遅延、高スループット
  - 同期的なネイティブメソッド呼び出しが可能

- **TurboModules**:
  - ネイティブモジュールの遅延ロード
  - 使用時のみロード → アプリ起動時間の短縮
  - 型安全なネイティブインターフェース

- **Fabric（新レンダリングシステム）**:
  - UIレンダリングの並列化
  - 60fps安定動作
  - 起動時間40%短縮、メモリ使用量20-30%削減

- **Codegen**:
  - TypeScript型定義から自動的にネイティブバインディング生成
  - コンパイル時の型チェック

#### プロジェクト構造（Feature-Based）

**Feature-Based構造** を採用し、機能単位でコードを凝集させます。

```
src/
├── features/              # 機能単位でコード凝集
│   ├── auth/
│   │   ├── components/    # 認証専用コンポーネント
│   │   ├── hooks/         # 認証専用Hooks
│   │   ├── screens/       # 認証画面
│   │   ├── services/      # 認証API呼び出し
│   │   └── types/         # 認証型定義
│   ├── crops/
│   ├── plots/
│   └── tasks/
├── shared/                # 共通コード
│   ├── components/        # 再利用可能なUIコンポーネント
│   ├── hooks/             # 汎用Hooks
│   ├── utils/             # ユーティリティ関数
│   ├── types/             # 共通型定義
│   └── api/               # APIクライアント設定
├── navigation/            # ナビゲーション設定
├── stores/                # グローバルステート（Zustand）
└── App.tsx
```

**原則**:
- 各 `features/` 配下のディレクトリは可能な限り自己完結
- 機能間の依存は `shared/` を経由
- ドメインロジックは `services/` に隔離

#### React Hooks 使用原則

**useMemo / useCallback / React.memo の使用基準**

**使用すべき場合** ✅:
- 明らかに遅い計算（ソート、フィルタリング、集計等）
- 子コンポーネントのpropsとして渡すオブジェクト/関数
- 頻繁に再レンダリングされるコンポーネント

**使用すべきでない場合** ❌:
- 単純な計算（算術演算、文字列連結等）
- 依存配列が頻繁に変わる場合（メモ化の効果が薄い）
- 軽量なコンポーネント

**ベストプラクティス**:
1. **まず計測**: React DevTools Profiler で遅延を確認してから最適化
2. **設計優先**: コンポーネント設計を工夫してメモ化を最小限に
3. **リスト最適化**: 10件以上のリストは必ず `FlatList` / `VirtualizedList` を使用

**例**:

```tsx
// ❌ 不要なメモ化
const fullName = useMemo(() => `${firstName} ${lastName}`, [firstName, lastName]);

// ✅ 適切なメモ化
const sortedList = useMemo(() =>
  items.sort((a, b) => a.name.localeCompare(b.name)),
  [items]
);

// ✅ 子コンポーネントのpropsメモ化
const handlePress = useCallback(() => {
  navigation.navigate('Details', { id });
}, [id, navigation]);
```

#### コンポーネント設計パターン

**Atomic Design 原則**

```
Atoms      → 最小単位（Button, Input, Icon）
Molecules  → Atomsの組み合わせ（SearchBar = Input + Icon + Button）
Organisms  → Moleculesの組み合わせ（CropCard, TaskListItem）
Templates  → ページレイアウト（DashboardLayout）
Screens    → 実データを持つ画面（DashboardScreen）
```

**Props型定義**:
- すべてのコンポーネントで `interface` による型定義必須
- オプショナルなpropsにはデフォルト値を設定

```tsx
interface CropCardProps {
  crop: Crop;
  onPress?: () => void;  // オプショナル
  showStatus?: boolean;  // デフォルト: true
}

export const CropCard: React.FC<CropCardProps> = ({
  crop,
  onPress,
  showStatus = true,
}) => { ... };
```

### パフォーマンス最適化

#### 計測ファースト原則

**最適化前に必ず計測**:
- **React DevTools Profiler**: コンポーネントレンダリング時間
- **60fps基準**: アニメーション・スクロールは60fps維持
- **Flipper Memory Profiler**: メモリリーク検出

**最適化は計測結果に基づいて実施**。感覚的な最適化は避ける。

#### 最適化手法

**画像最適化**:
- WebP形式採用（PNGの25-35%サイズ削減）
- CDN配信（CloudFront経由）
- `expo-image` の `contentFit="cover"` で自動最適化

**リスト最適化**:
```tsx
<FlatList
  data={crops}
  renderItem={renderCropCard}
  keyExtractor={item => item.id}
  initialNumToRender={10}
  maxToRenderPerBatch={5}
  windowSize={5}
  removeClippedSubviews={true}  // Android
/>
```

**ナビゲーション最適化**:
- 画面の遅延ロード（Lazy Loading）
- `React.lazy()` と `Suspense` の活用

**バンドルサイズ削減**:
- Hermes JavaScript Engine 有効化（デフォルト）
- 未使用コードの削除（Tree Shaking）
- `import { specific } from 'library'` で必要な関数のみインポート

### インフラストラクチャ（AWS）

| サービス | 用途 | 構成 |
|---------|------|------|
| **ECS Fargate** | バックエンドコンテナ実行 | 2タスク以上（高可用性） |
| **ALB** | ロードバランシング | ヘルスチェック、HTTPS終端 |
| **RDS PostgreSQL** | データベース | 16.x Multi-AZ、自動バックアップ |
| **S3** | 画像・静的ファイルストレージ | Lifecycle Policy（30日後Glacier） |
| **CloudFront** | CDN | S3オリジン、キャッシュTTL 1時間 |
| **SNS** | Push通知配信 | iOS/Android両対応 |
| **SES** | Email通知配信 | バウンス・苦情処理 |
| **Secrets Manager** | JWT Secret、DB認証情報 | 自動ローテーション |
| **CloudWatch** | ログ・メトリクス監視 | アラーム設定（エラー率、レイテンシ） |

#### Infrastructure as Code

- **Terraform**: すべてのAWSリソースをコード管理
- **バージョン管理**: Git + モジュール化
- **環境分離**: `dev`, `stg`, `prod` 環境を Workspace で管理

### モノレポ構成

**Turborepo 2.x + pnpm 9.x** によるモノレポ構成を採用します。

```
secure-scorecard/
├── packages/
│   ├── backend/              # Go Echo バックエンド
│   ├── mobile/               # React Native (Expo)
│   └── shared/               # 共通型定義・ユーティリティ（TypeScript）
├── infrastructure/           # Terraform設定
├── .kiro/
│   ├── steering/            # プロジェクト全体の設計原則
│   └── specs/               # 機能ごとの仕様
├── pnpm-workspace.yaml
├── turbo.json
└── package.json
```

## 認証・セキュリティ

### JWT認証

- **トークン配信**: HttpOnly Cookie（XSS対策）
- **有効期限**: 24時間
- **リフレッシュ**: 期限切れ前に自動更新
- **ブラックリスト**: ログアウト時にトークンを PostgreSQL に記録（Redis不使用）
  - テーブル: `token_blacklist`
  - カラム: `jti UUID`, `expires_at TIMESTAMP`
  - インデックス: `(jti, expires_at)` 複合インデックス

### パスワード管理

- **ハッシュアルゴリズム**: bcrypt（コスト係数12）
- **パスワードポリシー**: 最低8文字、英数字記号混在推奨
- **リセット**: Email経由のトークンベース（有効期限1時間）

### 認証不要エンドポイント

以下のエンドポイントは認証不要とする：

1. **`/api/health`**: ALBターゲットヘルスチェック用
   - レスポンス: `{"status": "healthy", "database": "connected", "timestamp": "..."}`
   - 用途: AWS ALB、外部監視システム（CloudWatch Synthetics等）

2. **`/api/auth/login`**: ログイン
3. **`/api/auth/register`**: ユーザー登録

## データベース設計原則

### 共通パターン

すべてのテーブルは以下のカラムを持つ：

```sql
id            UUID PRIMARY KEY DEFAULT gen_random_uuid()
created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
deleted_at    TIMESTAMP         -- 論理削除フラグ
```

### インデックス戦略

- **主キー**: UUID（分散ID生成）
- **外部キー**: すべて `ON DELETE RESTRICT`（誤削除防止）
- **検索頻度の高いカラム**: B-Tree インデックス
- **JSONB カラム**: GIN インデックス
- **削除フラグ**: `WHERE deleted_at IS NULL` 条件を含む部分インデックス

### マイグレーション

- **ツール**: GORM AutoMigrate（開発環境）、SQL Migration Files（本番環境）
- **ロールバック**: すべてのマイグレーションは `up` / `down` ペア
- **実行タイミング**: デプロイ前に自動実行（CI/CD）

## API設計原則

### RESTful設計

- **リソースベース**: `/api/crops`, `/api/plots`, `/api/tasks`
- **HTTPメソッド**: GET（取得）, POST（作成）, PUT（更新）, DELETE（削除）
- **ステータスコード**:
  - `200 OK`: 成功
  - `201 Created`: リソース作成成功
  - `400 Bad Request`: バリデーションエラー
  - `401 Unauthorized`: 認証エラー
  - `404 Not Found`: リソース不存在
  - `500 Internal Server Error`: サーバーエラー

### レスポンス形式

```json
{
  "data": { ... },           // 成功時のデータ
  "error": {                 // エラー時の詳細
    "code": "VALIDATION_ERROR",
    "message": "Invalid crop type",
    "details": [...]
  },
  "meta": {                  // ページネーション等
    "page": 1,
    "per_page": 20,
    "total": 150
  }
}
```

### ページネーション

- **パラメータ**: `?page=1&per_page=20`
- **デフォルト**: page=1, per_page=20
- **最大**: per_page=100

## エラーハンドリング

### ログレベル

- **ERROR**: ユーザー操作に影響するエラー（即座に対応必要）
- **WARN**: 将来エラーになる可能性のある状態（監視必要）
- **INFO**: 重要な状態変更（ユーザー作成、作物収穫等）
- **DEBUG**: 開発時のデバッグ情報（本番環境では無効化）

### エラー通知

- **Slack連携**: ERROR レベルのログを自動通知
- **CloudWatch Alarms**: エラー率が5%を超えた場合にアラート

## パフォーマンス要件

### レスポンスタイム

- **API**: P95 < 500ms, P99 < 1000ms
- **データベースクエリ**: 平均 < 100ms
- **画像アップロード**: 5MB以内、10秒以内

### スケーラビリティ

- **同時接続**: 1000ユーザー/秒を想定
- **スケーリング戦略**: ECS Auto Scaling（CPU使用率70%でスケールアウト）
- **データベース**: RDS Read Replica（読み取り負荷分散）

## テスト戦略

### バックエンド（Go）

- **単体テスト**: `testing` パッケージ、カバレッジ > 80%
- **統合テスト**: テスト用PostgreSQL（Docker Compose）
- **APIテスト**: `httptest` パッケージ
- **実行**: `go test ./...`

### フロントエンド（TypeScript）

- **単体テスト**: Jest, React Testing Library
- **E2Eテスト**: Playwright（クリティカルパスのみ）
- **ビジュアルリグレッション**: Storybook + Chromatic
- **実行**: `pnpm test`

### CI/CD

- **CI**: GitHub Actions
  - Lintチェック（golangci-lint, ESLint）
  - テスト実行
  - ビルド確認
- **CD**:
  - `main` ブランチへのマージで `stg` 環境に自動デプロイ
  - タグプッシュで `prod` 環境にデプロイ（手動承認）

## リトライ・冗長性

### AWS SDK リトライ戦略

- **標準モード**: Exponential Backoff（最大3回リトライ）
- **対象**: S3アップロード、SES Email送信、SNS Push通知
- **タイムアウト**: 10秒
- **エラー時**: CloudWatch Logsに記録、ユーザーにエラー通知

### データベース接続

- **コネクションプール**: 最大20接続
- **接続タイムアウト**: 5秒
- **リトライ**: 初回接続失敗時に3回リトライ（1秒間隔）

## 監視・アラート

### メトリクス

- **Application**: エラー率、レスポンスタイム、リクエスト数
- **Infrastructure**: CPU使用率、メモリ使用率、ディスクI/O
- **Database**: コネクション数、スロークエリ数、レプリケーション遅延

### アラート条件

- エラー率 > 5%（5分間）
- P95レスポンスタイム > 1000ms（5分間）
- CPU使用率 > 80%（10分間）
- ディスク使用率 > 85%

## セキュリティ対策

### 脆弱性対策

- **SQLインジェクション**: GORM Prepared Statement
- **XSS**: フロントエンドでのHTMLエスケープ、CSP Header
- **CSRF**: SameSite Cookie属性、CSRF Token（POST/PUT/DELETE）
- **認証**: JWT + HttpOnly Cookie
- **暗号化**: TLS 1.3（HTTPS）、パスワードbcrypt

### 定期的な対応

- **依存パッケージ更新**: 月次（Dependabot）
- **脆弱性スキャン**: 週次（Trivy、npm audit）
- **ペネトレーションテスト**: 半年ごと
