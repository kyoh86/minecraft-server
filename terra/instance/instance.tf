data "aws_ssm_parameter" "amzn2_ami" {
  name = "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
}

resource "aws_instance" "instance" {
  ami                  = data.aws_ssm_parameter.amzn2_ami.value
  iam_instance_profile = data.terraform_remote_state.iam.outputs.instance_profile_name
  availability_zone    = local.availability_zone
  instance_type        = "t3.medium"
  subnet_id            = module.vpc.public_subnets[0]
  vpc_security_group_ids = [
    aws_security_group.instance.id
  ]
  tags = {
    Name = "instance"
  }
}

resource "aws_volume_attachment" "instance_volume_attachment" {
  device_name = "/dev/sdh"
  instance_id = aws_instance.instance.id
  volume_id   = data.terraform_remote_state.volume.outputs.volume_id
}
