GO_ENV := GOCACHE=/tmp/minecraft-server-go-cache GOPATH=/tmp/minecraft-server-go
MCCTL := $(GO_ENV) go run ./cmd/mcctl

.PHONY: \
	asset-init asset-stage \
	server-up server-down server-restart server-ps server-logs server-reload \
	world-ensure world-regenerate world-drop world-delete world-setup world-function \
	world-spawn-profile world-spawn-stage world-spawn-apply \
	player-op-grant player-op-revoke player-admin-grant player-admin-revoke

asset-init:
	$(MCCTL) asset init

asset-stage:
	$(MCCTL) asset stage

server-up:
	$(MCCTL) server up

server-down:
	$(MCCTL) server down

server-restart:
	$(MCCTL) server restart world

server-ps:
	$(MCCTL) server ps

server-logs:
	$(MCCTL) server logs world

server-reload:
	$(MCCTL) server reload

world-ensure:
	$(MCCTL) world ensure

world-regenerate:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-regenerate WORLD=resource" && exit 1)
	$(MCCTL) world regenerate $(WORLD)

world-drop:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-drop WORLD=resource" && exit 1)
	$(MCCTL) world drop $(WORLD)

world-delete:
	@test -n "$(WORLD)" || (echo "WORLD is required. e.g. make world-delete WORLD=resource" && exit 1)
	$(MCCTL) world delete --yes $(WORLD)

world-setup:
	@if [ -n "$(WORLD)" ]; then \
		$(MCCTL) world setup --world $(WORLD); \
	else \
		$(MCCTL) world setup; \
	fi

world-function:
	@test -n "$(FUNCTION)" || (echo "FUNCTION is required. e.g. make world-function FUNCTION=mcserver:world_settings" && exit 1)
	$(MCCTL) world function run $(FUNCTION)

world-spawn-profile:
	$(MCCTL) world spawn profile

world-spawn-stage:
	$(MCCTL) world spawn stage

world-spawn-apply:
	$(MCCTL) world spawn apply

player-op-grant:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-op-grant PLAYER=kyoh86" && exit 1)
	$(MCCTL) player op grant $(PLAYER)

player-op-revoke:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-op-revoke PLAYER=kyoh86" && exit 1)
	$(MCCTL) player op revoke $(PLAYER)

player-admin-grant:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-admin-grant PLAYER=kyoh86" && exit 1)
	$(MCCTL) player admin grant $(PLAYER)

player-admin-revoke:
	@test -n "$(PLAYER)" || (echo "PLAYER is required. e.g. make player-admin-revoke PLAYER=kyoh86" && exit 1)
	$(MCCTL) player admin revoke $(PLAYER)
