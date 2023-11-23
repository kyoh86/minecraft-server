# data "aws_ssm_parameter" "amzn2_ami" {
#   name = "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
# }

locals {
  key_name = "minecraft_instance"
  key_dir  = pathexpand("~/.ssh")
}

resource "null_resource" "ssh_keygen" {
  provisioner "local-exec" {
    // ssh-keygenコマンドは、キーファイルが既に存在する場合、上書きしないが、エラーで終わるのでexit 0を追加している。
    command    = "ssh-keygen -m PEM  -C '' -N '' -f ${local.key_dir}/${local.key_name}"
    on_failure = continue
  }
}

// local_file (Data Source)にて公開キーの内容を取得。content属性にて確認できる
// file("ファイル名")によるファイル内容取得は、apply実行時点でファイルが存在しないとエラーとなってしまい使用できなかった。
data "local_file" "ssh_keygen" {
  filename   = "${local.key_dir}/${local.key_name}.pub"
  depends_on = [null_resource.ssh_keygen]
}

resource "aws_key_pair" "instance_key" {
  key_name   = local.key_name
  public_key = data.local_file.ssh_keygen.content
}

resource "aws_instance" "instance" {
  ami                  = "ami-0220a6b98b70e1279" # data.aws_ssm_parameter.amzn2_ami.value
  iam_instance_profile = data.terraform_remote_state.iam.outputs.instance_profile_name
  availability_zone    = local.availability_zone
  instance_type        = "t3.large"
  subnet_id            = module.vpc.public_subnets[0]
  vpc_security_group_ids = [
    aws_security_group.instance.id
  ]
  key_name                             = aws_key_pair.instance_key.key_name
  hibernation                          = false
  instance_initiated_shutdown_behavior = "stop"
  ipv6_address_count                   = 0
  monitoring                           = true
  tags = {
    Name = "instance"
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 1
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "unlimited"
  }

  enclave_options {
    enabled = false
  }

  maintenance_options {
    auto_recovery = "default"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
    instance_metadata_tags      = "disabled"
  }

  private_dns_name_options {
    enable_resource_name_dns_a_record    = false
    enable_resource_name_dns_aaaa_record = false
    hostname_type                        = "ip-name"
  }

  root_block_device {
    delete_on_termination = true
    encrypted             = false
    tags                  = {}
    throughput            = 0
    volume_size           = 8
    volume_type           = "gp2"
  }


}

resource "aws_volume_attachment" "instance_volume_attachment" {
  device_name = "/dev/sdh"
  instance_id = aws_instance.instance.id
  volume_id   = data.terraform_remote_state.volume.outputs.volume_id
}
