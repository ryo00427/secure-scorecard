# =============================================================================
# Terraform出力定義
# =============================================================================
# デプロイ後に必要な情報を出力します。

# -----------------------------------------------------------------------------
# VPC出力
# -----------------------------------------------------------------------------

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "パブリックサブネットID"
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "プライベートサブネットID"
  value       = module.vpc.private_subnet_ids
}

# -----------------------------------------------------------------------------
# RDS出力
# -----------------------------------------------------------------------------

output "db_endpoint" {
  description = "RDSエンドポイント"
  value       = module.rds.db_endpoint
}

output "db_secret_arn" {
  description = "データベース認証情報のSecrets Manager ARN"
  value       = module.rds.db_secret_arn
  sensitive   = true
}

# -----------------------------------------------------------------------------
# S3/CloudFront出力
# -----------------------------------------------------------------------------

output "s3_bucket_name" {
  description = "画像保存用S3バケット名"
  value       = module.s3.bucket_name
}

output "s3_bucket_arn" {
  description = "S3バケットARN"
  value       = module.s3.bucket_arn
}

output "cloudfront_url" {
  description = "CloudFront配信URL"
  value       = module.s3.cloudfront_url
}

output "cloudfront_distribution_id" {
  description = "CloudFront Distribution ID"
  value       = module.s3.cloudfront_distribution_id
}

# -----------------------------------------------------------------------------
# ECS出力
# -----------------------------------------------------------------------------

output "ecr_repository_url" {
  description = "ECRリポジトリURL"
  value       = module.ecs.ecr_repository_url
}

output "ecs_cluster_name" {
  description = "ECSクラスター名"
  value       = module.ecs.ecs_cluster_name
}

output "ecs_service_name" {
  description = "ECSサービス名"
  value       = module.ecs.ecs_service_name
}

output "alb_dns_name" {
  description = "ALB DNSエンドポイント"
  value       = module.ecs.alb_dns_name
}

output "api_endpoint" {
  description = "APIエンドポイントURL"
  value       = "http://${module.ecs.alb_dns_name}"
}

# -----------------------------------------------------------------------------
# 通知出力
# -----------------------------------------------------------------------------

output "sns_topic_arn" {
  description = "プッシュ通知用SNSトピックARN"
  value       = module.notifications.sns_topic_arn
}

output "dynamodb_table_name" {
  description = "デバイストークン保存用DynamoDBテーブル名"
  value       = module.notifications.dynamodb_table_name
}

# -----------------------------------------------------------------------------
# デプロイ情報
# -----------------------------------------------------------------------------

output "deployment_info" {
  description = "デプロイに必要な情報のまとめ"
  value = {
    api_endpoint     = "http://${module.ecs.alb_dns_name}"
    ecr_repository   = module.ecs.ecr_repository_url
    cloudfront_url   = module.s3.cloudfront_url
    s3_bucket        = module.s3.bucket_name
    ecs_cluster      = module.ecs.ecs_cluster_name
    ecs_service      = module.ecs.ecs_service_name
    db_endpoint      = module.rds.db_endpoint
    aws_region       = var.aws_region
  }
}
