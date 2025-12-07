# Home Garden Management - AWS Infrastructure

Terraform を使用した AWS インフラストラクチャ定義です。

## 構成

```
infrastructure/
├── main.tf                 # メインモジュール定義
├── variables.tf            # 変数定義
├── outputs.tf              # 出力定義
├── versions.tf             # Terraform/プロバイダーバージョン
├── terraform.tfvars.example # 変数設定サンプル
├── modules/                # 再利用可能なモジュール
│   ├── vpc/                # VPC、サブネット、セキュリティグループ
│   ├── rds/                # PostgreSQL RDS
│   ├── s3/                 # S3バケット、CloudFront
│   ├── ecs/                # ECR、ECS Fargate、ALB
│   └── notifications/      # SNS、SES、DynamoDB、EventBridge
└── environments/           # 環境別設定
    ├── dev/                # 開発環境
    └── prod/               # 本番環境
```

## アーキテクチャ図

```
                                  ┌─────────────────────────────────────────┐
                                  │              CloudFront                  │
                                  │           (画像CDN配信)                  │
                                  └────────────────┬────────────────────────┘
                                                   │
┌──────────────────────────────────────────────────┼──────────────────────────────────────────────────┐
│                                                  │                                      VPC          │
│  ┌───────────────────────────────────────────────┼───────────────────────────────────┐              │
│  │                                               │                  Public Subnets   │              │
│  │  ┌─────────────────────┐    ┌─────────────────┴─────────────────┐                 │              │
│  │  │   Internet Gateway  │────│      Application Load Balancer    │                 │              │
│  │  └─────────────────────┘    └─────────────────┬─────────────────┘                 │              │
│  │                                               │                                    │              │
│  └───────────────────────────────────────────────┼────────────────────────────────────┘              │
│                                                  │                                                   │
│  ┌───────────────────────────────────────────────┼───────────────────────────────────┐              │
│  │                                               │                 Private Subnets   │              │
│  │                                    ┌──────────┴──────────┐                        │              │
│  │  ┌───────────────┐                 │    ECS Fargate      │                        │              │
│  │  │  NAT Gateway  │                 │   (Go Backend)      │                        │              │
│  │  └───────┬───────┘                 └──────────┬──────────┘                        │              │
│  │          │                                    │                                    │              │
│  │          │                         ┌──────────┴──────────┐                        │              │
│  │          │                         │   RDS PostgreSQL    │                        │              │
│  │          │                         │    (Multi-AZ)       │                        │              │
│  │          │                         └─────────────────────┘                        │              │
│  └──────────┼─────────────────────────────────────────────────────────────────────────┘              │
│             │                                                                                        │
└─────────────┼────────────────────────────────────────────────────────────────────────────────────────┘
              │
              │  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
              └──│       S3        │    │      SNS        │    │    DynamoDB     │
                 │  (画像保存)     │    │ (プッシュ通知)  │    │(デバイストークン)│
                 └─────────────────┘    └─────────────────┘    └─────────────────┘
```

## クイックスタート

### 前提条件

- [Terraform](https://www.terraform.io/downloads) >= 1.5.0
- [AWS CLI](https://aws.amazon.com/cli/) 設定済み
- AWS IAMユーザー/ロールに必要な権限

### 開発環境のデプロイ

```bash
# 1. 設定ファイルを作成
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvarsを編集して環境に合わせた値を設定

# 2. Terraformを初期化
terraform init

# 3. 変更内容を確認
terraform plan

# 4. インフラをデプロイ
terraform apply
```

### 環境別デプロイ

```bash
# 開発環境
cd environments/dev
terraform init
terraform apply

# 本番環境
cd environments/prod
terraform init
terraform apply
```

## モジュール説明

### VPCモジュール (`modules/vpc`)

- VPC (CIDR: 10.0.0.0/16)
- パブリックサブネット x 2 (ALB用)
- プライベートサブネット x 2 (ECS/RDS用)
- Internet Gateway
- NAT Gateway
- セキュリティグループ (ALB, ECS, RDS)

### RDSモジュール (`modules/rds`)

- PostgreSQL 16.x
- Multi-AZ構成（本番のみ）
- AES-256暗号化
- Secrets Managerによる認証情報管理
- 自動バックアップ

### S3/CloudFrontモジュール (`modules/s3`)

- S3バケット（画像保存用）
- SSE-S3暗号化
- CloudFront CDN
- Origin Access Control (OAC)
- ライフサイクルポリシー

### ECSモジュール (`modules/ecs`)

- ECRリポジトリ
- ECS Cluster
- Fargate Task Definition
- ECS Service
- Application Load Balancer
- Auto Scaling (CPU/Memory)
- CloudWatch Logs

### 通知モジュール (`modules/notifications`)

- SNS Topic (通知イベント用)
- SES Configuration Set
- DynamoDB (デバイストークン保存)
- EventBridge Scheduler (Daily cron)

## 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|------------|
| `project_name` | プロジェクト名 | `home-garden` |
| `environment` | 環境名 | `dev` |
| `aws_region` | AWSリージョン | `ap-northeast-1` |
| `db_instance_class` | RDSインスタンスクラス | `db.t3.micro` |
| `db_multi_az` | Multi-AZ構成 | `false` |
| `ecs_task_cpu` | ECS CPU | `256` |
| `ecs_task_memory` | ECS Memory | `512` |

詳細は `variables.tf` を参照してください。

## コスト最適化

### 開発環境

- RDS: db.t3.micro（無料枠対象）
- ECS: 最小構成（1タスク）
- NAT Gateway: 1つのみ
- S3: 自動削除有効

### 本番環境

- RDS: Multi-AZ有効
- ECS: Auto Scaling (2-10タスク)
- CloudFront: キャッシュ最適化
- S3: バージョニング有効

## セキュリティ

- すべてのデータはAES-256で暗号化
- RDS/ECSはプライベートサブネットに配置
- セキュリティグループで最小権限アクセス
- Secrets Managerで認証情報管理
- CloudFront経由でのみS3アクセス可能

## トラブルシューティング

### Terraform state lock

```bash
terraform force-unlock <LOCK_ID>
```

### ECRへのプッシュ

```bash
aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.ap-northeast-1.amazonaws.com
docker push <ECR_REPOSITORY_URL>:latest
```

### ECSサービス更新

```bash
aws ecs update-service --cluster <CLUSTER_NAME> --service <SERVICE_NAME> --force-new-deployment
```

## ライセンス

MIT
