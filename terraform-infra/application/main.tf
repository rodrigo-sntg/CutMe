provider "aws" {
  region = var.aws_region
}

data "terraform_remote_state" "infra" {
  backend = "s3"
  config = {
    bucket = "terraform-s3-state-cutme"
    key    = "cutme/infra"
    region = "us-east-1"
  }
}

# CloudWatch
resource "aws_cloudwatch_log_group" "ecs_log_group" {
  name              = "/ecs/${var.api_name}"
  retention_in_days = 1
}

# Role
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "ecs_task_execution_role"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action    = "sts:AssumeRole",
        Effect    = "Allow",
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

# Policies
resource "aws_iam_role_policy_attachment" "ecs_task_cloudwatch_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}
resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}
resource "aws_iam_role_policy_attachment" "ecs_task_ecr_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_ecs_task_definition" "node_api_task" {
  family                   = "${var.api_name}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "${var.cpu}"
  memory                   = "${var.memory}"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name          = var.api_name
      image         = "${data.terraform_remote_state.infra.outputs.ecr_repository_url}:latest"
      cpu           = var.cpu
      memory        = var.memory
      portMappings  = [
        {
          containerPort = var.application_port
          protocol      = "tcp"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.ecs_log_group.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = var.api_name
        }
      }
      essential = true
    }
  ])
}

# Security Group
resource "aws_security_group" "ecs_security_group" {
  name_prefix = "ecs-sg-"
  vpc_id      = data.terraform_remote_state.infra.outputs.aws_vpc_id

  # HTTP traffic to the Load Balancer
  ingress {
    from_port   = 8080
    to_port     = var.application_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
resource "aws_security_group_rule" "allow_alb_to_ecs" {
  type                      = "ingress"
  from_port                 = 8080
  to_port                   = var.application_port
  protocol                  = "tcp"
  source_security_group_id  = aws_security_group.ecs_security_group.id
  security_group_id         = aws_security_group.ecs_security_group.id
}

# Target group
resource "aws_lb_target_group" "ecs_target_group" {
  name        = "${var.api_name}-target-group"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = data.terraform_remote_state.infra.outputs.aws_vpc_id
  target_type = "ip"

  health_check {
    path                = "/health"
    interval            = 60
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
    matcher             = "200"
  }
}

# Load balancer
resource "aws_lb" "ecs_load_balancer" {
  name               = "${var.api_name}-load-balancer"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.ecs_security_group.id]
  subnets            = [
    data.terraform_remote_state.infra.outputs.public_subnet_1_id,
    data.terraform_remote_state.infra.outputs.public_subnet_2_id
  ]
}
resource "aws_lb_listener" "ecs_listener" {
  load_balancer_arn = aws_lb.ecs_load_balancer.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.ecs_target_group.arn
  }
}

# ECS Service
resource "aws_ecs_service" "node_api_service" {
  name            = "${var.api_name}-service"
  cluster         = data.terraform_remote_state.infra.outputs.aws_ecs_cluster
  task_definition = aws_ecs_task_definition.node_api_task.arn

  deployment_controller {
    type = "ECS"
  }

  desired_count = 1

  network_configuration {
    subnets          = [
      data.terraform_remote_state.infra.outputs.public_subnet_1_id,
      data.terraform_remote_state.infra.outputs.public_subnet_2_id
    ]
    assign_public_ip = true
    security_groups  = [aws_security_group.ecs_security_group.id]
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.ecs_target_group.arn
    container_name   = var.api_name
    container_port   = var.application_port
  }
}

# Adiciona o recurso de auto-scaling para ECS Service
resource "aws_appautoscaling_target" "ecs_service_scaling_target" {
  max_capacity       = 10
  min_capacity       = 1
  resource_id        = "service/${data.terraform_remote_state.infra.outputs.aws_ecs_cluster}/${aws_ecs_service.node_api_service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Adiciona métricas de mensagens na fila SQS
resource "aws_cloudwatch_metric_alarm" "sqs_messages_alarm" {
  alarm_name          = "SQS-Messages-Alarm-${var.api_name}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Average"
  threshold           = 2
  dimensions = {
    QueueName = "MinhaFila2"
  }

  alarm_actions = [aws_appautoscaling_policy.scale_out_policy.arn]
  ok_actions    = [aws_appautoscaling_policy.scale_in_policy.arn]
}

# Escala para cima (scale-out) com base no número de mensagens
resource "aws_appautoscaling_policy" "scale_out_policy" {
  name               = "scale-out-${var.api_name}"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.ecs_service_scaling_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_service_scaling_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_service_scaling_target.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}

# Escala para baixo (scale-in) quando a fila estiver vazia
resource "aws_appautoscaling_policy" "scale_in_policy" {
  name               = "scale-in-${var.api_name}"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.ecs_service_scaling_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_service_scaling_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_service_scaling_target.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_upper_bound = 0
      scaling_adjustment          = -1
    }
  }
}


output "cloudwatch_log_group" {
  value       = aws_cloudwatch_log_group.ecs_log_group.name
  description = "CloudWatch Log Group for ECS Tasks"
}