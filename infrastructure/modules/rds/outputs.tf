# =============================================================================
# RDSモジュール - 出力定義
# =============================================================================

output "db_endpoint" {
  description = "RDSエンドポイント（ホスト名）"
  value       = aws_db_instance.main.address
}

output "db_port" {
  description = "RDSポート番号"
  value       = aws_db_instance.main.port
}

output "db_name" {
  description = "データベース名"
  value       = aws_db_instance.main.db_name
}

output "db_instance_id" {
  description = "RDSインスタンスID"
  value       = aws_db_instance.main.id
}

output "db_instance_arn" {
  description = "RDSインスタンスARN"
  value       = aws_db_instance.main.arn
}

output "db_secret_arn" {
  description = "DB認証情報のSecrets Manager ARN"
  value       = aws_secretsmanager_secret.db_credentials.arn
}

output "db_secret_name" {
  description = "DB認証情報のSecrets Manager名"
  value       = aws_secretsmanager_secret.db_credentials.name
}

output "db_security_group_id" {
  description = "RDSセキュリティグループID"
  value       = aws_security_group.rds.id
}
