variable "aws_region" {
  description = "AWS Region"
  type        = string
}

variable "api_name" {
  description = "API name"
  type        = string
}

variable "application_port" {
  description = "Application Port"
  type        = number
}

variable "cpu" {
  description = "CPU"
  type        = number
}

variable "memory" {
  description = "Memory"
  type        = number
}