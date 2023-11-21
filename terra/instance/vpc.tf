module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "minecraft"
  cidr = "10.0.0.0/16"

  azs            = [local.availability_zone]
  public_subnets = ["10.0.101.0/24"]

  map_public_ip_on_launch     = true
  create_igw                  = true
  default_security_group_name = "sg_default"
}

resource "aws_eip" "ip" {
  instance = aws_instance.instance.id
  domain   = "vpc"
  tags = {
    Name = "ip"
  }
}

resource "aws_security_group" "instance" {
  vpc_id = module.vpc.vpc_id
  name   = "sg_instance"
}

resource "aws_vpc_security_group_ingress_rule" "allow_internal_session" {
  referenced_security_group_id = module.vpc.default_security_group_id
  ip_protocol                  = "-1"
  security_group_id            = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_ingress"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_java_client" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = "25565"
  to_port           = "25565"
  security_group_id = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_java_client_ingress"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_bedrock_connect" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
  from_port         = "53"
  to_port           = "53"
  security_group_id = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_bedrock_client_ingress"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_bedrock_client" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "udp"
  from_port         = "19132"
  to_port           = "19132"
  security_group_id = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_bedrock_client_ingress"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_dynmap" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = "8123"
  to_port           = "8123"
  security_group_id = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_dynmap_ingress"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_http" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = "80"
  to_port           = "80"
  security_group_id = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_http_ingress"
  }
}

resource "aws_vpc_security_group_egress_rule" "allow_egress_all" {
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
  security_group_id = aws_security_group.instance.id
  tags = {
    Name = "sg_allow_egress_all"
  }
}
