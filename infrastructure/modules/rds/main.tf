# =============================================================================
# RDSモジュール - PostgreSQLデータベース
# =============================================================================
#
# 構成:
#   - RDS PostgreSQL 16.x インスタンス
#   - Multi-AZ構成（オプション）
#   - AES-256暗号化
#   - Secrets Managerによる認証情報管理
#   - DBサブネットグループ
#   - パラメータグループ
#
# =============================================================================

# -----------------------------------------------------------------------------
# ローカル変数
# -----------------------------------------------------------------------------

locals {
  name_prefix = "${var.project_name}-${var.environment}"
}

# -----------------------------------------------------------------------------
# ランダムパスワード生成
# -----------------------------------------------------------------------------

resource "random_password" "db_password" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# -----------------------------------------------------------------------------
# Secrets Manager - DB認証情報
# -----------------------------------------------------------------------------

resource "aws_secretsmanager_secret" "db_credentials" {
  name        = "${local.name_prefix}-db-credentials"
  description = "RDS PostgreSQL credentials for ${var.project_name}"

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = var.db_username
    password = random_password.db_password.result
    host     = aws_db_instance.main.address
    port     = 5432
    dbname   = var.db_name
  })
}

# -----------------------------------------------------------------------------
# DBサブネットグループ
# -----------------------------------------------------------------------------

resource "aws_db_subnet_group" "main" {
  name        = "${local.name_prefix}-db-subnet-group"
  description = "Database subnet group for ${var.project_name}"
  subnet_ids  = var.private_subnet_ids

  tags = merge(var.tags, {
    Name = "${local.name_prefix}-db-subnet-group"
  })
}

# -----------------------------------------------------------------------------
# パラメータグループ
# -----------------------------------------------------------------------------

resource "aws_db_parameter_group" "main" {
  family      = "postgres16"
  name        = "${local.name_prefix}-pg16"
  description = "Custom parameter group for PostgreSQL 16"

  # ログ設定
  parameter {
    name  = "log_statement"
    value = "all"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000" # 1秒以上のクエリをログ
  }

  # タイムゾーン設定
  parameter {
    name  = "timezone"
    value = "Asia/Tokyo"
  }

  # 接続設定
  parameter {
    name  = "max_connections"
    value = "100"
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# RDSインスタンス
# -----------------------------------------------------------------------------

resource "aws_db_instance" "main" {
  identifier = "${local.name_prefix}-postgres"

  # エンジン設定
  engine               = "postgres"
  engine_version       = "16.4"
  instance_class       = var.db_instance_class
  allocated_storage    = 20
  max_allocated_storage = 100 # ストレージ自動拡張の上限

  # データベース設定
  db_name  = var.db_name
  username = var.db_username
  password = random_password.db_password.result

  # ネットワーク設定
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.ecs_security_group_id]
  publicly_accessible    = false
  port                   = 5432

  # 高可用性設定
  multi_az = var.db_multi_az

  # パラメータグループ
  parameter_group_name = aws_db_parameter_group.main.name

  # ストレージ設定
  storage_type      = "gp3"
  storage_encrypted = true

  # バックアップ設定
  backup_retention_period = var.db_backup_retention_days
  backup_window           = "03:00-04:00" # UTC (JST 12:00-13:00)
  maintenance_window      = "Mon:04:00-Mon:05:00" # UTC

  # スナップショット設定
  skip_final_snapshot       = var.environment != "prod"
  final_snapshot_identifier = var.environment == "prod" ? "${local.name_prefix}-final-snapshot" : null
  copy_tags_to_snapshot     = true

  # 削除保護（本番のみ）
  deletion_protection = var.environment == "prod"

  # パフォーマンスインサイト（本番のみ）
  performance_insights_enabled = var.environment == "prod"

  # 自動マイナーバージョンアップグレード
  auto_minor_version_upgrade = true

  tags = merge(var.tags, {
    Name = "${local.name_prefix}-postgres"
  })
}

# -----------------------------------------------------------------------------
# セキュリティグループルール（VPCモジュールのRDS SGを参照）
# -----------------------------------------------------------------------------
# 注意: セキュリティグループ自体はVPCモジュールで作成済み
# ここではECS SGからの接続を許可するルールのみ追加が必要な場合に使用

resource "aws_security_group" "rds" {
  name        = "${local.name_prefix}-rds-sg"
  description = "Security group for RDS PostgreSQL"
  vpc_id      = var.vpc_id

  ingress {
    description     = "PostgreSQL from ECS"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.ecs_security_group_id]
  }

  tags = merge(var.tags, {
    Name = "${local.name_prefix}-rds-sg"
  })
}
