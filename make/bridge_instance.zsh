#!/bin/zsh

INSTANCE_ID="$(./instance_id.zsh)"
aws ssm start-session --target "${INSTANCE_ID}" --document-name AWS-StartPortForwardingSession --parameters '{"portNumber":["25575"],"localPortNumber":["25575"]}'
