COMPOSE_FILE := setup/wsl/docker-compose.yml
INIT_SCRIPT := ./setup/wsl/init.sh
CONFIGURE_PAPER_SCRIPT := ./setup/wsl/configure-paper.sh
APPLY_LOBBY_SETTINGS_SCRIPT := ./setup/wsl/apply-lobby-settings.sh
APPLY_LOBBY_GATE_SCRIPT := ./setup/wsl/apply-lobby-gate.sh
SYNC_LOBBY_DATAPACK_SCRIPT := ./setup/wsl/sync-lobby-datapack.sh
INSTALL_GATEBRIDGE_PLUGIN_SCRIPT := ./setup/wsl/install-gatebridge-plugin.sh

.PHONY: init up down restart ps logs logs-velocity logs-lobby logs-survival sync-secret configure-paper bootstrap op-lobby deop-lobby lp-admin lobby-datapack-sync lobby-apply lobby-gate-apply gatebridge-plugin-install lobby-gate-plugin-install

init:
	$(INIT_SCRIPT)

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

restart:
	docker compose -f $(COMPOSE_FILE) restart velocity lobby survival

ps:
	docker compose -f $(COMPOSE_FILE) ps

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

logs-velocity:
	docker compose -f $(COMPOSE_FILE) logs -f velocity

logs-lobby:
	docker compose -f $(COMPOSE_FILE) logs -f lobby

logs-survival:
	docker compose -f $(COMPOSE_FILE) logs -f survival

sync-secret:
	$(CONFIGURE_PAPER_SCRIPT) secret-only

configure-paper:
	$(CONFIGURE_PAPER_SCRIPT) all

bootstrap:
	$(INIT_SCRIPT)
	docker compose -f $(COMPOSE_FILE) up -d
	$(CONFIGURE_PAPER_SCRIPT) all
	docker compose -f $(COMPOSE_FILE) restart velocity lobby survival

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
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 survival mc-send-to-console "lp creategroup admin"
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 survival mc-send-to-console "lp group admin permission set * true"
	docker compose -f $(COMPOSE_FILE) exec -T --user 1000 survival mc-send-to-console "lp user $(PLAYER) parent set admin"

lobby-datapack-sync:
	$(SYNC_LOBBY_DATAPACK_SCRIPT)

lobby-apply:
	$(APPLY_LOBBY_SETTINGS_SCRIPT)

lobby-gate-apply:
	$(APPLY_LOBBY_GATE_SCRIPT)

gatebridge-plugin-install:
	$(INSTALL_GATEBRIDGE_PLUGIN_SCRIPT)
	docker compose -f $(COMPOSE_FILE) up -d --force-recreate lobby

# Backward-compatible alias
lobby-gate-plugin-install: gatebridge-plugin-install
