# Research & Design Decisions

---

**Purpose**: 家庭菜園管理アプリの技術設計に関する調査結果、アーキテクチャ検討、設計判断の根拠を記録する。

**Usage**:

- ディスカバリーフェーズでの調査活動と成果を記録
- design.md に記載するには詳細すぎる設計判断のトレードオフを文書化
- 将来の監査や再利用のための参照資料を提供

---

## Summary

- **Feature**: `home-garden-management`
- **Discovery Scope**: 新機能（Greenfield）
- **Key Findings**:
  - Turborepo によるモノレポ管理が 2025 年のベストプラクティス（シンプル、高速、採用容易）
  - PostgreSQL はデータ整合性と複雑なクエリに優れ、長期的な拡張性で有利
  - Go Echo + GORM + Repository pattern が推奨アーキテクチャ
  - AWS ECS Fargate + Terraform によるコンテナ化デプロイがスケーラブル
  - React Native Firebase が両プラットフォームのプッシュ通知実装を統一

## Research Log

### Topic: モノレポツールの選定（Nx vs Turborepo vs pnpm workspace）

- **Context**: React Native/TypeScript/Next.js と Go/Echo の混成モノレポを効率的に管理する必要がある
- **Sources Consulted**:
  - [Setting up Turborepo with React Native and Next.js: The 2025 Production Guide](https://medium.com/better-dev-nextjs-react/setting-up-turborepo-with-react-native-and-next-js-the-2025-production-guide-690478ad75af)
  - [Nx vs. Turborepo: Integrated Ecosystem or High-Speed Task Runner?](https://dev.to/thedavestack/nx-vs-turborepo-integrated-ecosystem-or-high-speed-task-runner-the-key-decision-for-your-monorepo-279)
  - [Why I Chose Turborepo Over Nx](https://dev.to/saswatapal/why-i-chose-turborepo-over-nx-monorepo-performance-without-the-complexity-1afp)
  - [Monorepo Insights: Nx, Turborepo, and PNPM](https://www.ekino.fr/publications/monorepo-insights-nx-turborepo-and-pnpm-4-4/)
- **Findings**:
  - **Turborepo**: Rust 製、高速、10 分以内に既存プロジェクトへ導入可能、ビルド速度に特化
  - **Nx**: 統合エコシステム、多機能、複雑な学習曲線、大規模プロジェクト向け（7x 高速のベンチマーク結果もあるが小規模では逆転）
  - **pnpm workspace**: 基本的な依存関係管理のみ、ビルドツールとの組み合わせが必要
  - Expo SDK 52 がモノレポを自動検出し、Metro bundler の設定問題を解消
  - React Native は Export Maps をサポートしないため、Metro 設定の調整が必要
- **Implications**: Turborepo と pnpmWorkspace の組み合わせが、学習コストとパフォーマンスのバランスで最適。Nx は過剰スペック。

### Topic: データベース選定（PostgreSQL vs MySQL）

- **Context**: 作物、区画、タスク、ユーザーデータの永続化に RDBMS が必要
- **Sources Consulted**:
  - [PostgreSQL vs MySQL in 2025: What 10 Years Managing Both at Fortune 500 Scale Taught Me](https://medium.com/@jholt1055/postgresql-vs-mysql-in-2025-what-10-years-managing-both-at-fortune-500-scale-taught-me-adc002c10453)
  - [PostgreSQL vs MySQL - AWS Comparison](https://aws.amazon.com/compare/the-difference-between-mysql-vs-postgresql/)
  - [Why you should consider using Amazon RDS for PostgreSQL](https://www.missioncloud.com/blog/why-you-should-consider-using-amazon-rds-for-postgresql)
- **Findings**:
  - **PostgreSQL**: ACID 完全準拠、複雑なクエリに強い、JSONB 高速、GIN index による高度な検索、トランザクション整合性
  - **MySQL**: シンプル、読み取り重視ワークロードで高速、水平スケーリングが容易、InnoDB 使用時のみ ACID 準拠
  - AWS RDS で両方サポート、PostgreSQL がデフォルト選択肢に
  - 2025 年の推奨: データ集約型アプリ →PostgreSQL、シンプルな CRUD→MySQL
- **Implications**: 本アプリは分析機能（収穫量集計、グラフ可視化）を含むため、PostgreSQL が適切。将来の拡張性も考慮。

### Topic: Go Echo + JWT 認証のベストプラクティス

- **Context**: バックエンド API の認証・認可メカニズムの設計
- **Sources Consulted**:
  - [JWT Recipe | Echo Official Cookbook](https://echo.labstack.com/cookbook/jwt/)
  - [Secure REST API using JWT + echo + GORM](https://techwasti.com/building-a-secure-rest-api-with-gorm-echo-jwt-token-authentication-and-postgresql-in-go)
  - [Securing REST APIs in Go (Echo Framework Edition)](https://dev.to/lovestaco/securing-rest-apis-in-go-echo-framework-edition-259g)
- **Findings**:
  - 公式`echo-jwt`ミドルウェア（v4 対応）が利用可能
  - HS256 アルゴリズムが標準、マイクロサービスでは RS256（公開鍵署名）を検討
  - JWT はデフォルトでコンテキストの'user'キーに格納、クレームは`jwt.MapClaims`型
  - セキュリティ:
    - HttpOnly cookie でクライアント側スクリプトからの保護
    - 環境変数でシークレットキー管理（定数は禁止）
    - レート制限、入力検証、MIME sniffing/Clickjacking/XSS 対策
  - ライブラリバージョン不一致で型キャストエラーが発生する可能性
- **Implications**: echo-jwt + 環境変数管理 + HttpOnly cookie の組み合わせが推奨。マイクロサービス化を見越して RS256 も検討余地あり。

### Topic: AWS ECS Fargate + Terraform デプロイ

- **Context**: Go バックエンドのコンテナ化と AWS インフラの IaC 管理
- **Sources Consulted**:
  - [Provision AWS infrastructure using Terraform: ECS Fargate Example](https://aws.amazon.com/blogs/developer/provision-aws-infrastructure-using-terraform-by-hashicorp-an-example-of-running-amazon-ecs-tasks-on-aws-fargate/)
  - [Deploying Docker Containers to AWS ECS Using Terraform](https://earthly.dev/blog/deploy-dockcontainers-to-awsecs-using-terraform/)
  - [DevOps Automation with Terraform, Docker and AWS—Golang APIs](https://medium.com/@calvineotieno010/devops-automation-with-terraform-docker-and-aws-implementing-a-complete-terraform-workflow-with-11090da023be)
- **Findings**:
  - ECS Fargate はサーバーレスコンピュートエンジン、インフラ管理不要
  - Terraform で以下をコード管理: ECR, ECS Cluster, Task Definition, Service, ALB, VPC, RDS, S3
  - CI/CD パイプライン: GitHub Actions → Docker Build → ECR Push → ECS Deploy
  - Go アプリのマルチステージビルドでイメージサイズ削減
  - ローリングアップデートによる無停止デプロイ
- **Implications**: Terraform + ECS Fargate がスケーラブルかつメンテナブル。GitHub Actions で CI/CD を自動化。

### Topic: React Native プッシュ通知実装（FCM + APNS）

- **Context**: モバイルアプリへのタスクリマインダーと収穫通知の配信
- **Sources Consulted**:
  - [React Native Firebase - Cloud Messaging](https://rnfirebase.io/messaging/usage)
  - [Mastering Push Notifications in React Native](https://medium.com/@bhuvin25/mastering-push-notifications-in-react-native-a-comprehensive-guide-with-examples-d0a8c24c9ab6)
  - [AWS SNS Mobile Push Notifications](https://docs.aws.amazon.com/sns/latest/dg/sns-mobile-application-as-subscriber.html)
- **Findings**:
  - **@react-native-firebase/messaging**が両プラットフォーム統一 API
  - FCM（Android）+ APNS（iOS）のネイティブ統合
  - デバイストークン管理: ログイン/アプリ起動時に登録、DynamoDB で永続化
  - iOS: Xcode で"Push Notifications"と"Background Modes"を有効化、APNs キー登録
  - Android: Firebase Console で FCM 設定
  - data-only メッセージ: 優先度 high（Android）、content-available: true（iOS）
  - バックグラウンドハンドラー: `setBackgroundMessageHandler`で登録
  - AWS SNS 経由でプッシュ配信（API Gateway → Lambda → SNS → FCM/APNS）
- **Implications**: React Native Firebase + AWS SNS アーキテクチャが推奨。サーバーレス構成でスケーラブル。

### Topic: Go Echo + GORM + Repository Pattern

- **Context**: バックエンドのレイヤードアーキテクチャ設計
- **Sources Consulted**:
  - [The Repository pattern in Go | Three Dots Labs](https://threedots.tech/post/repository-pattern-in-go/)
  - [PostgreSQL in Go: Repository pattern | gosamples.dev](https://gosamples.dev/postgresql-intro/repository-pattern/)
  - [GitHub - johnnyaustor/golang-skeleton](https://github.com/johnnyaustor/golang-skeleton)
- **Findings**:
  - Repository pattern でデータレイヤーとビジネスロジックを分離
  - 構造: Handler → Service → Repository → Database
  - GORM の gorm.DB は並行安全な接続を表現
  - ベストプラクティス: インターフェース定義でモック可能、テスト容易性向上
  - トランザクション管理は Service レイヤーで実施
- **Implications**: Clean Architecture ベースの Layered Architecture 採用。Repository pattern で疎結合を実現。

## Architecture Pattern Evaluation

| Option                       | Description                                       | Strengths                      | Risks / Limitations      | Notes                              |
| ---------------------------- | ------------------------------------------------- | ------------------------------ | ------------------------ | ---------------------------------- |
| Clean Architecture (Layered) | Repository pattern + Service layer + Domain model | テスタビリティ、疎結合、保守性 | 小規模には過剰な可能性   | Go Echo のベストプラクティスに合致 |
| MVC                          | Model-View-Controller                             | シンプル、学習容易             | ビジネスロジックが肥大化 | Go では非推奨                      |
| Hexagonal (Ports & Adapters) | Core domain + Adapter 抽象化                      | 外部依存の切り替え容易         | Adapter 層の構築コスト   | 将来的なマイクロサービス化に有利   |
| Event-Driven                 | イベントソーシング + CQRS                         | 高スケーラビリティ、監査ログ   | 複雑性、学習コスト       | 現段階では過剰                     |

**選定**: Clean Architecture ベースの Layered Architecture（Repository pattern）を採用。理由: Go エコシステムで標準的、テスト容易性、将来のマイクロサービス化への移行パスが明確。

## Design Decisions

### Decision: Turborepo + pnpm Workspace によるモノレポ管理

- **Context**: React Native/TypeScript/Next.js フロントエンドと Go/Echo バックエンドの統合管理
- **Alternatives Considered**:
  1. **Nx** — 多機能統合エコシステム、高度なツール群
  2. **Lerna + Yarn Workspaces** — レガシーだが実績あり
  3. **pnpm Workspace 単体** — 依存関係管理のみ、ビルドツール不在
- **Selected Approach**: Turborepo + pnpm Workspace
- **Rationale**:
  - Turborepo は 10 分で既存プロジェクトに導入可能、学習コスト低
  - Rust 製で高速、キャッシュ機構が強力
  - pnpm は高速かつディスク効率的
  - Nx は機能過多で学習曲線が急峻
- **Trade-offs**: Nx の高度な機能（依存関係グラフ可視化、影響分析）は失う。プロジェクト規模では不要と判断。
- **Follow-up**: CI/CD で Turborepo のキャッシュ機構を活用しビルド時間を短縮

### Decision: PostgreSQL on AWS RDS

- **Context**: リレーショナルデータベースの選定
- **Alternatives Considered**:
  1. **MySQL** — シンプル、読み取り高速
  2. **MongoDB** — NoSQL、スキーマレス
- **Selected Approach**: PostgreSQL
- **Rationale**:
  - 要件 4（記録・分析）で集計クエリ、グラフ可視化が必要
  - PostgreSQL は JSONB、GIN インデックスで柔軟な検索が可能
  - ACID 完全準拠でデータ整合性が保証される
  - AWS RDS で完全マネージド、自動バックアップ
- **Trade-offs**: MySQL より複雑、初期学習コスト若干高
- **Follow-up**: パフォーマンスチューニング（インデックス、EXPLAIN ANALYZE）を実装フェーズで実施

### Decision: JWT + HttpOnly Cookie 認証

- **Context**: ユーザー認証・セッション管理の実装
- **Alternatives Considered**:
  1. **Session-based (Redis)** — サーバー側ステート管理
  2. **OAuth2 (Cognito)** — 外部 IdP 統合
- **Selected Approach**: JWT + HttpOnly Cookie
- **Rationale**:
  - ステートレス、水平スケーリング容易
  - echo-jwt 公式ミドルウェアで実装が簡潔
  - HttpOnly cookie で XSS 攻撃を防御
  - 環境変数でシークレットキー管理（AWS Secrets Manager）
- **Trade-offs**: トークンの無効化が困難（ブラックリストが必要）
- **Follow-up**: リフレッシュトークン機構を検討、トークン有効期限を短く設定

### Decision: AWS SNS + React Native Firebase for Push Notifications

- **Context**: モバイルプッシュ通知とメール通知の実装
- **Alternatives Considered**:
  1. **OneSignal** — サードパーティサービス
  2. **FCM 直接統合** — AWS SNS 不使用
- **Selected Approach**: AWS SNS + React Native Firebase
- **Rationale**:
  - AWS SNS でプッシュ（FCM/APNS）とメール（SES）を統一
  - React Native Firebase が両プラットフォーム統一 API 提供
  - サーバーレス（Lambda + SNS）でスケーラブル
  - DynamoDB でデバイストークン永続化
- **Trade-offs**: Web Push 非対応（別途実装必要）
- **Follow-up**: Web Push は Service Worker + Web Push API で別途実装検討

## Risks & Mitigations

- **Risk 1: React Native と Next.js の依存関係衝突** — Mitigation: pnpm workspace で厳密なバージョン管理、定期的な依存関係監査
- **Risk 2: ECS Fargate のコールドスタート遅延** — Mitigation: 最小タスク数を 1 に設定、ALB ヘルスチェック最適化
- **Risk 3: PostgreSQL のパフォーマンスボトルネック** — Mitigation: インデックス戦略、EXPLAIN ANALYZE でクエリ最適化、必要に応じて Read Replica 導入
- **Risk 4: JWT トークン漏洩** — Mitigation: HttpOnly cookie、HTTPS 強制、短期有効期限、リフレッシュトークン機構
- **Risk 5: モノレポのビルド時間増大** — Mitigation: Turborepo のキャッシュ活用、影響範囲のみビルド、並列実行

## References

### Monorepo & Build Tools

- [Setting up Turborepo with React Native and Next.js: The 2025 Production Guide](https://medium.com/better-dev-nextjs-react/setting-up-turborepo-with-react-native-and-next-js-the-2025-production-guide-690478ad75af)
- [Nx vs. Turborepo Comparison](https://dev.to/thedavestack/nx-vs-turborepo-integrated-ecosystem-or-high-speed-task-runner-the-key-decision-for-your-monorepo-279)
- [React Native + Next.js Monorepo](https://ecklf.com/blog/rn-monorepo)

### Database

- [PostgreSQL vs MySQL in 2025](https://medium.com/@jholt1055/postgresql-vs-mysql-in-2025-what-10-years-managing-both-at-fortune-500-scale-taught-me-adc002c10453)
- [AWS RDS PostgreSQL](https://www.missioncloud.com/blog/why-you-should-consider-using-amazon-rds-for-postgresql)

### Backend (Go Echo)

- [Echo JWT Cookbook](https://echo.labstack.com/cookbook/jwt/)
- [Secure REST API with Echo + JWT + GORM](https://techwasti.com/building-a-secure-rest-api-with-gorm-echo-jwt-token-authentication-and-postgresql-in-go)
- [Repository Pattern in Go](https://threedots.tech/post/repository-pattern-in-go/)

### Infrastructure

- [AWS ECS Fargate + Terraform](https://aws.amazon.com/blogs/developer/provision-aws-infrastructure-using-terraform-by-hashicorp-an-example-of-running-amazon-ecs-tasks-on-aws-fargate/)
- [Deploying Docker to ECS with Terraform](https://earthly.dev/blog/deploy-dockcontainers-to-awsecs-using-terraform/)

### Notifications

- [AWS SNS Mobile Push](https://docs.aws.amazon.com/sns/latest/dg/sns-mobile-application-as-subscriber.html)
- [React Native Firebase Messaging](https://rnfirebase.io/messaging/usage)
- [Mastering Push Notifications in React Native](https://medium.com/@bhuvin25/mastering-push-notifications-in-react-native-a-comprehensive-guide-with-examples-d0a8c24c9ab6)
