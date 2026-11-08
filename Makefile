.PHONY: help build test lint typecheck clean

help:
	@echo "Available targets: build, test, lint, typecheck, clean"

build:
	@for d in packages/backend/*; do (cd $$d && go build ./...); done
	@cd packages/frontend/app && pnpm build

test:
	@bash scripts/coverage/backend.sh --all
	@cd packages/frontend/app && pnpm test

lint:
	@for d in packages/backend/*; do \
		(cd $$d && test -z "$$(gofmt -l .)" && go vet ./...) || exit 1; \
	done
	@cd packages/frontend/app && pnpm lint && pnpm format:check

typecheck:
	@for d in packages/backend/*; do (cd $$d && go vet ./...); done
	@cd packages/frontend/app && pnpm typecheck

clean:
	@rm -rf packages/frontend/app/dist packages/frontend/app/node_modules/.cache
