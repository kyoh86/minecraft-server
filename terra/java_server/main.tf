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

data "terraform_remote_state" "iam" {
  backend = "local"
  config = {
    path = "../iam/terraform.tfstate"
  }
}

data "terraform_remote_state" "vpc" {
  backend = "local"
  config = {
    path = "../vpc/terraform.tfstate"
  }
}

data "aws_ssm_parameter" "amzn2_ami" {
  name = "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
}

resource "aws_security_group" "sg_java" {
  vpc_id = data.terraform_remote_state.vpc.outputs.vpc_id
  name   = "minecraft_sg_java"
}

resource "aws_vpc_security_group_ingress_rule" "allow_internal_session" {
  referenced_security_group_id = data.terraform_remote_state.vpc.outputs.vpc_default_security_group
  ip_protocol                  = "-1"
  security_group_id            = aws_security_group.sg_java.id
  tags = {
    Name = "minecraft_sg_allow_java_client_ingress"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_java_client" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = "25565"
  to_port           = "25565"
  security_group_id = aws_security_group.sg_java.id
  tags = {
    Name = "minecraft_sg_allow_java_client_ingress"
  }
}

resource "aws_vpc_security_group_egress_rule" "allow_egress_all" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
  security_group_id = aws_security_group.sg_java.id
  tags = {
    Name = "minecraft_sg_allow_egress_all"
  }
}

resource "aws_instance" "minecraft_java" {
  ami                  = data.aws_ssm_parameter.amzn2_ami.value
  iam_instance_profile = data.terraform_remote_state.iam.outputs.java_server_profile_name
  availability_zone    = local.availability_zone
  instance_type        = "t3.medium"
  subnet_id            = data.terraform_remote_state.vpc.outputs.vpc_public_subnet
  vpc_security_group_ids = [
    aws_security_group.sg_java.id
  ]
  tags = {
    Name = "minecraft_java"
  }
}

resource "aws_ebs_volume" "minecraft_java_data" {
  availability_zone = local.availability_zone
  size              = 8 // GiB
  tags = {
    Name = "minecraft_java_data"
  }
}

resource "aws_volume_attachment" "minecraft_java_data" {
  device_name = "/dev/sdh"
  instance_id = aws_instance.minecraft_java.id
  volume_id   = aws_ebs_volume.minecraft_java_data.id
}

resource "aws_eip" "minecraft_java" {
  instance = aws_instance.minecraft_java.id
  domain   = "vpc"
  tags = {
    Name = "minecraft_java"
  }
}

output "ip" {
  value = aws_eip.minecraft_java.public_ip
}

output "instance" {
  value = aws_instance.minecraft_java.id
}

// - EBSのバックアップ設定
