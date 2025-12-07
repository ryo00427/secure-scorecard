# =============================================================================
# ECSモジュール - 出力定義
# =============================================================================

output "ecr_repository_url" {
  description = "ECRリポジトリURL"
  value       = aws_ecr_repository.backend.repository_url
}

output "ecr_repository_arn" {
  description = "ECRリポジトリARN"
  value       = aws_ecr_repository.backend.arn
}

output "ecs_cluster_id" {
  description = "ECSクラスターID"
  value       = aws_ecs_cluster.main.id
}

output "ecs_cluster_name" {
  description = "ECSクラスター名"
  value       = aws_ecs_cluster.main.name
}

output "ecs_cluster_arn" {
  description = "ECSクラスターARN"
  value       = aws_ecs_cluster.main.arn
}

output "ecs_service_id" {
  description = "ECSサービスID"
  value       = aws_ecs_service.backend.id
}

output "ecs_service_name" {
  description = "ECSサービス名"
  value       = aws_ecs_service.backend.name
}

output "task_definition_arn" {
  description = "ECSタスク定義ARN"
  value       = aws_ecs_task_definition.backend.arn
}

output "alb_arn" {
  description = "ALB ARN"
  value       = aws_lb.main.arn
}

output "alb_dns_name" {
  description = "ALB DNSエンドポイント"
  value       = aws_lb.main.dns_name
}

output "alb_zone_id" {
  description = "ALB Hosted Zone ID"
  value       = aws_lb.main.zone_id
}

output "target_group_arn" {
  description = "ターゲットグループARN"
  value       = aws_lb_target_group.backend.arn
}

output "cloudwatch_log_group_name" {
  description = "CloudWatchロググループ名"
  value       = aws_cloudwatch_log_group.ecs.name
}

output "ecs_task_execution_role_arn" {
  description = "ECSタスク実行ロールARN"
  value       = aws_iam_role.ecs_task_execution.arn
}

output "ecs_task_role_arn" {
  description = "ECSタスクロールARN"
  value       = aws_iam_role.ecs_task.arn
}
