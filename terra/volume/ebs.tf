resource "aws_ebs_volume" "instance_data" {
  availability_zone = local.availability_zone
  size              = 16 // GiB
  tags = {
    Name = "instance_data"
  }
}

// TODO: EBSのバックアップ設定
