output "vpc_public_subnet" {
  value = module.vpc.public_subnets[0]
}

output "vpc_default_security_group" {
  value = module.vpc.default_security_group_id
}

output "vpc_id" {
  value = module.vpc.vpc_id
}
