#!env zsh

VPC_NAME="minecraft"
INSTANCE_NAME="instance"

VPCID="$(aws ec2 describe-vpcs --query 'Vpcs[?Tags[?Key==`Name`].Value|[0]==`'$VPC_NAME'`]|[0].VpcId' --output text)"
aws ec2 describe-instances --query 'Reservations[].Instances[?Tags[?Key==`Name`].Value|[0]==`'$INSTANCE_NAME'`]|[0]|[0].InstanceId' --output text
