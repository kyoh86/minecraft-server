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
	$(WSLCTL) worlds-bootstrap

world-reset:
	$(WSLCTL) world-reset

resource-reset:
	WORLD=resource $(WSLCTL) world-reset

op-world:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make op-world PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 world mc-send-to-console "op $(PLAYER)"

deop-world:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make deop-world PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 world mc-send-to-console "deop $(PLAYER)"

lp-admin:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make lp-admin PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 world mc-send-to-console "lp creategroup admin"
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 world mc-send-to-console "lp group admin permission set * true"
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 world mc-send-to-console "lp user $(PLAYER) parent set admin"

lp-reset:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make lp-reset PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 world mc-send-to-console "lp user $(PLAYER) parent remove admin"

world-datapack-sync:
	$(WSLCTL) sync-world-datapack

world-apply:
	$(WSLCTL) apply-world-settings
