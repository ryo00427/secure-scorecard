# =============================================================================
# 通知モジュール - 変数定義
# =============================================================================

variable "project_name" {
  description = "プロジェクト名"
  type        = string
}

variable "environment" {
  description = "環境名"
  type        = string
}

variable "ses_email_identity" {
  description = "SESで検証するメールアドレス（空の場合はスキップ）"
  type        = string
  default     = ""
}

variable "tags" {
  description = "リソースに付与するタグ"
  type        = map(string)
  default     = {}
}
