login:
	@echo "Login to AWS"
	zsh ./aws_login.zsh
.PHONY: login

bridge: login
	@echo "Create a bridge for the instance"
	zsh ./bridge_instance.zsh 25575
.PHONY: bridge

connect: login
	@echo "Connect to the instance"
	zsh ./connect_instance.zsh
.PHONY: connect

mcrcon:
	@echo "Install mcrcon"
	zsh ./mcrcon_instance.zsh
