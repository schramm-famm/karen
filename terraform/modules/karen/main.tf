data "aws_region" "karen" {}

resource "aws_cloudwatch_log_group" "karen" {
  name              = "${var.name}_karen"
  retention_in_days = 1
}

resource "aws_ecs_task_definition" "karen" {
  family       = "${var.name}_karen"
  network_mode = "bridge"

  container_definitions = <<EOF
[
  {
    "name": "${var.name}_karen",
    "image": "343660461351.dkr.ecr.us-east-2.amazonaws.com/karen:${var.container_tag}",
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
            "awslogs-group": "${aws_cloudwatch_log_group.karen.name}",
            "awslogs-region": "${data.aws_region.karen.name}",
            "awslogs-stream-prefix": "${var.name}"
        }
    },
    "cpu": 10,
    "memory": 128,
    "essential": true,
    "environment": [
        {
            "name": "KAREN_DB_LOCATION",
            "value": "${var.db_location}"
        },
        {
            "name": "KAREN_DB_USERNAME",
            "value": "${var.db_username}"
        },
        {
            "name": "KAREN_DB_PASSWORD",
            "value": "${var.db_password}"
        }
    ],
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80,
        "protocol": "tcp"
      }
    ]
  }
]
EOF
}

resource "aws_elb" "karen" {
  name            = "${var.name}-karen"
  subnets         = var.subnets
  security_groups = var.security_groups
  internal        = var.internal

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_ecs_service" "karen" {
  name            = "${var.name}_karen"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.karen.arn

  load_balancer {
    elb_name       = aws_elb.karen.name
    container_name = "${var.name}_karen"
    container_port = 80
  }

  desired_count = 1
}
