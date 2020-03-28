provider "aws" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
}

module "ecs_base" {
  source = "github.com/schramm-famm/bespin//modules/ecs_base"
  name   = var.name
}

module "ecs_cluster" {
  source                  = "github.com/schramm-famm/bespin//modules/ecs_cluster"
  name                    = var.name
  security_group_ids      = [aws_security_group.karen.id]
  subnets                 = module.ecs_base.vpc_public_subnets
  ec2_instance_profile_id = module.ecs_base.ecs_instance_profile_id
}

resource "aws_security_group" "karen" {
  name        = "${var.name}_allow_testing"
  description = "Allow traffic necessary for integration testing"
  vpc_id      = module.ecs_base.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

module "karen" {
  source          = "./modules/karen"
  name            = var.name
  container_tag   = var.container_tag
  port            = 80
  cluster_id      = module.ecs_cluster.cluster_id
  security_groups = [aws_security_group.karen.id]
  subnets         = module.ecs_base.vpc_public_subnets
  internal        = false
  db_location     = module.rds_instance.db_endpoint
  db_username     = var.rds_username
  db_password     = var.rds_password
}

module "rds_instance" {
  source          = "github.com/schramm-famm/bespin//modules/rds_instance"
  name            = var.name
  engine          = "mariadb"
  engine_version  = "10.2.21"
  port            = 3306
  master_username = var.rds_username
  master_password = var.rds_password
  vpc_id          = module.ecs_base.vpc_id
  subnet_ids      = module.ecs_base.vpc_private_subnets
}
