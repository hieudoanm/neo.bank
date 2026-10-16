.PHONY: help build test lint typecheck clean

help:
	@echo "Available targets: build, test, lint, typecheck, clean"

build:
	@for d in packages/backend/*; do (cd $$d && go build ./...); done
	@cd packages/frontend/app && pnpm build

test:
	@for d in packages/backend/*; do (cd $$d && go test ./...); done
	@cd packages/frontend/app && pnpm test

lint:
	@echo "lint stub - configure golangci-lint and eslint"

typecheck:
	@for d in packages/backend/*; do (cd $$d && go vet ./...); done
	@cd packages/frontend/app && pnpm typecheck

clean:
	@rm -rf packages/frontend/app/dist packages/frontend/app/node_modules/.cache
