# =============================================================================
# 開発環境設定
# =============================================================================
# このファイルは開発環境用のTerraform設定です。
#
# 使用方法:
#   cd infrastructure/environments/dev
#   terraform init
#   terraform plan
#   terraform apply
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

  # 開発環境用バックエンド（ローカル）
  # 本番ではS3バックエンドを使用
  # backend "s3" {
  #   bucket         = "home-garden-terraform-state-dev"
  #   key            = "dev/terraform.tfstate"
  #   region         = "ap-northeast-1"
  #   encrypt        = true
  #   dynamodb_table = "home-garden-terraform-lock"
  # }
}

provider "aws" {
  region = "ap-northeast-1"

  default_tags {
    tags = {
      Project     = "home-garden"
      Environment = "dev"
      ManagedBy   = "Terraform"
    }
  }
}

# ルートモジュールを参照
module "infrastructure" {
  source = "../../"

  project_name = "home-garden"
  environment  = "dev"
  aws_region   = "ap-northeast-1"

  # VPC設定
  vpc_cidr             = "10.0.0.0/16"
  availability_zones   = ["ap-northeast-1a", "ap-northeast-1c"]
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24"]

  # RDS設定（開発用は小さいインスタンス）
  db_instance_class        = "db.t3.micro"
  db_name                  = "home_garden"
  db_username              = "postgres"
  db_multi_az              = false
  db_backup_retention_days = 1

  # ECS設定（開発用は最小構成）
  ecs_task_cpu      = 256
  ecs_task_memory   = 512
  ecs_desired_count = 1
  ecs_min_count     = 1
  ecs_max_count     = 2

  # S3設定（開発用は削除可能）
  s3_force_destroy       = true
  cloudfront_price_class = "PriceClass_200"

  # 通知設定
  ses_email_identity = ""

  # アプリケーション設定
  jwt_secret           = "dev-secret-change-in-production-32chars"
  cors_allowed_origins = ["http://localhost:3000", "http://localhost:8081"]
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
  value = module.infrastructure.db_endpoint
}
