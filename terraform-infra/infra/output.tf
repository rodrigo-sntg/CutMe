output "aws_vpc_id" {
    value = aws_vpc.main.id
}

output "public_subnet_1_id" {
  value = aws_subnet.public_subnet_1.id
}
output "public_subnet_2_id" {
  value = aws_subnet.public_subnet_2.id
}

output "ecr_repository_url" {
    value = aws_ecr_repository.ecr_api_cutme.repository_url
}

output "aws_ecs_cluster" {
    value = aws_ecs_cluster.cutme_cluster.id
}