GO_ENV := GOCACHE=/tmp/minecraft-server-go-cache GOPATH=/tmp/minecraft-server-go
WSLCTL := $(GO_ENV) go run ./cmd/wslctl
BOT_IMAGE := node:22-alpine
BOT_CONTAINER := mc-bot
BOT_AUTH ?= offline
BOT_VERSION ?=
BOT_USERNAME ?= codexbot

.PHONY: \
	setup-init \
	server-up server-down server-restart server-ps server-logs server-reload \
	world-ensure world-regenerate world-drop world-delete world-setup world-function \
	world-spawn-profile world-spawn-stage world-spawn-apply \
	player-op-grant player-op-revoke player-admin-grant player-admin-revoke \
	bot-up bot-down bot-test bot-test-core bot-report-latest bot-control bot-monitor

setup-init:
	$(WSLCTL) setup init

server-up:
	$(WSLCTL) server up

server-down:
	$(WSLCTL) server down

server-restart:
	$(WSLCTL) server restart world

server-ps:
	$(WSLCTL) server ps

server-logs:
	$(WSLCTL) server logs world

server-reload:
	$(WSLCTL) server reload

world-ensure:
	$(WSLCTL) world ensure

world-regenerate:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-regenerate WORLD=resource" && exit 1)
	$(WSLCTL) world regenerate $(WORLD)

world-drop:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-drop WORLD=resource" && exit 1)
	$(WSLCTL) world drop $(WORLD)

world-delete:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-delete WORLD=resource" && exit 1)
	$(WSLCTL) world delete --yes $(WORLD)

world-setup:
	@if [ -n "$(WORLD)" ]; then \
		$(WSLCTL) world setup --world $(WORLD); \
	else \
		$(WSLCTL) world setup; \
	fi

world-function:
	@test -n "$(FUNCTION)" || (echo "FUNCTION is required. e.g. make world-function FUNCTION=mcserver:world_settings" && exit 1)
	$(WSLCTL) world function run $(FUNCTION)

world-spawn-profile:
	$(WSLCTL) world spawn profile

world-spawn-stage:
	$(WSLCTL) world spawn stage

world-spawn-apply:
	$(WSLCTL) world spawn apply

player-op-grant:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-op-grant PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player op grant $(PLAYER)

player-op-revoke:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-op-revoke PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player op revoke $(PLAYER)

player-admin-grant:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-admin-grant PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player admin grant $(PLAYER)

player-admin-revoke:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-admin-revoke PLAYER=kyoh86" && exit 1)
	$(WSLCTL) player admin revoke $(PLAYER)

bot-up:
	@docker rm -f $(BOT_CONTAINER) >/dev/null 2>&1 || true
	@status=""; \
	for i in $$(seq 1 90); do \
		status=$$(docker inspect -f '{{.State.Health.Status}}' mc-world 2>/dev/null || true); \
		if [ "$$status" = "healthy" ]; then break; fi; \
		sleep 1; \
	done; \
	test "$$status" = "healthy" || (echo "mc-world is not healthy"; exit 1)
	$(WSLCTL) player op grant $(BOT_USERNAME)
	docker run -d --name $(BOT_CONTAINER) --network infra_mcnet --user $$(id -u):$$(id -g) \
		-v $(PWD)/tools/bot:/work \
		-v $(PWD)/runtime:/runtime \
		-w /work \
		$(BOT_IMAGE) sh -lc 'npm install --no-audit --no-fund && BOT_HOST=mc-world BOT_PORT=25565 BOT_AUTH=$(BOT_AUTH) BOT_VERSION=$(BOT_VERSION) BOT_USERNAME=$(BOT_USERNAME) BOT_REPORT_DIR=/runtime/bot-reports npm run idle'

bot-down:
	@docker rm -f $(BOT_CONTAINER) >/dev/null 2>&1 || true

