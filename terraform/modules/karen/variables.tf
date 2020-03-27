variable "name" {
  type        = string
  description = "Name used to identify resources"
}

variable "container_tag" {
  type        = string
  description = "Tag of the karen container in the registry to be used"
  default     = "latest"
}

variable "cluster_id" {
  type        = string
  description = "ID of the ECS cluster that the karen service will run in"
}

variable "security_groups" {
  type        = list(string)
  description = "VPC security groups for the karen service load balancer"
}

variable "subnets" {
  type        = list(string)
  description = "VPC subnets for the karen service load balancer"
}

variable "internal" {
  type        = bool
  description = "Toggle whether the load balancer will be internal"
}

variable "db_location" {
  type        = string
  description = "Location (host) of the MariaDB server"
}

variable "db_username" {
  type        = string
  description = "Username for accessing the MariaDB server"
}

variable "db_password" {
  type        = string
  description = "Password for accessing the MariaDB server"
}
