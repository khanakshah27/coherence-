terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "coherence-demo"
      ManagedBy   = "Terraform"
    }
  }
}

# Variables
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "instance_count" {
  description = "Number of EC2 instances"
  type        = number
  default     = 2
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "coherence-vpc"
  }
}

# Subnets
resource "aws_subnet" "main" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "coherence-subnet-${count.index + 1}"
  }
}

# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Security Group
resource "aws_security_group" "web" {
  name        = "coherence-web-sg"
  description = "Security group for web servers"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "coherence-web-sg"
  }
}

# EC2 Instances
resource "aws_instance" "web" {
  count                = var.instance_count
  ami                  = data.aws_ami.amazon_linux_2.id
  instance_type        = "t3.medium"
  subnet_id            = aws_subnet.main[count.index % 2].id
  vpc_security_group_ids = [aws_security_group.web.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 50
    delete_on_termination = true
    encrypted             = true
  }

  tags = {
    Name = "coherence-web-${count.index + 1}"
  }
}

# AMI data source
data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# S3 Bucket
resource "aws_s3_bucket" "app" {
  bucket = "coherence-app-bucket-${data.aws_caller_identity.current.account_id}"

  tags = {
    Name = "coherence-app-bucket"
  }
}

# S3 Bucket Versioning
resource "aws_s3_bucket_versioning" "app" {
  bucket = aws_s3_bucket.app.id

  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket Server-Side Encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "app" {
  bucket = aws_s3_bucket.app.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 Block Public Access
resource "aws_s3_bucket_public_access_block" "app" {
  bucket = aws_s3_bucket.app.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# RDS Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "coherence-db-subnet-group"
  subnet_ids = aws_subnet.main[*].id

  tags = {
    Name = "coherence-db-subnet-group"
  }
}

# RDS Instance
resource "aws_db_instance" "main" {
  identifier            = "coherence-postgres"
  engine                = "postgres"
  engine_version        = "15.1"
  instance_class        = "db.t3.micro"
  allocated_storage     = 20
  storage_type          = "gp3"
  storage_encrypted     = true
  db_subnet_group_name  = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.web.id]

  db_name  = "coherence"
  username = "coherence"
  password = "ChangeMe123!"

  multi_az               = false
  publicly_accessible    = false
  skip_final_snapshot    = true
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"

  enabled_cloudwatch_logs_exports = ["postgresql"]

  tags = {
    Name = "coherence-postgres"
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "coherence" {
  name              = "/aws/coherence/app"
  retention_in_days = 7

  tags = {
    Name = "coherence-logs"
  }
}

# Data source for current AWS account
data "aws_caller_identity" "current" {}

# Outputs
output "vpc_id" {
  value       = aws_vpc.main.id
  description = "VPC ID"
}

output "instance_ids" {
  value       = aws_instance.web[*].id
  description = "EC2 Instance IDs"
}

output "instance_private_ips" {
  value       = aws_instance.web[*].private_ip
  description = "EC2 Instance Private IPs"
}

output "s3_bucket_name" {
  value       = aws_s3_bucket.app.id
  description = "S3 Bucket Name"
}

output "rds_endpoint" {
  value       = aws_db_instance.main.endpoint
  description = "RDS Endpoint"
}

output "rds_database_name" {
  value       = aws_db_instance.main.db_name
  description = "RDS Database Name"
}
