# =============================================================================
# S3/CloudFrontモジュール - 変数定義
# =============================================================================

variable "project_name" {
  description = "プロジェクト名"
  type        = string
}

variable "environment" {
  description = "環境名"
  type        = string
}

variable "force_destroy" {
  description = "S3バケット削除時にオブジェクトも削除するか"
  type        = bool
  default     = false
}

variable "cloudfront_price_class" {
  description = "CloudFrontの価格クラス"
  type        = string
  default     = "PriceClass_200"
}

variable "tags" {
  description = "リソースに付与するタグ"
  type        = map(string)
  default     = {}
}
