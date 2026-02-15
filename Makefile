COMPOSE_FILE := setup/wsl/docker-compose.yml
INIT_SCRIPT := ./setup/wsl/init.sh
CONFIGURE_PAPER_SCRIPT := ./setup/wsl/configure-paper.sh

.PHONY: init up down restart ps logs logs-velocity logs-lobby logs-survival sync-secret configure-paper bootstrap

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
