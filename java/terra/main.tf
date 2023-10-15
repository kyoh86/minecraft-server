provider "aws" {
  region = "ap-northeast-1"
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
  source = "terraform-aws-modules/vpc/aws"

  name = "minecraft-java"
  cidr = "10.0.0.0/16"

  azs            = ["ap-northeast-1a"]
  public_subnets = ["10.0.101.0/24"]

  create_igw = true
}

resource "aws_security_group" "sg_java_server" {
  vpc_id = aws_vpc.minecraft.id
  name   = "minecraft_sg_java_server"
}

resource "aws_vpc_security_group_ingress_rule" "allow_java_client" {
  ip_protocol       = "tcp"
  from_port         = "25565"
  to_port           = "25565"
  security_group_id = aws_security_group.sg-java-server.id
  tags = {
    Name = "minecraft_sg__allow_java_client_ingress"
  }
}

resource "aws_vpc_security_group_egress_rule" "allow_egress_all" {
  ip_protocol       = "-1"
  security_group_id = aws_security_group.sg-java-server.id
  tags = {
    Name = "minecraft_sg_allow_egress_all"
  }
}

resource "aws_instance" "minecraft_java" {
  subnet_id = module.vpc.public_subnets[0]
  security_groups = [
    module.vpc.default_security_group_id,
    aws_security_group.aws_security_group.sg_java_server.id
  ]
  tags = {
    Name = "minecraft_java"
  }
}

// - EBSの割り当て
// - EBSのバックアップ設定
// - Instance Profileを作成、SSM許可のロールをアタッチ、Instanceにプロファイルを設定
// - Elastic IPの割り当て
