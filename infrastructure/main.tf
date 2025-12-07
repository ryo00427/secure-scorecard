# =============================================================================
# Home Garden Management - メインTerraform設定
# =============================================================================
#
# このファイルは全モジュールを統合し、インフラストラクチャ全体を構築します。
#
# 構成:
#   - VPC: ネットワーク基盤（パブリック/プライベートサブネット、NAT Gateway）
#   - RDS: PostgreSQLデータベース（Multi-AZ対応）
#   - S3/CloudFront: 画像保存とCDN配信
#   - ECR/ECS: コンテナレジストリとFargateサービス
#   - Notifications: SNS/SES/DynamoDBによる通知基盤
#
# =============================================================================

# -----------------------------------------------------------------------------
# ローカル変数
# -----------------------------------------------------------------------------

locals {
  # リソース命名用プレフィックス
  name_prefix = "${var.project_name}-${var.environment}"

  # 共通タグ
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# -----------------------------------------------------------------------------
# VPCモジュール (Task 2.1)
# -----------------------------------------------------------------------------
# VPC、サブネット、Internet Gateway、NAT Gateway、セキュリティグループを作成

module "vpc" {
  source = "./modules/vpc"

  project_name         = var.project_name
  environment          = var.environment
  vpc_cidr             = var.vpc_cidr
  availability_zones   = var.availability_zones
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# RDSモジュール (Task 2.2)
# -----------------------------------------------------------------------------
# PostgreSQL RDSインスタンスとSecrets Manager統合

module "rds" {
  source = "./modules/rds"

  project_name             = var.project_name
  environment              = var.environment
  vpc_id                   = module.vpc.vpc_id
  private_subnet_ids       = module.vpc.private_subnet_ids
  ecs_security_group_id    = module.vpc.ecs_security_group_id
  db_instance_class        = var.db_instance_class
  db_name                  = var.db_name
  db_username              = var.db_username
  db_multi_az              = var.db_multi_az
  db_backup_retention_days = var.db_backup_retention_days

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# S3/CloudFrontモジュール (Task 2.3)
# -----------------------------------------------------------------------------
# 画像保存用S3バケットとCloudFront CDN

module "s3" {
  source = "./modules/s3"

  project_name           = var.project_name
  environment            = var.environment
  force_destroy          = var.s3_force_destroy
  cloudfront_price_class = var.cloudfront_price_class

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# ECS/ECRモジュール (Task 2.4)
# -----------------------------------------------------------------------------
# ECRリポジトリ、ECS Cluster、Fargateサービス、ALB

module "ecs" {
  source = "./modules/ecs"

  project_name          = var.project_name
  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  public_subnet_ids     = module.vpc.public_subnet_ids
  private_subnet_ids    = module.vpc.private_subnet_ids
  alb_security_group_id = module.vpc.alb_security_group_id
  ecs_security_group_id = module.vpc.ecs_security_group_id
  task_cpu              = var.ecs_task_cpu
  task_memory           = var.ecs_task_memory
  desired_count         = var.ecs_desired_count
  min_count             = var.ecs_min_count
  max_count             = var.ecs_max_count

  # アプリケーション環境変数
  db_host              = module.rds.db_endpoint
  db_name              = var.db_name
  db_secret_arn        = module.rds.db_secret_arn
  s3_bucket_name       = module.s3.bucket_name
  cloudfront_url       = module.s3.cloudfront_url
  jwt_secret           = var.jwt_secret
  cors_allowed_origins = var.cors_allowed_origins

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# 通知モジュール (Task 2.5)
# -----------------------------------------------------------------------------
# SNS、SES、DynamoDB、EventBridge Scheduler

module "notifications" {
  source = "./modules/notifications"

  project_name       = var.project_name
  environment        = var.environment
  ses_email_identity = var.ses_email_identity

  tags = local.common_tags
}
