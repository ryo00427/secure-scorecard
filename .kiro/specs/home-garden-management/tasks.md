# Implementation Tasks

## 1. プロジェクト基盤・モノレポ構成

- [x] 1.1 (P) モノレポ初期化とTurborepo設定
  - pnpm workspaceでモノレポルート初期化
  - Turborepo設定ファイル作成（turbo.json、v2 tasks形式）
  - apps/mobile、apps/backend、packages/shared、packages/typescript-configディレクトリ作成
  - ルートpackage.jsonにワークスペース設定追加（apps/*, packages/*）
  - ビルドキャッシュ設定とタスク定義
  - _Requirements: 7.1, 7.2, 7.3_

- [x] 1.2 (P) 共有TypeScript設定とリンター設定
  - packages/shared配下に共通型定義とユーティリティ配置
  - ESLint、Prettier設定をモノレポ全体で統一
  - TypeScript共通設定（tsconfig.json）作成
  - gofmt設定（バックエンド用）
  - パッケージ間の型共有設定
  - _Requirements: 7.2, 7.6_

## 2. AWS インフラストラクチャ（Terraform）

- [x] 2.1 (P) Terraform基盤とVPC構成
  - Terraformプロジェクト初期化（infrastructure/ディレクトリ）
  - AWS Provider設定とバックエンド（S3 + DynamoDB）設定
  - VPC、サブネット（Public/Private）、Internet Gateway作成
  - セキュリティグループ定義（ALB, ECS, RDS用）
  - NAT Gateway設定
  - _Requirements: 7.3, 7.7_

- [x] 2.2 RDS PostgreSQL構成
  - RDS PostgreSQL 16.x インスタンス定義
  - Multi-AZ構成設定
  - データベース暗号化（AES-256）設定
  - セキュリティグループでアクセス制御
  - Secrets ManagerでDB認証情報管理
  - _Requirements: 7.3, 7.7_

- [x] 2.3 (P) S3とCloudFront構成
  - S3バケット作成（画像保存用）
  - S3暗号化（SSE-S3）設定
  - CloudFront Distribution設定（画像CDN）
  - CORS設定（フロントエンドアクセス許可）
  - ライフサイクルポリシー設定
  - _Requirements: 7.3_

- [x] 2.4 (P) ECRとECS Fargate構成
  - ECRリポジトリ作成（backend用）
  - ECS Cluster作成
  - Task Definition定義（CPU, メモリ、環境変数）
  - ECS Service設定（Auto Scaling、Health Check）
  - ALB設定（ターゲットグループ、リスナー）
  - _Requirements: 7.4_

- [x] 2.5 (P) SNS/SES/DynamoDB構成
  - SNS Topic作成（プッシュ通知用）
  - SNS Platform Application設定（FCM, APNS）
  - SES設定（メール通知用）
  - DynamoDBテーブル作成（デバイストークン保存用）
  - EventBridge Scheduler設定（Daily cron job）
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 7.3_

## 3. データベース設計（PostgreSQL）

- [x] 3.1 データベーススキーマ設計と作成
  - usersテーブル作成（email, password_hash, notification_settings）
  - cropsテーブル作成（name, variety, planted_date, expected_harvest_date, status）
  - growth_recordsテーブル作成（crop_id, record_date, growth_stage, notes, image_url）
  - harvestsテーブル作成（crop_id, harvest_date, quantity, quality）
  - plotsテーブル作成（name, width, height, soil_type, sunlight, status）
  - plot_assignmentsテーブル作成（plot_id, crop_id, assigned_date）
  - tasksテーブル作成（name, due_date, priority, status, recurrence）
  - token_blacklistテーブル作成（token_hash, revoked_at, expires_at）
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 3.1, 6.1, 6.3, 6.5_

- [x] 3.2 インデックスと制約設定
  - 各テーブルに適切なインデックス作成（user_id, status, due_date, expires_at等）
  - 外部キー制約設定（ON DELETE CASCADE, SET NULL）
  - CHECK制約設定（valid_dates等）
  - UNIQUE制約設定（email, plot_id等）
  - 論理削除用のdeleted_atカラム追加
  - token_blacklist期限切れトークン定期削除ジョブ設定（Daily cron）
  - _Requirements: 6.3, 6.5_

- [x] 3.3 Materialized Viewと分析用インデックス
  - mv_harvest_analyticsマテリアライズドビュー作成
  - 集計クエリ高速化用インデックス作成
  - 定期リフレッシュジョブ設定（毎日深夜）
  - パフォーマンスチューニング（EXPLAIN ANALYZE実行）
  - _Requirements: 4.1, 4.2, 4.4_

## 4. バックエンド - Go Echo基盤

- [x] 4.1 (P) Goプロジェクト構造とEcho初期化
  - apps/backendディレクトリにGoモジュール初期化
  - Echo v4、GORM v2、echo-jwt v4依存関係追加
  - Clean Architectureベースのディレクトリ構造構築（handler, service, repository, model）
  - main.goとサーバー起動処理実装
  - 環境変数読み込み設定（AWS Secrets Manager統合）
  - _Requirements: 7.1, 7.7_

- [x] 4.2 (P) データベース接続とRepository基盤
  - GORM PostgreSQL接続設定
  - Repository パターン基盤実装（interface定義）
  - トランザクション管理機構実装
  - 接続プーリング設定
  - マイグレーション管理設定
  - _Requirements: 6.3_

- [x] 4.3 (P) JWT認証ミドルウェア設定
  - echo-jwt v4ミドルウェア統合
  - JWT生成・検証ロジック実装（HS256）
  - HttpOnly Cookie設定（SameSite=Strict）
  - 認証エラーハンドリング
  - `/api/health`ヘルスチェックエンドポイント実装（認証不要）
  - _Requirements: 6.2, 6.4, 6.5_

- [x] 4.4 (P) エラーハンドリングとバリデーション基盤
  - 統一エラーレスポンス形式実装（JSON）
  - カスタムエラー型定義（ValidationError, AuthError等）
  - 入力バリデーションミドルウェア実装
  - ロギング設定（構造化ログ、CloudWatch Logs統合）
  - CORS設定（許可オリジン設定）
  - _Requirements: 6.2, 6.6_

## 5. バックエンド - AuthService実装

- [x] 5.1 AuthService - ユーザー登録機能
  - `POST /api/auth/register` エンドポイント実装
  - メールアドレス・パスワードバリデーション（RFC 5322、最小8文字）
  - パスワードハッシュ化（bcrypt, cost 12）
  - ユーザーテーブルへの挿入
  - JWT トークン生成と返却
  - 重複メールアドレスチェック（409エラー）
  - _Requirements: 6.1, 6.4_

- [x] 5.2 AuthService - ログイン機能
  - `POST /api/auth/login` エンドポイント実装
  - メールアドレス・パスワード検証
  - bcryptでパスワード照合
  - ログイン失敗回数カウント（increment_failed_login）
  - 3回失敗でアカウント30分ロック（locked_until設定）
  - 成功時にJWT発行とCookie設定
  - _Requirements: 6.2, 6.4, 6.6_

- [x] 5.3 (P) AuthService - ログアウトと認証状態確認
  - `POST /api/auth/logout` エンドポイント実装（Cookie削除）
  - `GET /api/auth/me` エンドポイント実装（現在のユーザー情報返却）
  - JWT検証ミドルウェア適用
  - セッション無効化処理
  - _Requirements: 6.5_

- [x] 5.4 AuthService - JWT トークンブラックリスト実装
  - ログアウト時にトークンハッシュ（SHA-256）をtoken_blacklistテーブルに保存
  - JWT検証ミドルウェアでブラックリストチェック追加（トークン検証前）
  - ブラックリスト登録済みトークンは401 Unauthorizedで拒否
  - expires_at設定（トークン有効期限と同じ24時間）
  - トークンハッシュ生成処理実装（crypto/sha256使用）
  - _Requirements: 6.5, 6.6_

- [x] 5.5* AuthService - ユニットテスト
  - register正常系テスト（トークン発行確認）
  - register異常系テスト（重複メール、弱いパスワード）
  - login正常系テスト（JWT発行確認）
  - login異常系テスト（無効なパスワード、アカウントロック）
  - _Requirements: 6.1, 6.2, 6.6_

## 6. バックエンド - CropService実装

- [x] 6.1 (P) CropService - 作物CRUD実装
  - `POST /api/crops` 作物新規登録エンドポイント
  - `GET /api/crops/:id` 作物詳細取得エンドポイント
  - `GET /api/crops` 作物一覧取得エンドポイント（フィルター対応）
  - Repository層実装（GORM CRUD操作）
  - ユーザーIDによるデータ分離（Row-level security）
  - plantedDate <= expectedHarvestDate バリデーション
  - _Requirements: 1.1, 6.3_

- [x] 6.2 CropService - 成長記録機能
  - `POST /api/crops/:id/records` 成長記録追加エンドポイント
  - `GET /api/crops/:id/records` 成長記録一覧取得エンドポイント（時系列ソート）
  - growth_recordsテーブルへの挿入
  - 成長段階（Seedling, Vegetative, Flowering, Fruiting）バリデーション
  - 画像URL保存（S3署名付きURL連携）
  - _Requirements: 1.2, 1.4, 4.3_

- [x] 6.3 (P) CropService - 収穫記録機能
  - `POST /api/crops/:id/harvest` 収穫記録エンドポイント
  - harvestsテーブルへの挿入
  - 作物ステータスを"Harvested"に更新
  - 収穫量・品質評価バリデーション
  - 収穫7日前の通知イベント発行（NotificationServiceへ）
  - _Requirements: 1.3, 1.5_

- [x] 6.4 (P) CropService - 画像アップロード機能
  - `POST /api/crops/images` 画像アップロードエンドポイント
  - S3署名付きURL生成（presigned URL, 15分有効）
  - 画像ファイルサイズ・形式バリデーション（5MB上限、JPEG/PNG/WEBP）
  - CloudFront CDN統合
  - multipart/form-data処理
  - S3アップロード失敗時のExponential backoffリトライ機構（初回1秒、最大3回）
  - _Requirements: 1.2_

- [x] 6.5* CropService - ユニットテスト
  - createCrop正常系テスト
  - addGrowthRecord正常系テスト（画像URLバリデーション含む）
  - recordHarvest正常系・異常系テスト（日付バリデーション）
  - _Requirements: 1.1, 1.2, 1.3_

## 7. バックエンド - PlotService実装

- [x] 7.1 (P) PlotService - 区画CRUD実装
  - `POST /api/plots` 区画新規登録エンドポイント
  - `GET /api/plots/:id` 区画詳細取得エンドポイント
  - `GET /api/plots` 区画一覧取得エンドポイント
  - Repository層実装（GORM CRUD操作）
  - width > 0, height > 0 バリデーション
  - 土壌タイプ・日当たり条件のEnum検証
  - _Requirements: 2.1_

- [x] 7.2 PlotService - 作物配置機能
  - `POST /api/plots/:id/assign` 作物配置エンドポイント
  - `DELETE /api/plots/:id/assign` 作物配置解除エンドポイント
  - plot_assignmentsテーブル操作
  - 区画ステータス更新（Available ⇔ Occupied）
  - 重複配置チェック（PlotAssignment unique constraint検証）
  - 配置履歴記録
  - _Requirements: 2.2, 2.5_

- [x] 7.3 (P) PlotService - レイアウトと履歴機能
  - `GET /api/plots/layout` レイアウト取得エンドポイント（グリッド表示用）
  - `GET /api/plots/:id/history` 区画使用履歴取得エンドポイント
  - 区画位置情報（positionX, positionY）を含むレイアウトデータ生成
  - 過去の栽培作物履歴クエリ
  - _Requirements: 2.3, 2.4_

- [x] 7.4* PlotService - ユニットテスト
  - createPlot正常系テスト
  - assignCrop正常系テスト（重複チェック含む）
  - assignCrop異常系テスト（PlotOccupiedエラー）
  - _Requirements: 2.1, 2.2, 2.5_

## 8. バックエンド - TaskService実装

- [x] 8.1 (P) TaskService - タスクCRUD実装
  - `POST /api/tasks` タスク作成エンドポイント
  - `GET /api/tasks` タスク一覧取得エンドポイント（フィルター・ソート対応）
  - `GET /api/tasks/today` 今日のタスク取得エンドポイント
  - `GET /api/tasks/overdue` 期限切れタスク取得エンドポイント
  - Repository層実装（GORM CRUD操作）
  - dueDate >= today バリデーション
  - _Requirements: 3.1, 3.2, 3.4_

- [x] 8.2 TaskService - タスク完了と繰り返し生成
  - `PUT /api/tasks/:id/complete` タスク完了エンドポイント
  - タスクステータスを"Completed"に更新
  - 繰り返し設定（recurrence）がある場合、次回タスク自動生成
  - 繰り返し頻度（Daily, Weekly, Monthly）とinterval処理
  - 繰り返し終了条件チェック（maxOccurrences または endDate到達時は生成停止）
  - タスク作成時にendCondition必須バリデーション追加
  - _Requirements: 3.3, 3.6_

- [x] 8.3 (P) TaskService - 期限切れ検出と通知連携
  - 期限切れタスク検出ロジック実装（due_date < today and status = 'Pending'）
  - 期限切れタスク3件以上で警告イベント発行（NotificationServiceへ）
  - AWS EventBridge Schedulerとの統合（Daily cron job）
  - 当日タスクの通知イベント発行
  - _Requirements: 3.5, 5.1, 5.5_

- [x] 8.4* TaskService - ユニットテスト
  - createTask正常系テスト
  - completeTask正常系テスト（繰り返しタスク自動生成確認）
  - getOverdueTasks正常系テスト
  - _Requirements: 3.1, 3.3, 3.6_

## 9. バックエンド - AnalyticsService実装

- [x] 9.1 (P) AnalyticsService - 収穫量集計機能
  - `GET /api/analytics/harvest` 収穫量集計エンドポイント
  - 作物ごとの総収穫量・平均成長期間を集計
  - フィルター対応（startDate, endDate, cropID）
  - Materialized View活用（mv_harvest_analytics）
  - _Requirements: 4.1_

- [x] 9.2 (P) AnalyticsService - グラフデータ生成機能
  - `GET /api/analytics/charts/:type` グラフデータ取得エンドポイント
  - 月別・作物別の収穫量データ生成（MonthlyHarvest, CropComparison）
  - 区画生産性データ生成（PlotProductivity）
  - フロントエンド用ChartDataフォーマット変換
  - _Requirements: 4.2, 4.4_

- [x] 9.3 (P) AnalyticsService - CSVエクスポート機能
  - `GET /api/analytics/export/:dataType` CSVエクスポートエンドポイント
  - データタイプ別エクスポート（Crops, Harvests, Tasks, All）
  - CSV生成処理（encoding/csv使用）
  - S3へのアップロードと署名付きURL生成（1時間有効）
  - _Requirements: 4.5_

- [x] 9.4* AnalyticsService - ユニットテスト
  - getHarvestSummary正常系テスト（フィルター適用確認）
  - getChartData正常系テスト（データフォーマット確認）
  - exportCSV正常系テスト（CSV生成確認）
  - _Requirements: 4.1, 4.2, 4.5_

## 10. バックエンド - NotificationService実装

- [x] 10.1 (P) NotificationService - デバイストークン管理
  - デバイストークン登録ロジック実装（PostgreSQL保存、DynamoDB移行可能）
  - デバイストークン削除ロジック実装（無効トークン対応）
  - (userID, platform) 複合キーでの管理
  - トークン有効性検証
  - `POST /api/v1/notifications/device-token` 登録エンドポイント
  - `DELETE /api/v1/notifications/device-token` 削除エンドポイント
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 10.2 NotificationService - プッシュ通知配信
  - AWS SNS統合（notification/sender.go）
  - FCM/APNS向けメッセージフォーマット変換（PushMessage構造体）
  - data-onlyメッセージ設定（priority: high）
  - ユーザー通知設定反映（pushEnabled確認）
  - 重複通知防止（notification_logsテーブル + deduplication_key、TTL 24時間）
  - SNS配信失敗時のExponential backoffリトライ機構（初回100ms、最大3回）
  - 無効デバイストークン検出時の自動削除処理（DeactivateToken）
  - _Requirements: 5.1, 5.2, 5.3, 5.5_

- [x] 10.3 (P) NotificationService - メール通知配信
  - AWS SES統合（notification/sender.go）
  - トランザクションメール送信（テンプレート使用: TaskReminder, OverdueAlert, HarvestReminder）
  - ユーザー通知設定反映（emailEnabled確認）
  - SES送信失敗時のExponential backoffリトライ機構（初回100ms、最大3回）
  - EmailMessage構造体によるHTML/TextBody対応
  - _Requirements: 5.1, 5.2, 6.6_

- [x] 10.4 NotificationService - イベント購読実装
  - notification/processor.go でイベント処理実装
  - CropHarvestReminderイベント処理（harvest_reminder）
  - TaskDueReminderイベント処理（task_due_reminder）
  - TaskOverdueAlertイベント処理（task_overdue_alert）
  - GrowthRecordAddedイベント処理（growth_record_added、オプション）
  - ProcessEvents メソッドで一括処理
  - _Requirements: 1.5, 5.1, 5.2, 5.3, 5.5_

- [x] 10.5 (P) NotificationService - 通知設定カスタマイズ
  - `GET /api/v1/users/settings/notifications` 通知設定取得エンドポイント
  - `PUT /api/v1/users/settings/notifications` 通知設定更新エンドポイント
  - notification_settings JSONB更新
  - 設定項目（pushEnabled, emailEnabled, taskReminders, harvestReminders, growthRecordNotifications）
  - _Requirements: 5.4_

- [ ] 10.6* NotificationService - 統合テスト
  - デバイストークン登録→プッシュ通知配信フロー確認
  - イベント発行→通知配信フロー確認（Cron job実行含む）
  - ユーザー設定による通知スキップ確認
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

## 11. Dockerコンテナ化

- [ ] 11.1 Goバックエンド Dockerfile作成
  - マルチステージビルド設定（ビルド + 実行ステージ）
  - 依存関係キャッシュ最適化
  - 最小イメージサイズ構成（alpine base）
  - ヘルスチェック設定
  - _Requirements: 7.4_

- [ ] 11.2 docker-compose開発環境設定
  - PostgreSQL、backend、redisコンテナ定義
  - ネットワーク設定
  - ボリュームマウント設定
  - 環境変数管理（.env）
  - _Requirements: 7.4_

## 12. フロントエンド - React Native Mobile実装

- [ ] 12.1 (P) React Native プロジェクト初期化
  - apps/mobileディレクトリにExpo初期化
  - Expo Managed Workflow設定
  - React Navigation設定
  - TypeScript設定
  - React Query（TanStack Query）セットアップ
  - _Requirements: 7.1_

- [ ] 12.2 (P) React Native Firebase統合
  - @react-native-firebase/app インストール
  - @react-native-firebase/messaging インストール
  - Firebase設定ファイル追加（google-services.json, GoogleService-Info.plist）
  - iOS: Xcode Push Notifications有効化、Background Modes設定
  - Android: Firebase Console FCM設定
  - デバイストークン取得・登録処理実装
  - _Requirements: 5.1, 5.2, 5.3_

- [ ] 12.3 プッシュ通知ハンドラー実装
  - フォアグラウンド通知ハンドラー実装
  - バックグラウンド通知ハンドラー実装（setBackgroundMessageHandler）
  - 通知タップ時のナビゲーション処理
  - 通知パーミッション取得（初回起動時）
  - _Requirements: 5.1, 5.2, 5.3_

- [ ] 12.4 認証画面実装
  - ログイン画面実装
  - ユーザー登録画面実装
  - AsyncStorageで認証状態管理
  - APIクライアント実装（Axios + React Query）
  - JWT Cookie管理（httpOnly対応）
  - _Requirements: 6.1, 6.2, 6.5_

- [ ] 12.5 作物・区画・タスク画面実装
  - 作物一覧・詳細画面実装
  - 作物登録・成長記録追加画面実装（画像アップロード含む）
  - 区画一覧・レイアウト表示画面実装
  - タスク一覧・今日のタスク画面実装
  - タスク作成・完了機能実装
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.3, 3.1, 3.2, 3.3_

- [ ] 12.6 分析画面実装
  - ダッシュボード画面実装
  - グラフコンポーネント実装（react-native-chart-kit等使用）
  - 収穫量集計・グラフ表示
  - CSVエクスポート機能実装（Share APIで共有）
  - _Requirements: 4.1, 4.2, 4.4, 4.5_

- [ ] 12.7* Mobile UIテスト
  - 主要画面のレンダリングテスト（Jest + React Native Testing Library）
  - ナビゲーションテスト
  - プッシュ通知受信テスト（モック）
  - _Requirements: 1.1, 5.1, 6.1_

## 14. 統合テスト

- [ ] 14.1 バックエンド統合テスト
  - Auth Flow統合テスト（登録→ログイン→JWT検証→保護API）
  - Crop Lifecycle統合テスト（作物登録→成長記録→収穫→通知）
  - Task Notification統合テスト（タスク作成→Cron job→通知配信）
  - Plot Assignment統合テスト（区画作成→配置→重複エラー）
  - CSV Export統合テスト（データ生成→S3→署名付きURL）
  - _Requirements: 1.1, 1.2, 1.3, 1.5, 2.1, 2.2, 3.1, 4.5, 5.1, 6.1, 6.2_

- [ ] 14.2 E2Eテスト（Playwright）
  - ユーザー登録・ログインフロー
  - 作物管理フロー（登録→成長記録→画像アップロード→収穫）
  - 区画レイアウトフロー（区画作成→作物配置→レイアウト保存）
  - タスク管理フロー（タスク作成→今日のタスク表示→完了）
  - 分析ダッシュボードフロー（グラフ表示→CSVエクスポート）
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1, 4.2, 4.5_

- [ ] 14.3* パフォーマンステスト
  - API Throughput テスト（1000 req/sec、レイテンシ < 200ms）
  - データベースクエリパフォーマンステスト（集計クエリ < 1秒）
  - 画像アップロードパフォーマンステスト（5MB画像 < 3秒）
  - 同時アクセステスト（100ユーザー、競合エラーなし）
  - _Requirements: 3.1, 4.1, 4.2_

## 15. CI/CDパイプライン構築

- [ ] 15.1 GitHub Actions ワークフロー設定
  - Turborepoキャッシュ統合
  - リンター・フォーマッター実行（ESLint, Prettier, gofmt）
  - 全パッケージのビルド・テスト自動実行
  - PR作成時の自動チェック
  - _Requirements: 7.5, 7.6_

- [ ] 15.2 Docker Build & ECR Push設定
  - Dockerfile最適化（マルチステージビルド）
  - GitHub ActionsでDocker Build実行
  - ECR Pushワークフロー設定
  - イメージタグ管理（commit SHA、latest）
  - _Requirements: 7.4, 7.5_

- [ ] 15.3 ECS Deploy設定
  - ECS Task Definition更新
  - ECS Serviceローリングアップデート設定
  - デプロイ検証（Health Check確認）
  - ロールバック戦略設定
  - _Requirements: 7.4, 7.5_

- [ ] 15.4* CI/CDパイプライン検証
  - PR作成→ビルド・テスト→レビュー承認→マージ→デプロイフロー確認
  - ビルドキャッシュ効果確認（Turborepo）
  - デプロイ成功・失敗時の動作確認
  - _Requirements: 7.5_
