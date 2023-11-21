output "proxy" {
  value = aws_instance.proxy.id
}

output "ip" {
  value = aws_eip.proxy_ip.public_ip
}
