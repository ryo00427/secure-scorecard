# =============================================================================
# Terraform Version Constraints
# =============================================================================
# Terraformとプロバイダーのバージョン制約を定義します。

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
