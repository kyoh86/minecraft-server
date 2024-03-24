#!/bin/zsh

terraform -chdir=../terra/instance output -json | jq -r '.instance.value'
