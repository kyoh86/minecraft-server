COMPOSE_FILE := setup/wsl/docker-compose.yml
GO_ENV := GOCACHE=/tmp/minecraft-server-go-cache GOPATH=/tmp/minecraft-server-go
WSLCTL := $(GO_ENV) go run ./cmd/wslctl

.PHONY: init up down restart ps logs logs-world bootstrap worlds-bootstrap world-reset resource-reset op-world deop-world lp-admin lp-reset world-datapack-sync world-apply

init:
	$(WSLCTL) init

up:
	docker compose -f $(COMPOSE_FILE) up -d --remove-orphans

down:
	docker compose -f $(COMPOSE_FILE) down

restart:
	docker compose -f $(COMPOSE_FILE) restart world

ps:
	docker compose -f $(COMPOSE_FILE) ps

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

logs-world:
	docker compose -f $(COMPOSE_FILE) logs -f world

bootstrap:
	$(WSLCTL) init
	docker compose -f $(COMPOSE_FILE) up -d --remove-orphans
	docker compose -f $(COMPOSE_FILE) restart world

worlds-bootstrap:
	$(WSLCTL) world bootstrap

world-reset:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-reset WORLD=resource" && exit 1)
	$(WSLCTL) world reset $(WORLD)

resource-reset:
	$(WSLCTL) world reset resource

op-world:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make op-world PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player op $(PLAYER)

deop-world:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make deop-world PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player deop $(PLAYER)

lp-admin:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make lp-admin PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player perms grant-admin $(PLAYER)

lp-reset:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make lp-reset PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player perms revoke-admin $(PLAYER)

world-datapack-sync:
	$(WSLCTL) datapack sync

world-apply:
	$(WSLCTL) world apply-settings
