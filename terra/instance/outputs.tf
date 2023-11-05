output "ip" {
  value = aws_eip.ip.public_ip
}

output "instance" {
  value = aws_instance.instance.id
}
