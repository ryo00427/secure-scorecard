# =============================================================================
# 本番環境設定
# =============================================================================
# このファイルは本番環境用のTerraform設定です。
#
# 使用方法:
#   cd infrastructure/environments/prod
#   terraform init
#   terraform plan
#   terraform apply
#
# 注意事項:
#   - 本番環境のデプロイは慎重に行ってください
#   - 必ずterraform planで変更内容を確認してください
#   - 重要な変更はレビューを経てから適用してください
#
# =============================================================================

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  # 本番環境用S3バックエンド
  backend "s3" {
    bucket         = "home-garden-terraform-state-prod"
    key            = "prod/terraform.tfstate"
    region         = "ap-northeast-1"
    encrypt        = true
    dynamodb_table = "home-garden-terraform-lock"
  }
}

provider "aws" {
  region = "ap-northeast-1"

  default_tags {
    tags = {
      Project     = "home-garden"
      Environment = "prod"
      ManagedBy   = "Terraform"
    }
  }
}

# ルートモジュールを参照
module "infrastructure" {
  source = "../../"

  project_name = "home-garden"
  environment  = "prod"
  aws_region   = "ap-northeast-1"

  # VPC設定
  vpc_cidr             = "10.1.0.0/16"
  availability_zones   = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
  public_subnet_cidrs  = ["10.1.1.0/24", "10.1.2.0/24", "10.1.3.0/24"]
  private_subnet_cidrs = ["10.1.11.0/24", "10.1.12.0/24", "10.1.13.0/24"]

  # RDS設定（本番用は高可用性）
  db_instance_class        = "db.t3.small"
  db_name                  = "home_garden"
  db_username              = "postgres"
  db_multi_az              = true
  db_backup_retention_days = 30

  # ECS設定（本番用は冗長構成）
  ecs_task_cpu      = 512
  ecs_task_memory   = 1024
  ecs_desired_count = 2
  ecs_min_count     = 2
  ecs_max_count     = 10

  # S3設定（本番用は削除保護）
  s3_force_destroy       = false
  cloudfront_price_class = "PriceClass_200"

  # 通知設定
  ses_email_identity = "" # 本番用メールアドレスを設定

  # アプリケーション設定（本番用シークレットは環境変数で渡す）
  jwt_secret           = var.jwt_secret
  cors_allowed_origins = var.cors_allowed_origins
}

# 本番用変数
variable "jwt_secret" {
  description = "JWT署名用シークレット"
  type        = string
  sensitive   = true
}

variable "cors_allowed_origins" {
  description = "CORS許可オリジン"
  type        = list(string)
  default     = []
}

# 出力
output "api_endpoint" {
  value = module.infrastructure.api_endpoint
}

output "ecr_repository_url" {
  value = module.infrastructure.ecr_repository_url
}

output "cloudfront_url" {
  value = module.infrastructure.cloudfront_url
}

output "db_endpoint" {
  value     = module.infrastructure.db_endpoint
  sensitive = true
}
