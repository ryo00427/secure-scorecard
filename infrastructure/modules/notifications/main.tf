# =============================================================================
# 通知モジュール - SNS/SES/DynamoDB/EventBridge
# =============================================================================
#
# 構成:
#   - SNS Topic（プッシュ通知用）
#   - SNS Platform Application（FCM/APNS用プレースホルダー）
#   - SES Identity（メール通知用）
#   - DynamoDB Table（デバイストークン保存用）
#   - EventBridge Scheduler（Daily cron job）
#
# =============================================================================

# -----------------------------------------------------------------------------
# ローカル変数
# -----------------------------------------------------------------------------

locals {
  name_prefix = "${var.project_name}-${var.environment}"
}

# -----------------------------------------------------------------------------
# SNS Topic（通知イベント用）
# -----------------------------------------------------------------------------

resource "aws_sns_topic" "notifications" {
  name = "${local.name_prefix}-notifications"

  tags = var.tags
}

# SNSトピックポリシー
resource "aws_sns_topic_policy" "notifications" {
  arn = aws_sns_topic.notifications.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowECSPublish"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.notifications.arn
      }
    ]
  })
}

# プッシュ通知用トピック（FCM/APNS配信用）
resource "aws_sns_topic" "push_notifications" {
  name = "${local.name_prefix}-push"

  tags = var.tags
}

# -----------------------------------------------------------------------------
# SES Identity（メール通知用）
# -----------------------------------------------------------------------------

# メールアドレス検証（設定されている場合のみ）
resource "aws_ses_email_identity" "notification" {
  count = var.ses_email_identity != "" ? 1 : 0
  email = var.ses_email_identity
}

# SES設定セット（メトリクス・バウンス処理用）
resource "aws_ses_configuration_set" "main" {
  name = "${local.name_prefix}-config"

  reputation_metrics_enabled = true
  sending_enabled            = true
}

# バウンス・苦情通知用SNSトピック
resource "aws_sns_topic" "ses_feedback" {
  name = "${local.name_prefix}-ses-feedback"

  tags = var.tags
}

# SES Event Destination（配信イベントの追跡）
resource "aws_ses_event_destination" "feedback" {
  name                   = "${local.name_prefix}-feedback"
  configuration_set_name = aws_ses_configuration_set.main.name
  enabled                = true
  matching_types         = ["bounce", "complaint", "delivery"]

  sns_destination {
    topic_arn = aws_sns_topic.ses_feedback.arn
  }
}

# -----------------------------------------------------------------------------
# DynamoDB Table（デバイストークン保存用）
# -----------------------------------------------------------------------------

resource "aws_dynamodb_table" "device_tokens" {
  name           = "${local.name_prefix}-device-tokens"
  billing_mode   = "PAY_PER_REQUEST" # オンデマンドキャパシティ
  hash_key       = "user_id"
  range_key      = "platform"

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "platform"
    type = "S"
  }

  # TTL設定（トークン無効化用）
  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  # ポイントインタイムリカバリ（本番環境推奨）
  point_in_time_recovery {
    enabled = var.environment == "prod"
  }

  tags = var.tags
}

# 通知履歴テーブル（重複通知防止用）
resource "aws_dynamodb_table" "notification_history" {
  name           = "${local.name_prefix}-notification-history"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "notification_id"

  attribute {
    name = "notification_id"
    type = "S"
  }

  # TTL設定（24時間で自動削除）
  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# EventBridge Scheduler（日次タスク用）
# -----------------------------------------------------------------------------

# EventBridge Scheduler用IAMロール
resource "aws_iam_role" "scheduler" {
  name = "${local.name_prefix}-scheduler"

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

  tags = var.tags
}

# Scheduler用ポリシー（SNS Publish権限）
resource "aws_iam_role_policy" "scheduler_sns" {
  name = "${local.name_prefix}-scheduler-sns"
  role = aws_iam_role.scheduler.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = [
          aws_sns_topic.notifications.arn
        ]
      }
    ]
  })
}

# 日次通知スケジュール（毎日朝8時JST = 23:00 UTC前日）
resource "aws_scheduler_schedule" "daily_task_reminder" {
  name       = "${local.name_prefix}-daily-task-reminder"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = "cron(0 23 * * ? *)" # 毎日23:00 UTC（翌日8:00 JST）
  schedule_expression_timezone = "UTC"

  target {
    arn      = aws_sns_topic.notifications.arn
    role_arn = aws_iam_role.scheduler.arn

    input = jsonencode({
      type    = "DAILY_TASK_REMINDER"
      message = "Daily task reminder triggered"
    })
  }

  state = "ENABLED"
}

# 収穫リマインダースケジュール（毎日9時JST = 00:00 UTC）
resource "aws_scheduler_schedule" "harvest_reminder" {
  name       = "${local.name_prefix}-harvest-reminder"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = "cron(0 0 * * ? *)" # 毎日00:00 UTC（9:00 JST）
  schedule_expression_timezone = "UTC"

  target {
    arn      = aws_sns_topic.notifications.arn
    role_arn = aws_iam_role.scheduler.arn

    input = jsonencode({
      type    = "HARVEST_REMINDER"
      message = "Check for upcoming harvests (7 days)"
    })
  }

  state = "ENABLED"
}

# 期限切れタスク検出スケジュール（毎日10時JST = 01:00 UTC）
resource "aws_scheduler_schedule" "overdue_task_check" {
  name       = "${local.name_prefix}-overdue-task-check"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = "cron(0 1 * * ? *)" # 毎日01:00 UTC（10:00 JST）
  schedule_expression_timezone = "UTC"

  target {
    arn      = aws_sns_topic.notifications.arn
    role_arn = aws_iam_role.scheduler.arn

    input = jsonencode({
      type    = "OVERDUE_TASK_CHECK"
      message = "Check for overdue tasks"
    })
  }

  state = "ENABLED"
}
