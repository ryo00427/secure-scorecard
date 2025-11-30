家庭菜園アプリのモノレポ構成メモ（全体像＋技術の役割を固定するためのたたき台）。

1. ユースケース/機能
- ユーザー登録/ログイン
- 畑・プランター管理（名前、場所、日当たりメモ）
- 作物管理（品種、種まき日、収穫予定）
- ケアログ（水やり/肥料/防除/収穫など）、写真アップロード
- カレンダー/タイムライン表示、リマインド（例: 水やりタイミング、収穫目安）

2. モノレポ構成（予定）
- /frontend: Next.js + TypeScript + Tailwind
- /backend: Go + Echo（REST/JSON API）
- /shared: API スキーマや共通型（OpenAPI/JSON Schema/TS 型など）
- /infra: docker-compose、IaC（必要なら Terraform/CloudFormation）
- /scripts: 開発/CI 用スクリプト
- /docs: 要件・設計ドキュメント
- ルートに Makefile/justfile, .editorconfig, lint 設定、GitHub Actions を配置

3. 技術スタック
- フロント: Next.js App Router, fetch/axios/react-query/SWR, Tailwind UI
- モバイル: React Native + Expo（後から展開）
- バックエンド: Go + Echo, RDB アクセスは GORM or database/sql + sqlc
- インフラ (AWS): ECS(Fargate) or Lightsail/EKS、RDS for PostgreSQL、S3（写真）、ALB、CloudFront(任意)
- CI/CD: GitHub Actions（lint/test/build/ECR push/ECS deploy）

4. DB 設計（PostgreSQL 想定）
- users: id, email(unique), password_hash, name, created_at, updated_at
- gardens: id, user_id(FK), name, location_text, memo, created_at, updated_at
- plants: id, garden_id(FK), name, variety, sowing_date, planting_date, expected_harvest_start_date, expected_harvest_end_date, memo, created_at, updated_at
- care_logs: id, plant_id(FK), log_type(watering/fertilizer/harvest/pruning/etc), amount, memo, logged_at, created_at, updated_at
- photos: id, plant_id(FK), s3_key, taken_at, created_at, updated_at

5. バックエンド構成（例）
- backend/cmd/api/main.go
- backend/internal/config
- backend/internal/http/{handlers,middleware,router.go}
- backend/internal/domain/{model,repository,service}
- backend/pkg/{logger,util}

6. フロントエンド構成（例）
- ページ: /, /gardens, /gardens/[id], /plants/[id], /calendar
- UI: Tailwind カードレイアウト（プランターカード/作物カード）

7. Docker/ローカル開発
- backend Dockerfile: Go ビルド → distroless 実行
- frontend Dockerfile: Next.js 2 段階ビルド
- infra/docker-compose.yml: db(postgres:16), backend, frontend
- ルートの make dev などで一括起動する想定

8. CI/CD（GitHub Actions 例）
 - PR: frontend → lint/eslint + typecheck(tsc) + unit test、backend → go test ./...
 - main マージ: Docker build → ECR push → ECS 更新
 - ワークフロー例: .github/workflows/frontend.yml, backend.yml（paths フィルタで対象だけ動かす）

9. ER 図
- docs/ERD.md にテキスト版を配置

10. アーキテクチャ図
- docs/ARCHITECTURE.md に Mermaid 図を配置
