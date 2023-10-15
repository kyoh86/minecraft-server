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

resource "aws_vpc" "minecraft" {
  enable_dns_support   = true
  enable_dns_hostnames = true
  cidr_block           = "10.0.0.0/16"
  tags = {
    Name = "minecraft_vpc"
  }
}

resource "aws_subnet" "public_subnet_ap_northeast_1a" {
  availability_zone    = "ap-northeast-1a"
  availability_zone_id = data.aws_availability_zone.a.id
  vpc_id               = aws_vpc.minecraft.id
  cidr_block           = "10.0.1.0/24"
  tags = {
    Name = "minecraft_public_subnet_ap_northeast_1a"
  }
}

resource "aws_subnet" "private_subnet_ap_northeast_1a" {
  availability_zone    = "ap-northeast-1a"
  availability_zone_id = data.aws_availability_zone.a.id
  vpc_id               = aws_vpc.minecraft.id
  cidr_block           = "10.0.0.0/24"
  tags = {
    Name = "minecraft_private_subnet_ap_northeast_1a"
  }
}

resource "aws_internet_gateway" "main_gw" {
  vpc_id = aws_vpc.minecraft.id
  tags = {
    Name = "minecraft_main_gw"
  }
}

resource "aws_route_table" "rtb_public" {
  vpc_id = aws_vpc.minecraft.id
  route = [{
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main-gw.id
  }]
  tags = {
    Name = "minecraft_route_table"
  }
}

resource "aws_route_table_association" "rtb_assoc_ap_northeast_1a" {
  subnet_id      = aws_subnet.public-subnet-ap-northeast-1a.id
  route_table_id = aws_route_table.rtb-public
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
    Name = "minecraft_sg_ingress_allow_java_client"
  }
}

resource "aws_instance" "minecraft_java" {
  subnet_id = aws_subnet.public-subnet-ap-northeast-1a.id
  security_groups = [
    aws_security_group.aws_security_group.sg_java_server
  ]
  tags = {
    Name = "minecraft_java"
  }
}
// - EBSの割り当て
// - EBSのバックアップ設定
// - Instance Profileを作成、SSM許可のロールをアタッチ、Instanceにプロファイルを設定
// - Elastic IPの割り当て