bot-test:
	@test -n "$(SCENARIO)" || (echo "SCENARIO is required. e.g. make bot-test SCENARIO=portal_resource_to_mainhall" && exit 1)
	@status=0; \
	$(MAKE) --no-print-directory bot-down >/dev/null 2>&1 || true; \
	$(WSLCTL) server down >/dev/null 2>&1 || true; \
	ONLINE_MODE=FALSE $(WSLCTL) server up || exit $$?; \
	if ! $(MAKE) --no-print-directory bot-test-core SCENARIO="$(SCENARIO)" BOT_AUTH="$(BOT_AUTH)" BOT_VERSION="$(BOT_VERSION)" BOT_USERNAME="$(BOT_USERNAME)"; then status=$$?; fi; \
	$(MAKE) --no-print-directory bot-down >/dev/null 2>&1 || true; \
	$(WSLCTL) server down >/dev/null 2>&1 || true; \
	ONLINE_MODE=TRUE $(WSLCTL) server up || exit $$?; \
	exit $$status

bot-test-core:
	@test -n "$(SCENARIO)" || (echo "SCENARIO is required. e.g. make bot-test-core SCENARIO=portal_resource_to_mainhall" && exit 1)
	@status=""; \
	for i in $$(seq 1 90); do \
		status=$$(docker inspect -f '{{.State.Health.Status}}' mc-world 2>/dev/null || true); \
		if [ "$$status" = "healthy" ]; then break; fi; \
		sleep 1; \
	done; \
	test "$$status" = "healthy" || (echo "mc-world is not healthy"; exit 1)
	$(WSLCTL) player op grant $(BOT_USERNAME)
	docker run --rm --network infra_mcnet --user $$(id -u):$$(id -g) \
		-v $(PWD)/tools/bot:/work \
		-v $(PWD)/runtime:/runtime \
		-w /work \
		$(BOT_IMAGE) sh -lc 'npm install --no-audit --no-fund && BOT_HOST=mc-world BOT_PORT=25565 BOT_AUTH=$(BOT_AUTH) BOT_VERSION=$(BOT_VERSION) BOT_USERNAME=$(BOT_USERNAME) BOT_REPORT_DIR=/runtime/bot-reports BOT_SCENARIO=$(SCENARIO) npm run test'

bot-report-latest:
	@latest=$$(ls -1t runtime/bot-reports/*.json 2>/dev/null | head -n1); \
	if [ -z "$$latest" ]; then \
		echo "no report found under runtime/bot-reports"; \
		exit 1; \
	fi; \
	echo "$$latest"; \
	cat "$$latest"

bot-control:
	docker run --rm -i --network infra_mcnet --user $$(id -u):$$(id -g) \
		-v $(PWD)/tools/bot:/work \
		-v $(PWD)/runtime:/runtime \
		-w /work \
		$(BOT_IMAGE) sh -lc 'npm install --no-audit --no-fund && BOT_HOST=mc-world BOT_PORT=25565 BOT_AUTH=$(BOT_AUTH) BOT_VERSION=$(BOT_VERSION) BOT_USERNAME=$(BOT_USERNAME) BOT_REPORT_DIR=/runtime/bot-reports npm run control'

bot-monitor:
	docker run --rm --network infra_mcnet --user $$(id -u):$$(id -g) \
		-v $(PWD)/tools/bot:/work \
		-v $(PWD)/runtime:/runtime \
		-w /work \
		$(BOT_IMAGE) sh -lc 'npm install --no-audit --no-fund && BOT_HOST=mc-world BOT_PORT=25565 BOT_AUTH=$(BOT_AUTH) BOT_VERSION=$(BOT_VERSION) BOT_USERNAME=$(BOT_USERNAME) BOT_REPORT_DIR=/runtime/bot-reports BOT_MONITOR_SCENARIO=$${BOT_MONITOR_SCENARIO:-smoke} BOT_MONITOR_INTERVAL_SEC=$${BOT_MONITOR_INTERVAL_SEC:-300} npm run monitor'
