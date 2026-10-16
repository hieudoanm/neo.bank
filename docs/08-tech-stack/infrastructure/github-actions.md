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
        service: [api-gateway, auth, accounts, transfer, transactions]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      - run: cd packages/backend/${{ matrix.service }} && golangci-lint run
      - run: cd packages/backend/${{ matrix.service }} && go test ./... -coverprofile=coverage.out

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 8
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm lint && pnpm typecheck && pnpm test && pnpm build
```

Rules:

- Independent jobs (`backend`, `frontend`) run in parallel; keep them independent.
- The backend job is a **matrix over the independent Go modules** under `packages/backend/` (the gateway plus each service, named by its directory: `api-gateway`, `auth`, `accounts`, `transfer`, `transactions`); there is no single Go module at `packages/backend/`.
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
        service: [auth, accounts, transfer, transactions]
    services:
      cassandra:
        image: cassandra:4.1
        ports: ['9042:9042']
        options: >-
          --health-cmd "cqlsh -e 'describe keyspaces'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 12
      zookeeper:
        image: confluentinc/cp-zookeeper:7.4.0
        env:
          ZOOKEEPER_CLIENT_PORT: 2181
        options: >-
          --health-cmd "bash -c 'echo srvr | nc -w 2 localhost 2181'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 12
      kafka:
        image: confluentinc/cp-kafka:7.4.0
        ports: ['9092:9092']
        env:
          KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
          KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092
          KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
          KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
          KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
          KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
          KAFKA_DEFAULT_REPLICATION_FACTOR: 1
          KAFKA_MIN_INSYNC_REPLICAS: 1
          KAFKA_AUTO_CREATE_TOPICS_ENABLE: "false"
        options: >-
          --health-cmd "kafka-topics --bootstrap-server localhost:9092 --list"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 12
    steps:
      - uses: actions/checkout@v4
      - run: cd packages/backend/${{ matrix.service }} && go run cmd/migrate/main.go
      - run: cd packages/backend/${{ matrix.service }} && go test ./...

Rules:

- Never replace Cassandra or Kafka with mocks in integration tests; the service is the contract.
- Use `options` health checks so the job waits for readiness, not a fixed sleep.
- **Kafka in CI needs a `zookeeper` service** and `KAFKA_ZOOKEEPER_CONNECT` — `confluentinc/cp-kafka` in ZooKeeper mode will not start without it, and KRaft settings (`CLUSTER_ID`/`KAFKA_NODE_ID`) must not be mixed with `KAFKA_ZOOKEEPER_CONNECT`.
- CI runs a **single broker at RF=1** (`DEFAULT_REPLICATION_FACTOR=1`, `OFFSETS_TOPIC_REPLICATION_FACTOR=1`, `MIN_INSYNC_REPLICAS=1`); internal topics with the default RF=3 never come up on one node. Production keeps RF=3 / `min.insync.replicas=2` (see [`NFR-013`](../../01-requirements/non-functional-requirements.md#availability)).
- Create the `transfers` and `transfers.dlq` topics explicitly (auto-create disabled) before tests that publish to them.
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
        with: { node-version: 22, cache: pnpm }
      - run: pnpm install --frozen-lockfile
      - run: pnpm exec playwright install --with-deps
      - run: pnpm test:e2e
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: packages/frontend/app/playwright-report/
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
  plan:
    needs: [backend, frontend, integration]
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: |
          cd infrastructure/aws
          terraform init
          terraform fmt -check -recursive
          terraform validate
          terraform plan -out=tfplan
      - uses: actions/upload-artifact@v4
        with: { name: tfplan, path: infrastructure/aws/tfplan }

  deploy:
    needs: [plan]
    environment: production
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - uses: actions/download-artifact@v4
        with: { name: tfplan, path: infrastructure/aws }
      # Applies the plan produced by `plan` — never -auto-approve.
      - run: cd infrastructure/aws && terraform init && terraform apply tfplan
      - run: aws eks update-kubeconfig --name neo-bank-production
      - run: kubectl apply -k infrastructure/kubernetes/overlays/production
      - run: kubectl rollout status deployment/api-gateway -n neo-bank
```

Rules:

- Deploy `needs:` every test job; never deploy on red.
- Use a protected `environment` so production requires review — and make the reviewer see the plan: produce `terraform plan -out=tfplan` in an un-gated job, hand the artifact to the gated job, and apply **that file** (`terraform apply tfplan`). Never `terraform apply -auto-approve`, and never plan+apply in one unreviewed step.
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
