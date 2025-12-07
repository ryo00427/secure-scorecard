# =============================================================================
# S3 and CloudFront Configuration
# =============================================================================
# 画像保存用S3バケットとCloudFront CDNを定義します。

# -----------------------------------------------------------------------------
# S3 Bucket for Images
# -----------------------------------------------------------------------------
# 作物の成長記録画像を保存するバケット

resource "aws_s3_bucket" "images" {
  bucket = "${var.project_name}-${var.environment}-images"

  force_destroy = var.environment == "dev" ? true : false

  tags = {
    Name = "${var.project_name}-${var.environment}-images"
  }
}

# バケットのバージョニング
resource "aws_s3_bucket_versioning" "images" {
  bucket = aws_s3_bucket.images.id

  versioning_configuration {
    status = var.environment == "prod" ? "Enabled" : "Suspended"
  }
}

# バケットの暗号化（SSE-S3）
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

# CORS設定（フロントエンドからのアップロード許可）
resource "aws_s3_bucket_cors_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "HEAD"]
    allowed_origins = ["http://localhost:3000", "http://localhost:8081"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3600
  }
}

# ライフサイクルポリシー
resource "aws_s3_bucket_lifecycle_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  # 一時ファイル（uploads/tmp/）の自動削除
  rule {
    id     = "cleanup-temp-files"
    status = "Enabled"

    filter {
      prefix = "uploads/tmp/"
    }

    expiration {
      days = 1
    }
  }

  # 古いバージョンの削除（本番環境のみ）
  dynamic "rule" {
    for_each = var.environment == "prod" ? [1] : []
    content {
      id     = "cleanup-old-versions"
      status = "Enabled"

      filter {
        prefix = ""
      }

      noncurrent_version_expiration {
        noncurrent_days = 90
      }
    }
  }
}

# -----------------------------------------------------------------------------
# CloudFront Origin Access Control
# -----------------------------------------------------------------------------
# CloudFrontからS3へのアクセスを制御

resource "aws_cloudfront_origin_access_control" "images" {
  name                              = "${var.project_name}-${var.environment}-images-oac"
  description                       = "OAC for images bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# -----------------------------------------------------------------------------
# S3 Bucket Policy for CloudFront
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

  depends_on = [aws_cloudfront_distribution.images]
}

# -----------------------------------------------------------------------------
# CloudFront Distribution
# -----------------------------------------------------------------------------
# 画像配信用CDN

resource "aws_cloudfront_distribution" "images" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "${var.project_name} ${var.environment} images CDN"
  default_root_object = ""
  price_class         = "PriceClass_200" # アジア、オーストラリア、ヨーロッパ、北米

  # S3オリジン
  origin {
    domain_name              = aws_s3_bucket.images.bucket_regional_domain_name
    origin_id                = "S3-${aws_s3_bucket.images.id}"
    origin_access_control_id = aws_cloudfront_origin_access_control.images.id
  }

  # デフォルトキャッシュ動作
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${aws_s3_bucket.images.id}"
    viewer_protocol_policy = "redirect-to-https"

    # キャッシュポリシー（Managed-CachingOptimized）
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6"

    # オリジンリクエストポリシー（CORS-S3Origin）
    origin_request_policy_id = "88a5eaf4-2fd4-4709-b370-b4c650ea3fcf"

    compress = true
  }

  # 地理的制限（なし）
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # SSL証明書（CloudFront デフォルト）
  viewer_certificate {
    cloudfront_default_certificate = true
  }

  # カスタムエラーレスポンス
  custom_error_response {
    error_code         = 403
    response_code      = 404
    response_page_path = ""
  }

  custom_error_response {
    error_code         = 404
    response_code      = 404
    response_page_path = ""
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-images-cdn"
  }
}
