# =============================================================================
# SNS/SES/DynamoDB Configuration
# =============================================================================
# プッシュ通知・メール通知のためのAWSサービスを定義します。

# -----------------------------------------------------------------------------
# SNS Topic for Notifications
# -----------------------------------------------------------------------------
# アプリケーション内部の通知イベント配信用トピック

resource "aws_sns_topic" "notifications" {
  name = "${var.project_name}-${var.environment}-notifications"

  tags = {
    Name = "${var.project_name}-${var.environment}-notifications"
  }
}

# SNS Topic Policy
resource "aws_sns_topic_policy" "notifications" {
  arn = aws_sns_topic.notifications.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowECSPublish"
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_role.ecs_task.arn
        }
        Action   = "SNS:Publish"
        Resource = aws_sns_topic.notifications.arn
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# SNS Platform Application (FCM for Android)
# -----------------------------------------------------------------------------
# 注意: FCM Server Keyは手動で設定するか、Secrets Managerから取得してください

# resource "aws_sns_platform_application" "fcm" {
#   name     = "${var.project_name}-${var.environment}-fcm"
#   platform = "GCM"
#   platform_credential = "YOUR_FCM_SERVER_KEY" # Secrets Managerから取得推奨
# }

# -----------------------------------------------------------------------------
# SNS Platform Application (APNS for iOS)
# -----------------------------------------------------------------------------
# 注意: APNS証明書は手動で設定するか、Secrets Managerから取得してください

# resource "aws_sns_platform_application" "apns" {
#   name                = "${var.project_name}-${var.environment}-apns"
#   platform            = var.environment == "prod" ? "APNS" : "APNS_SANDBOX"
#   platform_credential = "YOUR_APNS_PRIVATE_KEY"
#   platform_principal  = "YOUR_APNS_CERTIFICATE"
# }

# -----------------------------------------------------------------------------
# DynamoDB Table for Device Tokens
# -----------------------------------------------------------------------------
# プッシュ通知用デバイストークンを保存

resource "aws_dynamodb_table" "device_tokens" {
  name         = "${var.project_name}-${var.environment}-device-tokens"
  billing_mode = "PAY_PER_REQUEST"

  # パーティションキー: ユーザーID
  hash_key = "user_id"
  # ソートキー: プラットフォーム（ios, android）
  range_key = "platform"

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "platform"
    type = "S"
  }

  # TTL設定（無効なトークンの自動削除用）
  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  # ポイントインタイムリカバリ（本番環境のみ）
  point_in_time_recovery {
    enabled = var.environment == "prod" ? true : false
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-device-tokens"
  }
}

# -----------------------------------------------------------------------------
# DynamoDB Table for Notification Deduplication
# -----------------------------------------------------------------------------
# 重複通知防止用（24時間TTL）

resource "aws_dynamodb_table" "notification_dedup" {
  name         = "${var.project_name}-${var.environment}-notification-dedup"
  billing_mode = "PAY_PER_REQUEST"

  # パーティションキー: 通知のユニークID
  hash_key = "notification_id"

  attribute {
    name = "notification_id"
    type = "S"
  }

  # TTL設定（24時間で自動削除）
  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-notification-dedup"
  }
}

# -----------------------------------------------------------------------------
# SES Configuration
# -----------------------------------------------------------------------------
# メール通知用のSES設定
# 注意: SESはサンドボックスモードで開始されます。
# 本番利用には production access をリクエストしてください。

# SES Email Identity (送信元ドメイン)
# 注意: ドメイン検証が必要です
resource "aws_ses_email_identity" "notifications" {
  count = var.domain_name != "" ? 1 : 0
  email = "notifications@${var.domain_name}"
}

# -----------------------------------------------------------------------------
# EventBridge Scheduler (Daily Cron Job)
# -----------------------------------------------------------------------------
# 期限切れトークン削除、通知チェックなどの定期実行用

resource "aws_scheduler_schedule_group" "main" {
  name = "${var.project_name}-${var.environment}-schedules"

  tags = {
    Name = "${var.project_name}-${var.environment}-schedules"
  }
}

# 毎日午前6時（JST）に実行する通知チェックジョブ
resource "aws_scheduler_schedule" "daily_notification_check" {
  name       = "${var.project_name}-${var.environment}-daily-notification"
  group_name = aws_scheduler_schedule_group.main.name

  flexible_time_window {
    mode = "OFF"
  }

  # 毎日21:00 UTC（翌日06:00 JST）
  schedule_expression = "cron(0 21 * * ? *)"

  target {
    arn      = aws_lambda_function.notification_scheduler.arn
    role_arn = aws_iam_role.scheduler.arn

    retry_policy {
      maximum_event_age_in_seconds = 3600
      maximum_retry_attempts       = 3
    }
  }
}

# -----------------------------------------------------------------------------
# Lambda for Scheduled Tasks
# -----------------------------------------------------------------------------
# EventBridgeから呼び出される定期タスク用Lambda

resource "aws_iam_role" "lambda_scheduler" {
  name = "${var.project_name}-${var.environment}-lambda-scheduler"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-${var.environment}-lambda-scheduler"
  }
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_scheduler.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "lambda_scheduler_permissions" {
  name = "scheduler-permissions"
  role = aws_iam_role.lambda_scheduler.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecs:RunTask"
        ]
        Resource = aws_ecs_task_definition.backend.arn
      },
      {
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = [
          aws_iam_role.ecs_task_execution.arn,
          aws_iam_role.ecs_task.arn
        ]
      }
    ]
  })
}

# Lambda関数（プレースホルダー - 実際のコードはデプロイ時に置き換え）
resource "aws_lambda_function" "notification_scheduler" {
  function_name = "${var.project_name}-${var.environment}-notification-scheduler"
  role          = aws_iam_role.lambda_scheduler.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
  timeout       = 60

  # プレースホルダーコード（実際のコードはCI/CDでデプロイ）
  filename         = data.archive_file.lambda_placeholder.output_path
  source_code_hash = data.archive_file.lambda_placeholder.output_base64sha256

  environment {
    variables = {
      ECS_CLUSTER_ARN      = aws_ecs_cluster.main.arn
      ECS_TASK_DEFINITION  = aws_ecs_task_definition.backend.arn
      ECS_SUBNETS          = join(",", aws_subnet.private[*].id)
      ECS_SECURITY_GROUPS  = aws_security_group.ecs.id
    }
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-notification-scheduler"
  }
}

# Lambdaプレースホルダーコード
data "archive_file" "lambda_placeholder" {
  type        = "zip"
  output_path = "${path.module}/lambda_placeholder.zip"

  source {
    content  = <<EOF
exports.handler = async (event) => {
  console.log('Notification scheduler invoked', JSON.stringify(event));
  // TODO: Implement notification check logic
  return { statusCode: 200, body: 'OK' };
};
EOF
    filename = "index.js"
  }
}

# -----------------------------------------------------------------------------
# EventBridge Scheduler IAM Role
# -----------------------------------------------------------------------------

resource "aws_iam_role" "scheduler" {
  name = "${var.project_name}-${var.environment}-scheduler"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "scheduler.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-${var.environment}-scheduler"
  }
}

resource "aws_iam_role_policy" "scheduler_invoke_lambda" {
  name = "invoke-lambda"
  role = aws_iam_role.scheduler.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "lambda:InvokeFunction"
        Resource = aws_lambda_function.notification_scheduler.arn
      }
    ]
  })
}
