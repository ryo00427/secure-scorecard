# =============================================================================
# S3/CloudFrontモジュール - 出力定義
# =============================================================================

output "bucket_name" {
  description = "S3バケット名"
  value       = aws_s3_bucket.images.id
}

output "bucket_arn" {
  description = "S3バケットARN"
  value       = aws_s3_bucket.images.arn
}

output "bucket_regional_domain_name" {
  description = "S3バケットのリージョナルドメイン名"
  value       = aws_s3_bucket.images.bucket_regional_domain_name
}

output "cloudfront_distribution_id" {
  description = "CloudFront Distribution ID"
  value       = aws_cloudfront_distribution.images.id
}

output "cloudfront_domain_name" {
  description = "CloudFrontドメイン名"
  value       = aws_cloudfront_distribution.images.domain_name
}

output "cloudfront_url" {
  description = "CloudFront URL（https://）"
  value       = "https://${aws_cloudfront_distribution.images.domain_name}"
}

output "cloudfront_arn" {
  description = "CloudFront Distribution ARN"
  value       = aws_cloudfront_distribution.images.arn
}
