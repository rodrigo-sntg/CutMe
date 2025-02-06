variable "aws_region" {
  type        = string
  description = "Região AWS onde os recursos serão criados"
  default     = "us-east-1"
}

variable "aws_profile" {
  type        = string
  description = "Perfil AWS do CLI (opcional)"
  default     = "default"
}

variable "s3_bucket_name" {
  type        = string
  description = "Nome do bucket S3 para uploads"
  default     = "meu-bucket-processamento-1"
}

variable "dynamodb_table_name" {
  type        = string
  description = "Nome da tabela DynamoDB"
  default     = "ArquivosProcessados1"
}

variable "sqs_queue_name" {
  type        = string
  description = "Nome da fila SQS"
  default     = "MinhaFila1"
}
