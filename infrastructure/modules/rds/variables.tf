# =============================================================================
# RDSモジュール - 変数定義
# =============================================================================

variable "project_name" {
  description = "プロジェクト名"
  type        = string
}

variable "environment" {
  description = "環境名"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "プライベートサブネットID"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "ECSセキュリティグループID（RDSへのアクセスを許可）"
  type        = string
}

variable "db_instance_class" {
  description = "RDSインスタンスクラス"
  type        = string
  default     = "db.t3.micro"
}

variable "db_name" {
  description = "データベース名"
  type        = string
}

variable "db_username" {
  description = "データベース管理者ユーザー名"
  type        = string
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

variable "tags" {
  description = "リソースに付与するタグ"
  type        = map(string)
  default     = {}
}
