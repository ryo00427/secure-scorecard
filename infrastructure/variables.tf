# =============================================================================
# Input Variables
# =============================================================================
# Terraformで使用する入力変数を定義します。

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
    error_message = "environment は dev, staging, prod のいずれかである必要があります。"
  }
}

variable "aws_region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

# -----------------------------------------------------------------------------
# ネットワーク設定
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
  default     = ["10.0.10.0/24", "10.0.20.0/24"]
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
  description = "データベースマスターユーザー名"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "db_multi_az" {
  description = "Multi-AZ配置を有効にするか"
  type        = bool
  default     = false
}

# -----------------------------------------------------------------------------
# ECS設定
# -----------------------------------------------------------------------------

variable "ecs_task_cpu" {
  description = "ECSタスクのCPU単位（256, 512, 1024, 2048, 4096）"
  type        = number
  default     = 256
}

variable "ecs_task_memory" {
  description = "ECSタスクのメモリ（MB）"
  type        = number
  default     = 512
}

variable "ecs_desired_count" {
  description = "ECSサービスの希望タスク数"
  type        = number
  default     = 1
}

variable "ecs_min_capacity" {
  description = "ECS Auto Scalingの最小タスク数"
  type        = number
  default     = 1
}

variable "ecs_max_capacity" {
  description = "ECS Auto Scalingの最大タスク数"
  type        = number
  default     = 3
}

# -----------------------------------------------------------------------------
# アプリケーション設定
# -----------------------------------------------------------------------------

variable "app_port" {
  description = "アプリケーションのポート番号"
  type        = number
  default     = 8080
}

variable "health_check_path" {
  description = "ヘルスチェックのパス"
  type        = string
  default     = "/health"
}

# -----------------------------------------------------------------------------
# ドメイン設定（オプション）
# -----------------------------------------------------------------------------

variable "domain_name" {
  description = "カスタムドメイン名（オプション）"
  type        = string
  default     = ""
}

variable "certificate_arn" {
  description = "ACM証明書のARN（HTTPS用、オプション）"
  type        = string
  default     = ""
}
