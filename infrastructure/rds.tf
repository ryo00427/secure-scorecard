# =============================================================================
# RDS PostgreSQL Configuration
# =============================================================================
# RDS PostgreSQL インスタンスとSecrets Manager統合を定義します。

# -----------------------------------------------------------------------------
# DB Subnet Group
# -----------------------------------------------------------------------------
# RDSインスタンスを配置するサブネットグループ

resource "aws_db_subnet_group" "main" {
  name        = "${var.project_name}-${var.environment}-db-subnet"
  description = "Database subnet group for ${var.project_name}"
  subnet_ids  = aws_subnet.private[*].id

  tags = {
    Name = "${var.project_name}-${var.environment}-db-subnet"
  }
}

# -----------------------------------------------------------------------------
# Secrets Manager - RDS Credentials
# -----------------------------------------------------------------------------
# データベース認証情報をSecrets Managerで管理します。

resource "random_password" "rds_password" {
  length           = 32
  special          = true
  override_special = "!#$%^&*()-_=+[]{}|;:,.<>?"
}

resource "aws_secretsmanager_secret" "rds_credentials" {
  name                    = "${var.project_name}-${var.environment}-rds-credentials"
  description             = "RDS PostgreSQL credentials for ${var.project_name}"
  recovery_window_in_days = var.environment == "prod" ? 30 : 0

  tags = {
    Name = "${var.project_name}-${var.environment}-rds-credentials"
  }
}

resource "aws_secretsmanager_secret_version" "rds_credentials" {
  secret_id = aws_secretsmanager_secret.rds_credentials.id
  secret_string = jsonencode({
    username = var.db_username
    password = random_password.rds_password.result
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    dbname   = var.db_name
  })

  depends_on = [aws_db_instance.main]
}

# -----------------------------------------------------------------------------
# RDS Parameter Group
# -----------------------------------------------------------------------------
# PostgreSQL用のカスタムパラメータグループ

resource "aws_db_parameter_group" "main" {
  name        = "${var.project_name}-${var.environment}-pg16"
  family      = "postgres16"
  description = "Custom parameter group for ${var.project_name}"

  # ログ設定
  parameter {
    name  = "log_min_duration_statement"
    value = "1000" # 1秒以上のクエリをログ出力
  }

  parameter {
    name  = "log_statement"
    value = "ddl" # DDLステートメントをログ出力
  }

  # 接続設定
  parameter {
    name  = "idle_in_transaction_session_timeout"
    value = "60000" # 60秒
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-pg16"
  }
}

# -----------------------------------------------------------------------------
# RDS Instance
# -----------------------------------------------------------------------------
# PostgreSQL 16.x インスタンス

resource "aws_db_instance" "main" {
  identifier = "${var.project_name}-${var.environment}-postgres"

  # エンジン設定
  engine               = "postgres"
  engine_version       = "16.4"
  instance_class       = var.db_instance_class
  allocated_storage    = 20
  max_allocated_storage = 100 # ストレージ自動スケーリング

  # データベース設定
  db_name  = var.db_name
  username = var.db_username
  password = random_password.rds_password.result
  port     = 5432

  # ネットワーク設定
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  # 高可用性設定
  multi_az = var.db_multi_az

  # パラメータグループ
  parameter_group_name = aws_db_parameter_group.main.name

  # ストレージ設定
  storage_type      = "gp3"
  storage_encrypted = true
  # KMSキーを指定しない場合、AWS管理キーが使用される（AES-256）

  # バックアップ設定
  backup_retention_period = var.environment == "prod" ? 30 : 7
  backup_window           = "03:00-04:00" # UTC (JST 12:00-13:00)
  maintenance_window      = "sun:04:00-sun:05:00" # UTC (JST 13:00-14:00)

  # 削除保護
  deletion_protection = var.environment == "prod" ? true : false
  skip_final_snapshot = var.environment == "dev" ? true : false
  final_snapshot_identifier = var.environment != "dev" ? "${var.project_name}-${var.environment}-final-snapshot" : null

  # パフォーマンスインサイト（本番環境のみ）
  performance_insights_enabled          = var.environment == "prod" ? true : false
  performance_insights_retention_period = var.environment == "prod" ? 7 : 0

  # 自動マイナーバージョンアップグレード
  auto_minor_version_upgrade = true

  # CloudWatch Logs
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  tags = {
    Name = "${var.project_name}-${var.environment}-postgres"
  }
}
