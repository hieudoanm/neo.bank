# Contributing to neo.bank

Thanks for contributing to neo.bank, an event-driven banking microservices
demonstration: Go services plus a Next.js frontend, with normative specs in
`docs/`.

## Repository layout

- `packages/` — Go microservices and the web frontend (pnpm workspace).
- `docs/` — requirements, features, data model, API spec, UI/UX, tech stack,
  testing, and deployment docs.
- `docker-compose.yaml` — local topology.

See [AGENTS.md](AGENTS.md) for contributor conventions.

## Setup

```bash
git clone https://github.com/hieudoanm/neo.bank.git
cd neo.bank
docker compose up
```

Or run services individually — see `make help` for targets (`build`, `test`,
`lint`, `typecheck`, `compose-up`).

## Checks

```bash
make lint
go test ./...
pnpm typecheck && pnpm lint && pnpm test && pnpm build   # frontend
```

## Conventions

- `docs/` holds the normative specifications — keep code and docs in sync.
- Follow existing service boundaries; do not merge services casually.
- Exact dependency versions are preferred over floating ranges.

## License

By contributing, you agree that your contributions will be licensed under the
repository license in [`LICENSE`](LICENSE) (GPL-3.0).
