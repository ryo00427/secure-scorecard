# =============================================================================
# 通知モジュール - 出力定義
# =============================================================================

output "sns_topic_arn" {
  description = "通知イベント用SNSトピックARN"
  value       = aws_sns_topic.notifications.arn
}

output "sns_topic_name" {
  description = "通知イベント用SNSトピック名"
  value       = aws_sns_topic.notifications.name
}

output "push_notifications_topic_arn" {
  description = "プッシュ通知用SNSトピックARN"
  value       = aws_sns_topic.push_notifications.arn
}

output "ses_feedback_topic_arn" {
  description = "SESフィードバック用SNSトピックARN"
  value       = aws_sns_topic.ses_feedback.arn
}

output "ses_configuration_set_name" {
  description = "SES設定セット名"
  value       = aws_ses_configuration_set.main.name
}

output "dynamodb_table_name" {
  description = "デバイストークン保存用DynamoDBテーブル名"
  value       = aws_dynamodb_table.device_tokens.name
}

output "dynamodb_table_arn" {
  description = "デバイストークン保存用DynamoDBテーブルARN"
  value       = aws_dynamodb_table.device_tokens.arn
}

output "notification_history_table_name" {
  description = "通知履歴用DynamoDBテーブル名"
  value       = aws_dynamodb_table.notification_history.name
}

output "notification_history_table_arn" {
  description = "通知履歴用DynamoDBテーブルARN"
  value       = aws_dynamodb_table.notification_history.arn
}

output "scheduler_role_arn" {
  description = "EventBridge SchedulerロールARN"
  value       = aws_iam_role.scheduler.arn
}
