#!/bin/zsh

SSO_ACCOUNT=$(aws sts get-caller-identity --query "Account" --profile kyoh86-admin)
#you can add a better check, but this is just an idea for quick check
if [ ${#SSO_ACCOUNT} -eq 14 ];  then 
  echo "session still valid" ;
else 
  echo "Seems like session expired"
  aws sso login
fi
