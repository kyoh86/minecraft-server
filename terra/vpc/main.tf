provider "aws" {
  region = local.region
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "minecraft"
  cidr = "10.0.0.0/16"

  azs            = [local.availability_zone]
  public_subnets = ["10.0.101.0/24"]

  map_public_ip_on_launch = true
  create_igw              = true
}

