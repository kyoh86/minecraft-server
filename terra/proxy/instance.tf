data "aws_ssm_parameter" "amzn2_ami" {
  name = "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
}

resource "aws_eip" "proxy_ip" {
  instance = aws_instance.proxy.id
  domain   = "vpc"
  tags = {
    Name = "proxy-ip"
  }
}

resource "aws_instance" "proxy" {
  ami                  = data.aws_ssm_parameter.amzn2_ami.value
  iam_instance_profile = data.terraform_remote_state.iam.outputs.instance_profile_name
  availability_zone    = local.availability_zone
  instance_type        = "t3.small"
  subnet_id            = "subnet-077a7a091ef941c07" # data.terraform_remote_state.instance.outputs.public_subnet
  vpc_security_group_ids = [
    data.terraform_remote_state.instance.outputs.security_group
  ]
  tags = {
    Name = "proxy"
  }
}
