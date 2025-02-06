###############################################################################
# Outputs corrigidos
###############################################################################
output "s3_bucket_name" {
  description = "Nome do bucket S3"
  value       = aws_s3_bucket.meu_bucket_processamento.bucket
}

output "dynamodb_table_name" {
  description = "Nome da tabela do DynamoDB"
  value       = aws_dynamodb_table.arquivos_processados.name
}

output "sqs_queue_url" {
  description = "URL da fila SQS"
  value       = aws_sqs_queue.main_queue.url
}


output "cloudfront_domain" {
  description = "URL da distribuição CloudFront"
  value       = aws_cloudfront_distribution.this.domain_name
}
