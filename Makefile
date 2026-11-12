.PHONY: help build test lint typecheck clean compose-up compose-down compose-logs compose-ps

help:
	@echo "Available targets: build, test, lint, typecheck, clean, compose-up, compose-down, compose-logs, compose-ps"
	@echo "  compose-up      Start the full local stack (docker compose up -d --build)"
	@echo "  compose-down    Stop the local stack (docker compose down)"
	@echo "  compose-logs    Tail logs from the local stack"
	@echo "  compose-ps      Show running stack containers"

build:
	@for d in works/*; do (cd $$d && go build ./...); done
	@cd packages/app && pnpm build

test:
	@bash scripts/coverage/backend.sh --all
	@cd packages/app && pnpm test

lint:
	@for d in works/*; do \
		(cd $$d && test -z "$$(gofmt -l .)" && go vet ./...) || exit 1; \
	done
	@cd packages/app && pnpm lint && pnpm format:check

typecheck:
	@for d in works/*; do (cd $$d && go vet ./...); done
	@cd packages/app && pnpm typecheck

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f --tail=100

compose-ps:
	docker compose ps

clean:
	@rm -rf packages/app/dist packages/app/node_modules/.cache
