#!/bin/zsh

INSTANCE_ID="$(./instance_id.zsh)"
ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"
