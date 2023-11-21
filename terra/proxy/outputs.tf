output "ip" {
  value = aws_eip.proxy_ip.public_ip
}
