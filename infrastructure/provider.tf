# =============================================================================
# AWS Provider Configuration
# =============================================================================
# AWSプロバイダーの設定を定義します。
# リージョンはデフォルトで東京（ap-northeast-1）を使用します。

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
