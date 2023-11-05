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
      name    = "minecraft-volume"
    }
  }
}
