# Infrastructure

AWS インフラストラクチャ（Terraform）

## 構成

```
infrastructure/
├── versions.tf          # Terraformバージョン制約
├── provider.tf          # AWSプロバイダー設定
├── variables.tf         # 入力変数定義
├── outputs.tf           # 出力値定義
├── backend.tf           # リモートステート設定（S3 + DynamoDB）
├── vpc.tf               # VPC、サブネット、NAT Gateway
├── security_groups.tf   # ALB、ECS、RDS用セキュリティグループ
├── rds.tf               # RDS PostgreSQL + Secrets Manager
├── s3_cloudfront.tf     # S3バケット + CloudFront CDN
├── ecr_ecs.tf           # ECR + ECS Fargate + ALB
├── notifications.tf     # SNS + DynamoDB + EventBridge
└── terraform.tfvars.example  # 変数設定例
```

## アーキテクチャ

```
                    ┌─────────────────────────────────────────────────────┐
                    │                      VPC (10.0.0.0/16)              │
                    │  ┌─────────────────┐     ┌─────────────────┐        │
    Internet ───────┼──│  Public Subnet  │     │  Public Subnet  │        │
         │          │  │  (10.0.1.0/24)  │     │  (10.0.2.0/24)  │        │
         │          │  │      ┌───┐      │     │      ┌───┐      │        │
         │          │  │      │NAT│      │     │      │NAT│      │        │
         │          │  │      └───┘      │     │      └───┘      │        │
         │          │  └────────┬────────┘     └────────┬────────┘        │
         │          │           │                       │                  │
    ┌────┴────┐     │  ┌────────┴───────────────────────┴────────┐        │
    │   ALB   │─────┼──│              Private Subnets             │        │
    └────┬────┘     │  │  ┌─────────────────┐ ┌─────────────────┐ │        │
         │          │  │  │  ECS Fargate    │ │  ECS Fargate    │ │        │
         │          │  │  │  (10.0.10.0/24) │ │  (10.0.20.0/24) │ │        │
         │          │  │  └────────┬────────┘ └────────┬────────┘ │        │
         │          │  │           │                   │          │        │
         │          │  │           └─────────┬─────────┘          │        │
         │          │  │                     │                    │        │
         │          │  │              ┌──────┴──────┐             │        │
         │          │  │              │    RDS      │             │        │
         │          │  │              │ PostgreSQL  │             │        │
         │          │  │              └─────────────┘             │        │
         │          │  └──────────────────────────────────────────┘        │
         │          └─────────────────────────────────────────────────────┘
         │
    ┌────┴────┐     ┌────────────┐     ┌────────────┐
    │CloudFront│────│   S3       │     │ DynamoDB   │
    └─────────┘     │  (Images)  │     │ (Tokens)   │
                    └────────────┘     └────────────┘
```

## セットアップ

### 前提条件

- Terraform >= 1.6.0
- AWS CLI 設定済み
- 適切なIAM権限

### 1. 変数設定

```bash
cd infrastructure
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvars を編集して値を設定
```

### 2. 初期化と適用

```bash
# 初回初期化
terraform init

# 実行計画確認
terraform plan

# 適用
terraform apply
```

### 3. リモートステート移行（推奨）

初回 apply 後、S3バケットとDynamoDBテーブルが作成されます。
リモートステートに移行するには：

1. `backend.tf` のコメントアウトされた terraform backend ブロックを有効化
2. バケット名を確認して設定
3. 以下を実行:

```bash
terraform init -migrate-state
```

## 環境別設定

### 開発環境 (dev)

```hcl
environment       = "dev"
db_instance_class = "db.t3.micro"
db_multi_az       = false
ecs_task_cpu      = 256
ecs_task_memory   = 512
```

### 本番環境 (prod)

```hcl
environment       = "prod"
db_instance_class = "db.t3.small"
db_multi_az       = true
ecs_task_cpu      = 512
ecs_task_memory   = 1024
```

## コスト見積もり（開発環境）

| リソース | 月額概算（USD） |
|---------|---------------|
| NAT Gateway (x2) | ~$65 |
| RDS (db.t3.micro) | ~$15 |
| ECS Fargate | ~$10 |
| ALB | ~$20 |
| S3 + CloudFront | ~$5 |
| DynamoDB | ~$1 |
| **合計** | **~$116/月** |

※ 実際のコストは使用量により異なります。

### コスト削減のヒント

1. 開発環境ではNAT Gatewayを1つに減らす
2. FARGATE_SPOTを使用（最大70%削減）
3. RDSを停止可能に設定（8時間/日運用で66%削減）

## トラブルシューティング

### ECSタスクが起動しない

1. CloudWatch Logsを確認
2. セキュリティグループ設定を確認
3. Secrets Manager権限を確認

### RDS接続エラー

1. セキュリティグループのインバウンドルールを確認
2. Secrets Managerの認証情報を確認
3. VPCエンドポイント設定を確認

## 削除

```bash
# 開発環境の削除（force_destroy=true）
terraform destroy

# 本番環境は削除保護が有効なため、先に無効化が必要
```

## 注意事項

- 本番環境では必ず `deletion_protection = true` を維持
- Secrets Manager のシークレットは削除後30日間保持される（prod）
- S3バケットは force_destroy=false の場合、オブジェクトがあると削除できない
