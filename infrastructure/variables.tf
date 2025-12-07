# =============================================================================
# 共通変数定義
# =============================================================================
# プロジェクト全体で使用される変数を定義します。

# -----------------------------------------------------------------------------
# 基本設定
# -----------------------------------------------------------------------------

variable "project_name" {
  description = "プロジェクト名（リソース命名に使用）"
  type        = string
  default     = "home-garden"
}

variable "environment" {
  description = "環境名（dev, staging, prod）"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "aws_region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

# -----------------------------------------------------------------------------
# VPC設定
# -----------------------------------------------------------------------------

variable "vpc_cidr" {
  description = "VPCのCIDRブロック"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "使用するアベイラビリティゾーン"
  type        = list(string)
  default     = ["ap-northeast-1a", "ap-northeast-1c"]
}

variable "public_subnet_cidrs" {
  description = "パブリックサブネットのCIDRブロック"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "プライベートサブネットのCIDRブロック"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24"]
}

# -----------------------------------------------------------------------------
# RDS設定
# -----------------------------------------------------------------------------

variable "db_instance_class" {
  description = "RDSインスタンスクラス"
  type        = string
  default     = "db.t3.micro"
}

variable "db_name" {
  description = "データベース名"
  type        = string
  default     = "home_garden"
}

variable "db_username" {
  description = "データベース管理者ユーザー名"
  type        = string
  default     = "postgres"
}

variable "db_multi_az" {
  description = "Multi-AZ構成を有効にするか"
  type        = bool
  default     = false
}

variable "db_backup_retention_days" {
  description = "バックアップ保持日数"
  type        = number
  default     = 7
}

# -----------------------------------------------------------------------------
# ECS設定
# -----------------------------------------------------------------------------

variable "ecs_task_cpu" {
  description = "ECSタスクのCPUユニット（256 = 0.25 vCPU）"
  type        = number
  default     = 256
}

variable "ecs_task_memory" {
  description = "ECSタスクのメモリ（MiB）"
  type        = number
  default     = 512
}

variable "ecs_desired_count" {
  description = "ECSサービスの希望タスク数"
  type        = number
  default     = 1
}

variable "ecs_min_count" {
  description = "ECSオートスケーリングの最小タスク数"
  type        = number
  default     = 1
}

variable "ecs_max_count" {
  description = "ECSオートスケーリングの最大タスク数"
  type        = number
  default     = 4
}

# -----------------------------------------------------------------------------
# S3/CloudFront設定
# -----------------------------------------------------------------------------

variable "s3_force_destroy" {
  description = "S3バケット削除時にオブジェクトも削除するか（開発用）"
  type        = bool
  default     = false
}

variable "cloudfront_price_class" {
  description = "CloudFrontの価格クラス"
  type        = string
  default     = "PriceClass_200" # アジア、ヨーロッパ、北米

  validation {
    condition     = contains(["PriceClass_100", "PriceClass_200", "PriceClass_All"], var.cloudfront_price_class)
    error_message = "Price class must be one of: PriceClass_100, PriceClass_200, PriceClass_All."
  }
}

# -----------------------------------------------------------------------------
# 通知設定
# -----------------------------------------------------------------------------

variable "ses_email_identity" {
  description = "SESで検証するメールアドレス（空の場合はスキップ）"
  type        = string
  default     = ""
}

# -----------------------------------------------------------------------------
# アプリケーション設定
# -----------------------------------------------------------------------------

variable "jwt_secret" {
  description = "JWT署名用シークレット"
  type        = string
  sensitive   = true
  default     = ""
}

variable "cors_allowed_origins" {
  description = "CORS許可オリジン"
  type        = list(string)
  default     = ["http://localhost:3000", "http://localhost:8081"]
}
