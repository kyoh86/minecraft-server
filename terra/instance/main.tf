provider "aws" {
  region = local.region
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
  cloud {
    organization = "kyoh86-org"
    workspaces {
      project = "minecraft"
      name    = "minecraft-instance"
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

data "terraform_remote_state" "volume" {
  backend = "remote"
  config = {
    organization = "kyoh86-org"
    workspaces = {
      name = "minecraft-volume"
    }
  }
}
