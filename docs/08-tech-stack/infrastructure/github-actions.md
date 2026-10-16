# NEO Bank — GitHub Actions Best Practices

_Infrastructure · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [infrastructure index](./README.md)._

GitHub Actions runs the quality gates and the optional cloud deploy. The full workflow definitions live in [`11-deployment.md`](../../11-deployment.md).

## 1. Workflow Shape

Separate, focused workflows rather than one mega-pipeline.

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true
```

Rules:

- One workflow per concern: `ci.yml` (lint/test), integration, `e2e.yml`, `deploy-aws.yml`, `deploy-gcp.yml`.
- Trigger on both `push` and `pull_request` for the main branches.
- Set `concurrency` with `cancel-in-progress` so a new push supersedes the old run.
- Pin action versions (`actions/checkout@v4`); never use a floating ref in a security-sensitive step.

## 2. Jobs & Parallelism

```yaml
jobs:
  backend:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [api-gateway, auth-service, account-service, transfer-service, transaction-service]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      - run: cd packages/${{ matrix.service }} && golangci-lint run
      - run: cd packages/${{ matrix.service }} && go test ./... -coverprofile=coverage.out

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 8
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm lint && pnpm typecheck && pnpm test && pnpm build
```

Rules:

- Independent jobs (`backend`, `frontend`) run in parallel; keep them independent.
- The backend job is a **matrix over the independent Go modules** under `packages/` (the gateway and each service); there is no single `packages/backend` module.
- Cache dependencies with the built-in `cache:` option (`go`, `pnpm`); pin the tool version.
- `pnpm install --frozen-lockfile` — CI never updates the lock file.
- Fail on the first real error; do not mask exit codes with `|| true`.

## 3. Integration with Real Services

Integration and consumer tests need real Cassandra and Kafka, provided as service containers.

```yaml
jobs:
  integration:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [auth-service, account-service, transfer-service, transaction-service]
    services:
      cassandra:
        image: cassandra:4.1
        ports: ['9042:9042']
        options: >-
          --health-cmd "cqlsh -e 'describe keyspaces'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 12
      kafka:
        image: confluentinc/cp-kafka:7.4.0
        ports: ['9092:9092']
    steps:
      - uses: actions/checkout@v4
      - run: cd packages/${{ matrix.service }} && go run cmd/migrate/main.go
      - run: cd packages/${{ matrix.service }} && go test ./...

Rules:

- Never replace Cassandra or Kafka with mocks in integration tests; the service is the contract.
- Use `options` health checks so the job waits for readiness, not a fixed sleep.
- Apply migrations before running integration tests; each service migrates only its own keyspace.
- Keep the service image tags identical to local Compose and production providers.

## 4. End-to-End

```yaml
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with: { version: 8 }
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: pnpm }
      - run: pnpm install --frozen-lockfile
      - run: pnpm exec playwright install --with-deps
      - run: pnpm test:e2e
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: packages/frontend/playwright-report/
```

Rules:

- E2E is the last gate; it depends on a built frontend and a running backend.
- Always upload the report (`if: always()`) so failures are diagnosable.
- Keep the E2E job under 10 minutes (see [`playwright.md`](../frontend/testing/playwright.md)).

## 5. Security & Permissions

- Set `permissions:` to the minimum (usually `contents: read`); grant more only per job that needs it.
- Prefer **OIDC** (`id-token: write`) over long-lived cloud keys for deploys.
- Store credentials as encrypted secrets; never echo them or pass them through `env` that logs.
- Run Trivy (filesystem/image) and upload SARIF; a high-severity finding blocks the merge.

## 6. Deploy Gates & Caching

```yaml
  deploy:
    needs: [backend, frontend, integration]
    environment: production
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: cd infrastructure/aws && terraform init && terraform plan -out=tfplan && terraform apply tfplan
      - run: aws eks update-kubeconfig --name neo-bank-production
      - run: kubectl apply -k infrastructure/kubernetes/overlays/production
      - run: kubectl rollout status deployment/api-gateway -n neo-bank
```

Rules:

- Deploy `needs:` every test job; never deploy on red.
- Use a protected `environment` so production requires review.
- Validate manifests (`kubectl apply --dry-run=server`) in the PR job, before the deploy job runs.
- Cache what is expensive and deterministic (Go modules, pnpm store, Docker layers); never cache secrets.
- Tag images by commit SHA so a deploy maps to an exact revision and rollback is a pointer change (`kubectl set image` / `rollout undo`).

## 7. Code Style & Rules

- Focused workflows; `concurrency` with `cancel-in-progress`; pinned actions.
- Parallel independent jobs; frozen lock file; dependency caching.
- Real Cassandra/Kafka as service containers with health gating; migrations before tests.
- Least-privilege `permissions`; OIDC for deploys; secrets never logged.
- Deploy gated with `needs` on green tests and a protected environment.
- Failures upload artifacts (Playwright report, SARIF) for diagnosis.
- Related: [`docker.md`](./docker.md) for images, [`kubernetes.md`](./kubernetes.md) for what gets applied, [`terraform.md`](./terraform.md) for the apply step, [`11-deployment.md`](../../11-deployment.md) for full workflow definitions.
