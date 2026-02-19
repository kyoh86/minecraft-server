GO_ENV := GOCACHE=/tmp/minecraft-server-go-cache GOPATH=/tmp/minecraft-server-go
WSLCTL := $(GO_ENV) go run ./cmd/wslctl

.PHONY: \
	setup-init \
	server-up server-down server-restart server-ps server-logs server-reload \
	world-ensure world-regenerate world-drop world-delete world-setup world-function \
	world-spawn-profile world-spawn-stage world-spawn-apply \
	player-op-grant player-op-revoke player-admin-grant player-admin-revoke

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
