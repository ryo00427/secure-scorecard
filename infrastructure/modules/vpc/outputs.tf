# =============================================================================
# VPCモジュール - 出力定義
# =============================================================================

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "vpc_cidr_block" {
  description = "VPCのCIDRブロック"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "パブリックサブネットID"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "プライベートサブネットID"
  value       = aws_subnet.private[*].id
}

output "internet_gateway_id" {
  description = "Internet Gateway ID"
  value       = aws_internet_gateway.main.id
}

output "nat_gateway_id" {
  description = "NAT Gateway ID"
  value       = aws_nat_gateway.main.id
}

output "alb_security_group_id" {
  description = "ALBセキュリティグループID"
  value       = aws_security_group.alb.id
}

output "ecs_security_group_id" {
  description = "ECSセキュリティグループID"
  value       = aws_security_group.ecs.id
}

output "rds_security_group_id" {
  description = "RDSセキュリティグループID"
  value       = aws_security_group.rds.id
}
