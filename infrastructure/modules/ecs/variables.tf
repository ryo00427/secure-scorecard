# =============================================================================
# ECSモジュール - 変数定義
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

variable "public_subnet_ids" {
  description = "パブリックサブネットID（ALB用）"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "プライベートサブネットID（ECSタスク用）"
  type        = list(string)
}

variable "alb_security_group_id" {
  description = "ALBセキュリティグループID"
  type        = string
}

variable "ecs_security_group_id" {
  description = "ECSセキュリティグループID"
  type        = string
}

variable "task_cpu" {
  description = "ECSタスクのCPUユニット"
  type        = number
  default     = 256
}

variable "task_memory" {
  description = "ECSタスクのメモリ（MiB）"
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "希望タスク数"
  type        = number
  default     = 1
}

variable "min_count" {
  description = "最小タスク数"
  type        = number
  default     = 1
}

variable "max_count" {
  description = "最大タスク数"
  type        = number
  default     = 4
}

# アプリケーション設定
variable "db_host" {
  description = "データベースホスト"
  type        = string
}

variable "db_name" {
  description = "データベース名"
  type        = string
}

variable "db_secret_arn" {
  description = "DB認証情報のSecrets Manager ARN"
  type        = string
}

variable "s3_bucket_name" {
  description = "画像保存用S3バケット名"
  type        = string
}

variable "cloudfront_url" {
  description = "CloudFront URL"
  type        = string
}

variable "jwt_secret" {
  description = "JWT署名用シークレット"
  type        = string
  sensitive   = true
  default     = ""
}

variable "cors_allowed_origins" {
  description = "CORS許可オリジン"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "リソースに付与するタグ"
  type        = map(string)
  default     = {}
}
