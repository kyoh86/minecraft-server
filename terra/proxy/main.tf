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
  cloud {
    organization = "kyoh86-org"
    workspaces {
      project = "minecraft"
      name    = "minecraft-proxy"
    }
  }
}

data "terraform_remote_state" "iam" {
  backend = "remote"
  config = {
    organization = "kyoh86-org"
    workspaces = {
      name = "minecraft-iam"
    }
  }
}

data "terraform_remote_state" "instance" {
  backend = "remote"
  config = {
    organization = "kyoh86-org"
    workspaces = {
      name = "minecraft-instance"
    }
  }
}
