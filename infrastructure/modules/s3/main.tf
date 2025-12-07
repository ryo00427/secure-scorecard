# =============================================================================
# S3/CloudFrontモジュール - 画像保存とCDN配信
# =============================================================================
#
# 構成:
#   - S3バケット（画像保存用）
#   - S3暗号化（SSE-S3）
#   - CORS設定
#   - ライフサイクルポリシー
#   - CloudFront Distribution（CDN）
#   - Origin Access Control（OAC）
#
# =============================================================================

# -----------------------------------------------------------------------------
# ローカル変数
# -----------------------------------------------------------------------------

locals {
  name_prefix = "${var.project_name}-${var.environment}"
  bucket_name = "${local.name_prefix}-images-${random_string.bucket_suffix.result}"
}

# バケット名のユニーク化用ランダム文字列
resource "random_string" "bucket_suffix" {
  length  = 8
  special = false
  upper   = false
}

# -----------------------------------------------------------------------------
# S3バケット
# -----------------------------------------------------------------------------

resource "aws_s3_bucket" "images" {
  bucket        = local.bucket_name
  force_destroy = var.force_destroy

  tags = merge(var.tags, {
    Name = local.bucket_name
  })
}

# バケットバージョニング
resource "aws_s3_bucket_versioning" "images" {
  bucket = aws_s3_bucket.images.id

  versioning_configuration {
    status = "Enabled"
  }
}

# サーバーサイド暗号化
resource "aws_s3_bucket_server_side_encryption_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# パブリックアクセスブロック
resource "aws_s3_bucket_public_access_block" "images" {
  bucket = aws_s3_bucket.images.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# CORS設定（フロントエンドからの直接アップロード用）
resource "aws_s3_bucket_cors_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "HEAD"]
    allowed_origins = ["*"] # 本番では特定オリジンに制限
    expose_headers  = ["ETag", "Content-Length"]
    max_age_seconds = 3600
  }
}

# ライフサイクルポリシー
resource "aws_s3_bucket_lifecycle_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  # 古いバージョンの削除
  rule {
    id     = "delete-old-versions"
    status = "Enabled"

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }

  # マルチパートアップロードの失敗分をクリーンアップ
  rule {
    id     = "abort-incomplete-multipart"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }

  # 一時ファイルの自動削除
  rule {
    id     = "delete-temp-files"
    status = "Enabled"

    filter {
      prefix = "temp/"
    }

    expiration {
      days = 1
    }
  }
}

# -----------------------------------------------------------------------------
# CloudFront Origin Access Control
# -----------------------------------------------------------------------------

resource "aws_cloudfront_origin_access_control" "images" {
  name                              = "${local.name_prefix}-oac"
  description                       = "OAC for ${local.bucket_name}"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# -----------------------------------------------------------------------------
# CloudFront Distribution
# -----------------------------------------------------------------------------

resource "aws_cloudfront_distribution" "images" {
  origin {
    domain_name              = aws_s3_bucket.images.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.images.id
    origin_id                = "S3-${local.bucket_name}"
  }

  enabled             = true
  is_ipv6_enabled     = true
  comment             = "CDN for ${var.project_name} images"
  default_root_object = ""

  # 価格クラス
  price_class = var.cloudfront_price_class

  # デフォルトキャッシュ動作
  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-${local.bucket_name}"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 86400    # 1日
    max_ttl                = 31536000 # 1年

    compress = true
  }

  # 地理的制限（なし）
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # SSL/TLS証明書（CloudFrontデフォルト）
  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# S3バケットポリシー（CloudFrontからのアクセスのみ許可）
# -----------------------------------------------------------------------------

resource "aws_s3_bucket_policy" "images" {
  bucket = aws_s3_bucket.images.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCloudFrontAccess"
        Effect = "Allow"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.images.arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.images.arn
          }
        }
      }
    ]
  })
}
