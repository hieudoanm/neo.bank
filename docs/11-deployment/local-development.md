# Local Development

## Prerequisites

- Docker and Docker Compose (v2.0+)
- Go (version declared in each service's `go.mod`; Go 1.22+ required)
- Node.js (version pinned in `packages/app/package.json` `engines` and CI)
- pnpm
- GNU Make (optional, for convenience commands)
- Git (v2.30+)

## Repository Structure

```text
neo.bank/
├── packages/
│   ├── backend/
│   │   ├── api-gateway/     # Go API gateway (routing, auth enforcement, rate limiting)
│   │   ├── auth/            # Go Auth Service
│   │   ├── accounts/        # Go Account Service
│   │   ├── transfer/        # Go Transfer Service — saga orchestrator
│   │   └── transactions/    # Go Transaction Service — read model
│   └── frontend/
│       └── app/             # React + TypeScript frontend (Vite)
├── infrastructure/          # Terraform for AWS/GCP, Kubernetes manifests
├── docker-compose.yaml      # Local development services
├── .env.example             # Template for local configuration (copy to `.env`)
├── Makefile                 # Convenience commands
└── docs/                    # Documentation
```

## Docker Compose Services

`docker compose up -d` brings up the stack described in
[`docker.md`](../08-tech-stack/infrastructure/docker.md#3-compose-topology). Compose
service names are `zookeeper`, `kafka`, `kafka-init`, `cassandra`, `api-gateway`,
`auth-service`, `account-service`, `transfer-service`, `transaction-service`,
`account-consumer`, `transfer-consumer`, `transaction-consumer`, and `frontend`.

![Local Docker Compose topology](../11-deployment/svg/01-local-topology.svg)


### Core Services

- **cassandra**: Apache Cassandra 4.1 (primary database, `cassandra:4.1`)
  - Port: 9042 (CQL native transport)
  - Configuration: single-node cluster — `SimpleStrategy`, `replication_factor: 1`; repositories issue reads/writes at `QUORUM` (on a single-node RF=1 cluster this resolves to the only replica, so multi-node quorum behaviour is not exercised locally)
  - Persistence: named volume `cassandra_data`

- **zookeeper**: Apache ZooKeeper (Kafka coordination), image `confluentinc/cp-zookeeper:7.4.0`
  - Port: 2181 (client connections, internal only)

- **kafka**: Apache Kafka (event streaming), image `confluentinc/cp-kafka:7.4.0` (identical tag to CI and [`docker.md`](../08-tech-stack/infrastructure/docker.md#3-compose-topology))
  - Ports: `9092` (`INTERNAL`, containers) and `29092` (`EXTERNAL`, host)
  - Dependencies: zookeeper
  - Configuration: **ZooKeeper mode** (single broker) — `KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181`, no `KAFKA_NODE_ID`/`KAFKA_PROCESS_ROLES`/`KAFKA_CLUSTER_ID`; `KAFKA_DEFAULT_REPLICATION_FACTOR=1`, `KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1`, `KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1`, `KAFKA_MIN_INSYNC_REPLICAS=1`; `KAFKA_AUTO_CREATE_TOPICS_ENABLE=false`
  - Listeners: `INTERNAL` advertises `kafka:9092` (container clients use `KAFKA_BROKERS=kafka:9092`); `EXTERNAL` advertises `localhost:29092` (host-run clients use `localhost:29092`)
  - Note: local runs at RF=1 so `QUORUM` resolves to the single replica; production uses RF=3 and true multi-node `QUORUM` (see [`NFR-012`](../01-requirements/non-functional-requirements.md#availability))

- **kafka-init**: one-shot container that creates topics after the broker is healthy (auto-create is disabled)
  - Topics: `transfers`, `transfers.dlq`

### Application Services

- **api-gateway**: Go API gateway (routing, auth enforcement, rate limiting)
  - Port: 8080 (HTTP API under `/api/v1`, health endpoints at root)
  - Build context: `./works/api-gateway`
  - Dependencies: auth-service, account-service, transfer-service, transaction-service
  - Environment: `AUTH_SERVICE_URL`, `ACCOUNT_SERVICE_URL`, `TRANSFER_SERVICE_URL`, `TRANSACTION_SERVICE_URL`, `JWT_PUBLIC_KEY_PEM`, `JWT_KEY_ID`, `CORS_ORIGINS`, `RATE_LIMIT_AUTH_PER_MIN`, `RATE_LIMIT_USER_PER_MIN`, `LOGIN_LOCKOUT_FAILURES`, `LOGIN_LOCKOUT_WINDOW_MINUTES`, `LOG_LEVEL`

- **auth-service**: Go Auth Service (owns `auth_keyspace`)
  - Port: 8081
  - Build context: `./works/auth`
  - Dependencies: cassandra, kafka
  - Environment: `CASSANDRA_HOST`, `CASSANDRA_KEYSPACE=auth_keyspace`, `KAFKA_BROKERS`, `JWT_PRIVATE_KEY_PEM`, `JWT_PUBLIC_KEY_PEM`, `JWT_KEY_ID`, `LOG_LEVEL`

- **account-service**: Go Account Service (owns `account_keyspace`); the compose container runs the HTTP API only. The saga consumer runs as a separate `account-consumer` container from the same image (entry point `/consumer`)
  - Port: 8082
  - Build context: `./works/accounts`
  - Dependencies: cassandra, kafka
  - Environment: `CASSANDRA_HOST`, `CASSANDRA_KEYSPACE=account_keyspace`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TRANSFERS`, `KAFKA_TOPIC_TRANSFERS_DLQ`, `KAFKA_CONSUMER_GROUP=account-service-saga`, `JWT_PUBLIC_KEY_PEM`, `JWT_KEY_ID`, `LOG_LEVEL`

- **transfer-service**: Go Transfer Service — saga orchestrator (owns `transfer_keyspace`); the compose container runs the HTTP API only. The saga consumer runs as a separate `transfer-consumer` container from the same image (entry point `/consumer`)
  - Port: 8083
  - Build context: `./works/transfer`
  - Dependencies: cassandra, kafka
  - Environment: `CASSANDRA_HOST`, `CASSANDRA_KEYSPACE=transfer_keyspace`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TRANSFERS`, `KAFKA_TOPIC_TRANSFERS_DLQ`, `KAFKA_CONSUMER_GROUP=transfer-service-saga`, `JWT_PUBLIC_KEY_PEM`, `JWT_KEY_ID`, `LOG_LEVEL`

- **transaction-service**: Go Transaction Service — read model (owns `transaction_keyspace`); the compose container runs the HTTP API only. The projection consumer runs as a separate `transaction-consumer` container from the same image (entry point `/consumer`)
  - Port: 8084
  - Build context: `./works/transactions`
  - Dependencies: cassandra, kafka
  - Environment: `CASSANDRA_HOST`, `CASSANDRA_KEYSPACE=transaction_keyspace`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TRANSFERS`, `KAFKA_TOPIC_TRANSFERS_DLQ`, `KAFKA_CONSUMER_GROUP=transaction-service-projection`, `JWT_PUBLIC_KEY_PEM`, `JWT_KEY_ID`, `LOG_LEVEL`

- **account-consumer**, **transfer-consumer**, **transaction-consumer**: the Kafka consumers for the three services above, each built from its service's context and run with `entrypoint: ["/consumer"]` (same image, separate process)
  - Dependencies: cassandra, kafka
  - Environment: `CASSANDRA_HOST`, `CASSANDRA_KEYSPACE`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TRANSFERS`, `KAFKA_TOPIC_TRANSFERS_DLQ`, `KAFKA_CONSUMER_GROUP`, `LOG_LEVEL`

- **frontend**: React frontend served by nginx from the production build
  - Port: 3000 (host) → 80 (nginx)
  - Build context: `./packages/app`
  - Dependencies: api-gateway
  - Environment: the API base URL is compiled into the bundle at build time from `VITE_API_BASE` (empty by default = same-origin; nginx proxies `/api` to the gateway)

## Environment Variables (.env.example)

Local configuration is read from a gitignored `.env` file at the repository root,
interpolated by Docker Compose. Copy the template:

```bash
cp .env.example .env
```

### Authentication & Security

Multi-line PEM values are double-quoted with literal `\n` escapes, which Compose
expands into newlines. Generate a development RSA key pair (PKCS#8 private / PKIX
public — both are accepted) and escape the newlines:

```bash
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out private.pem
openssl rsa -in private.pem -pubout -out public.pem
awk 'BEGIN{ORS="\\n"}{print}' private.pem   # paste into JWT_PRIVATE_KEY_PEM
awk 'BEGIN{ORS="\\n"}{print}' public.pem    # paste into JWT_PUBLIC_KEY_PEM
```

```env
JWT_KEY_ID=dev-key-1
JWT_PRIVATE_KEY_PEM="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
JWT_PUBLIC_KEY_PEM="-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----\n"
APP_ENV=development
```

### Database Configuration

```env
# Cassandra contact points (comma-separated). Inside Compose, use the service name.
CASSANDRA_HOST=cassandra
# auth-service -> auth_keyspace, account-service -> account_keyspace,
# transfer-service -> transfer_keyspace, transaction-service -> transaction_keyspace
CASSANDRA_KEYSPACE=<service>_keyspace
```

### Event Streaming

```env
# Container clients use the internal listener; host-run clients use localhost:29092.
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_TRANSFERS=transfers
KAFKA_TOPIC_TRANSFERS_DLQ=transfers.dlq
KAFKA_CONSUMER_GROUP=<service>-saga
```

### API Configuration

```env
PORT=8080
LOG_LEVEL=info
CORS_ORIGINS=http://localhost:3000
```

### Frontend Configuration

```env
# Baked into the bundle at build time. Empty (default) = same-origin, which
# the vite dev server and the nginx image both proxy to the gateway.
VITE_API_BASE=
```

## Development Workflow

### Starting Services

```bash
# Start all services in detached mode
docker compose up -d

# Verify services are running
docker compose ps
```

### Database Initialization

Each service ships a `/migrate` binary that applies its own versioned CQL
migrations to its keyspace (including `CREATE KEYSPACE`). Docker Compose runs
them automatically as one-shot `*-migrate` services, and every API and consumer
waits for its migration to finish before starting — so a single command
bootstraps a fresh Cassandra:

```bash
docker compose up -d
```

To run a migration by hand (it is idempotent), invoke the binary through the
service image; it reads `CASSANDRA_HOST` from the environment:

```bash
docker compose run --rm --entrypoint /migrate auth-service
docker compose run --rm --entrypoint /migrate account-service
docker compose run --rm --entrypoint /migrate transfer-service
docker compose run --rm --entrypoint /migrate transaction-service
```

The API must not accept traffic before migrations have run (see
[`cassandra.md`](../08-tech-stack/backend/cassandra.md)); the Compose
dependency conditions enforce this.

### Backend Development (Hot Reload)

Run services on the host for faster iteration. Point Cassandra at `localhost` and
Kafka at the external listener (`localhost:29092`):

```bash
cd works/api-gateway && go run cmd/api/main.go
cd works/auth && go run cmd/api/main.go
cd works/accounts && go run cmd/api/main.go
cd works/transfer && go run cmd/api/main.go
cd works/transactions && go run cmd/api/main.go
```

### Frontend Development (Hot Module Replacement)

```bash
cd packages/app
pnpm dev
```

### Kafka Consumers

The saga and projection consumers are separate processes that share the `transfers`
topic, each in its own consumer group. `docker compose up` starts them as the
`account-consumer`, `transfer-consumer`, and `transaction-consumer` services (the
same image as each service, with the entry point overridden to `/consumer`).

To run a consumer against the host toolchain instead, use the external listener at
`localhost:29092`:

```bash
cd works/transfer && go run cmd/consumer/main.go
cd works/accounts && go run cmd/consumer/main.go
cd works/transactions && go run cmd/consumer/main.go
```

To run one as a detached one-off container from the built image:

```bash
docker compose run -d --rm --entrypoint /consumer transfer-service
docker compose run -d --rm --entrypoint /consumer account-service
docker compose run -d --rm --entrypoint /consumer transaction-service
```

### Stopping and Cleaning Up

```bash
# Stop all services (preserves volumes)
docker compose down

# Stop services and remove volumes (fresh start)
docker compose down -v

# Remove orphaned containers and networks
docker compose down --remove-orphans
```

## Makefile Commands

The root Makefile provides repository-wide checks:

```bash
make build      # go build ./... for every backend service, then pnpm build
make test       # go test ./... for every backend service, then pnpm test
make typecheck  # go vet ./... for every backend service, then pnpm typecheck
make lint       # gofmt + go vet for every backend service, eslint + prettier --check for the frontend
make clean      # remove frontend build output and caches
make compose-up    # docker compose up -d --build (full local stack)
make compose-down  # docker compose down
make compose-logs  # tail logs from the local stack
make compose-ps    # show running stack containers
```

## Troubleshooting Common Issues

### Port Conflicts

If ports 9042, 29092, 8080, or 3000 are already in use:
- Identify the conflicting process: `sudo lsof -i :9042`
- Or modify the published ports in `docker-compose.yaml`

### Cassandra Connection Issues

- Cassandra can take up to a minute to become ready; Compose gates dependent services on its healthcheck
- Check logs: `docker compose logs cassandra`
- Verify with: `docker compose exec cassandra cqlsh`

### Kafka Connection Issues

- Ensure ZooKeeper starts before Kafka (Compose gates this on health)
- Check logs: `docker compose logs kafka`
- Verify with: `docker compose exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092`
- Confirm topics exist: `docker compose run --rm kafka-init`

### Docker Desktop Resource Limits

- Increase memory allocation in Docker Desktop settings (minimum 4GB recommended)
- Adjust CPUs if experiencing slow build times

## Performance Considerations for Local Development

### Resource Usage

- Approximate memory usage: 2-3GB when all services running
- CPU usage: minimal when idle, spikes during builds/tests
- Disk space: ~5GB for images, containers, and volumes

### Optimization Tips

- Use `:z` or `:Z` suffix on volume mounts for SELinux systems
- Use `docker compose` (v2) rather than the legacy `docker-compose` binary
- Prune unused images/containers periodically: `docker system prune -f`

## Security Notes for Development

### JWT Keys

- The `.env.example` values are placeholders for development only
- Never commit actual private keys to version control
- In production, use a proper secret management system

### Database Access

- The development Cassandra instance has no authentication enabled
- Never expose development ports to public networks
- Use Docker network isolation for container communication

## Technology Versions

| Component | Version | Purpose |
|-----------|---------|---------|
| Docker Compose | v2.0+ | Orchestration |
| Cassandra | 4.1 | Primary database |
| Kafka | `confluentinc/cp-kafka:7.4.0` | Event streaming |
| Go | 1.22+ | Backend language |
| Node.js | pinned in `package.json` `engines` | Frontend runtime |
| React | 19.x | Frontend library |
| Vite | 8.x | Frontend build tool |
