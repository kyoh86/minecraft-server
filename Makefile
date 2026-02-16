COMPOSE_FILE := setup/wsl/docker-compose.yml
INIT_SCRIPT := ./setup/wsl/init.sh
APPLY_WORLD_SETTINGS_SCRIPT := ./setup/wsl/apply-world-settings.sh
SYNC_WORLD_DATAPACK_SCRIPT := ./setup/wsl/sync-world-datapack.sh
WORLDS_BOOTSTRAP_SCRIPT := ./setup/wsl/worlds-bootstrap.sh
WORLD_RESET_SCRIPT := ./setup/wsl/world-reset.sh

.PHONY: init up down restart ps logs logs-world bootstrap worlds-bootstrap world-reset resource-reset op-world deop-world lp-admin lp-reset world-datapack-sync world-apply

init:
	$(INIT_SCRIPT)

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
	$(INIT_SCRIPT)
	docker compose -f $(COMPOSE_FILE) up -d --remove-orphans
	docker compose -f $(COMPOSE_FILE) restart world

worlds-bootstrap:
	$(WORLDS_BOOTSTRAP_SCRIPT)

world-reset:
	$(WORLD_RESET_SCRIPT)

resource-reset:
	WORLD=resource $(WORLD_RESET_SCRIPT)

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
	$(SYNC_WORLD_DATAPACK_SCRIPT)

world-apply:
	$(APPLY_WORLD_SETTINGS_SCRIPT)
