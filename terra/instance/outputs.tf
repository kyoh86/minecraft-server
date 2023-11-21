output "ip" {
  value = aws_eip.ip.public_ip
}

output "instance" {
  value = aws_instance.instance.id
}

output "security_group" {
  value = aws_security_group.instance.id
}

output "public_subnet" {
  value = module.vpc.public_subnets[0]
}
