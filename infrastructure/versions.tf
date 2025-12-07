# =============================================================================
# Terraform バージョンとプロバイダー設定
# =============================================================================
# 必要なTerraformバージョンとAWSプロバイダーを定義します。

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

  # S3バックエンド設定（本番環境用）
  # 初回は以下をコメントアウトし、S3バケットとDynamoDBテーブル作成後に有効化
  # backend "s3" {
  #   bucket         = "home-garden-terraform-state"
  #   key            = "terraform.tfstate"
  #   region         = "ap-northeast-1"
  #   encrypt        = true
  #   dynamodb_table = "home-garden-terraform-lock"
  # }
}

# AWSプロバイダー設定
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# ランダム文字列生成プロバイダー（パスワード等に使用）
provider "random" {}
