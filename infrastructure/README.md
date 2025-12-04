# Infrastructure

AWS インフラストラクチャ（Terraform）

## 構成予定

- VPC、サブネット（Public/Private）
- ECS Fargate（バックエンド）
- RDS PostgreSQL
- S3 + CloudFront（画像CDN）
- SNS/SES（通知）

## セットアップ

```bash
cd infrastructure
terraform init
terraform plan
terraform apply
```

## 注意

このディレクトリは Task 2.x で実装予定です。
