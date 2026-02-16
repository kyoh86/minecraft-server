COMPOSE_FILE := setup/wsl/docker-compose.yml
INIT_SCRIPT := ./setup/wsl/init.sh
APPLY_LOBBY_SETTINGS_SCRIPT := ./setup/wsl/apply-lobby-settings.sh
SYNC_LOBBY_DATAPACK_SCRIPT := ./setup/wsl/sync-lobby-datapack.sh

.PHONY: init up down restart ps logs logs-lobby bootstrap op-lobby deop-lobby lp-admin lobby-datapack-sync lobby-apply

init:
	$(INIT_SCRIPT)

up:
	docker compose -f $(COMPOSE_FILE) up -d --remove-orphans

down:
	docker compose -f $(COMPOSE_FILE) down

restart:
	docker compose -f $(COMPOSE_FILE) restart lobby

ps:
	docker compose -f $(COMPOSE_FILE) ps

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

logs-lobby:
	docker compose -f $(COMPOSE_FILE) logs -f lobby

bootstrap:
	$(INIT_SCRIPT)
	docker compose -f $(COMPOSE_FILE) up -d --remove-orphans
	docker compose -f $(COMPOSE_FILE) restart lobby

op-lobby:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make op-lobby PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 lobby mc-send-to-console "op $(PLAYER)"

deop-lobby:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make deop-lobby PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 lobby mc-send-to-console "deop $(PLAYER)"

lp-admin:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make lp-admin PLAYER=kyoh86" && exit 1)
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 lobby mc-send-to-console "lp creategroup admin"
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 lobby mc-send-to-console "lp group admin permission set * true"
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 lobby mc-send-to-console "lp user $(PLAYER) parent set admin"

lobby-datapack-sync:
	$(SYNC_LOBBY_DATAPACK_SCRIPT)

lobby-apply:
	$(APPLY_LOBBY_SETTINGS_SCRIPT)
