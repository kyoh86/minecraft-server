provider "aws" {
  region  = local.region
  profile = "kyoh86-iam-owner"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
