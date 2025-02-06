###############################################################################
# Terraform Initialization
###############################################################################
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  # Caso use perfis do AWS CLI, descomente:
  # profile = var.aws_profile
}

###############################################################################
# S3 Bucket Configuration
###############################################################################
resource "aws_s3_bucket" "meu_bucket_processamento" {
  bucket = var.s3_bucket_name
  acl    = "private"

  versioning {
    enabled = true
  }

  # pra permitir acesso do CloudFront
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "arn:aws:s3:::${var.s3_bucket_name}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = "arn:aws:cloudfront::058264063116:distribution/ESO1M1W7PTKQM"
          }
        }
      }
    ]
  })
}

resource "aws_s3_bucket_public_access_block" "meu_bucket_processamento" {
  bucket                  = aws_s3_bucket.meu_bucket_processamento.id
  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_notification" "meu_bucket_notifications" {
  bucket = aws_s3_bucket.meu_bucket_processamento.id

  queue {
    id          = "NjM3ZDAzNWItNzc1MS00YTI4LWE0MGMtY2YyYWQ4NGIzMzQy"
    queue_arn   = "arn:aws:sqs:us-east-1:058264063116:MinhaFila"
    events      = ["s3:ObjectCreated:*"]
  }
}


###############################################################################
# DynamoDB Table Configuration
###############################################################################

resource "aws_dynamodb_table" "arquivos_processados" {
  name     = var.dynamodb_table_name

  hash_key = "id"

  attribute {
    name = "id"
    type = "S" # Tipo String
  }

  attribute {
    name = "userId"
    type = "S" # Tipo String
  }

  billing_mode = "PAY_PER_REQUEST"

  global_secondary_index {
    name            = "UserID-index"
    hash_key        = "userId"
    projection_type = "ALL"
  }

  stream_enabled = false



  deletion_protection_enabled = false

}


###############################################################################
# SQS Queues
###############################################################################

# Fila Principal (MinhaFila)
resource "aws_sqs_queue" "main_queue" {
  name                      = var.sqs_queue_name
  visibility_timeout_seconds = 30
  message_retention_seconds = 345600 # 4 dias
  delay_seconds             = 0
  max_message_size          = 262144 # 256 KB
  receive_wait_time_seconds = 0

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action    = "sqs:SendMessage"
        Resource  = "arn:aws:sqs:${var.aws_region}:058264063116:${var.sqs_queue_name}"
        Condition = {
          ArnLike = {
            "aws:SourceArn" = "arn:aws:s3:::${var.s3_bucket_name}"
          }
        }
      }
    ]
  })

  sqs_managed_sse_enabled = true
}

# Dead Letter Queue
resource "aws_sqs_queue" "dead_letter_queue" {
  name                      = "dead-letter"
  visibility_timeout_seconds = 30
  message_retention_seconds = 345600 # 4 dias
  delay_seconds             = 0
  max_message_size          = 262144 # 256 KB
  receive_wait_time_seconds = 0

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "__default_policy_ID"
    Statement = [
      {
        Sid       = "__owner_statement"
        Effect    = "Allow"
        Principal = {
          AWS = "arn:aws:iam::058264063116:root"
        }
        Action    = "SQS:*"
        Resource  = "arn:aws:sqs:us-east-1:058264063116:dead-letter"
      }
    ]
  })

  sqs_managed_sse_enabled = true
}

# Configurar a política de Redrive para Dead Letter Queue
resource "aws_sqs_queue_policy" "dead_letter_redrive" {
  queue_url = aws_sqs_queue.dead_letter_queue.id

  policy = jsonencode({
    redrivePermission = "byQueue"
    sourceQueueArns   = ["arn:aws:sqs:us-east-1:058264063116:MinhaFila"]
  })
}

resource "aws_sqs_queue_redrive_policy" "main_queue_redrive_policy" {
  queue_url = aws_sqs_queue.main_queue.id

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dead_letter_queue.arn
    maxReceiveCount     = 10
  })
}


###############################################################################
# CloudFront Distribution (apontando para o S3)
###############################################################################
resource "aws_cloudfront_distribution" "this" {
  origin {
    domain_name = "${var.s3_bucket_name}.s3.amazonaws.com"
    origin_id   = "${var.s3_bucket_name}.s3.amazonaws.com-1736340185-149685"

    s3_origin_config {
      origin_access_identity = "E1J851ODIFGFDP"
    }

    connection_attempts = 3
    connection_timeout  = 10
  }

  default_cache_behavior {
    target_origin_id       = "${var.s3_bucket_name}.s3.amazonaws.com-1736340185-149685"
    viewer_protocol_policy = "allow-all"

    allowed_methods = [
      "HEAD",
      "DELETE",
      "POST",
      "GET",
      "OPTIONS",
      "PUT",
      "PATCH"
    ]

    cached_methods = [
      "HEAD",
      "GET"
    ]

    compress       = true
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6"
    origin_request_policy_id = "88a5eaf4-2fd4-4709-b370-b4c650ea3fcf"
    response_headers_policy_id = "5cc3b908-e619-4b99-88e5-2cf7f45965bd"

    lambda_function_association {
      event_type   = "viewer-request"
      lambda_arn   = "arn:aws:lambda:us-east-1:058264063116:function:JWTValidator:7"
      include_body = false
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
    ssl_support_method             = "vip"
    minimum_protocol_version       = "TLSv1"
  }

  default_root_object = "index.html"
  enabled             = true
  http_version        = "http2"
  is_ipv6_enabled     = true
  price_class         = "PriceClass_All"
}