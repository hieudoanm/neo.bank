# NEO Bank — Docker Best Practices

_Infrastructure · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [infrastructure index](./README.md)._

Docker provides the reproducible local topology and the image that runs in every environment — local Compose, CI, and the Kubernetes cluster ([`kubernetes.md`](./kubernetes.md)). The full service list and Compose file are specified in [`11-deployment.md`](../../11-deployment.md).

## 1. Images

Build small, deterministic, non-root images with a multi-stage build.

### Backend (Go)

Every service is an independent Go module with its own binary, so each owns a small Dockerfile. The example below is `auth-service`; `api-gateway`, `account-service`, `transfer-service`, and `transaction-service` follow the same shape.

```dockerfile
# packages/backend/auth/Dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/auth-service ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12 AS runtime
COPY --from=build /out/auth-service /auth-service
COPY --from=build /out/migrate /migrate
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/auth-service"]
```

Rules:

- **One image per service.** `api-gateway`, `auth-service`, `account-service`, `transfer-service`, and `transaction-service` each build their own binary and image from their own module.
- Multi-stage: build in the SDK image, ship only the artifact in a minimal runtime image.
- Pin base image tags (`golang:1.22-alpine`, `cassandra:4.1`, `kafka:3.x`); never `latest`.
- Run as a non-root user (`USER nonroot`); the process needs no root.
- One process per image; each service's CQL migrations ship as a separate entrypoint that it owns (see [`cassandra.md`](../backend/cassandra.md)).
- Prefer distroless/static for Go; the frontend is built to static assets and served (see [`vite.md`](../frontend/core/vite.md)).

### Frontend

```dockerfile
# packages/frontend/app/Dockerfile
FROM node:22-alpine AS build
WORKDIR /src
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM nginx:alpine AS runtime
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80
ENTRYPOINT ["nginx", "-g", "daemon off;"]
```

## 2. Build Context

Keep the context small and explicit.

```dockerignore
# packages/backend/auth/.dockerignore
.git
**/*_test.go
coverage.out
bin/
*.env
.env*
```

Rules:

- Copy `go.mod`/`go.sum` and download dependencies **before** copying source, so the dependency layer caches.
- A `.dockerignore` is mandatory; never ship `.git`, tests, local artifacts, or `.env` files into the image.
- Builds are reproducible from the lockfile; do not fetch unpinned dependencies at build time.
- Multi-stage: dependency installation in build stage, only compiled output in runtime stage.

## 3. Compose Topology

One command brings up the full stack for local development.

```yaml
# docker-compose.yaml (excerpt)
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    environment:
      - ZOOKEEPER_CLIENT_PORT=2181
      - ZOOKEEPER_TICK_TIME=2000
    healthcheck:
      test: ["CMD-SHELL", "cub zk-ready localhost:2181 30"]
      interval: 10s
      timeout: 5s
      retries: 12

  kafka:
    image: confluentinc/cp-kafka:7.4.0
    environment:
      # ZooKeeper mode (the stack is ZooKeeper-based, not KRaft):
      # set KAFKA_ZOOKEEPER_CONNECT and leave KAFKA_NODE_ID/KAFKA_CLUSTER_ID unset.
      - KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092
      # Single broker (local): replication factor 1 everywhere, otherwise
      # internal topics with RF=3 never come up on one node.
      - KAFKA_DEFAULT_REPLICATION_FACTOR=1
      - KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1
      - KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1
      - KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1
      - KAFKA_MIN_INSYNC_REPLICAS=1
      - KAFKA_AUTO_CREATE_TOPICS_ENABLE=false
    healthcheck:
      test: ["CMD-SHELL", "kafka-topics --bootstrap-server localhost:9092 --list || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 12
    depends_on:
      zookeeper:
        condition: service_healthy

  cassandra:
    image: cassandra:4.1@sha256:...
    healthcheck:
      test: ["CMD-SHELL", "cqlsh -e 'describe keyspaces' || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 12
    volumes:
      - cassandra_data:/var/lib/cassandra

  api-gateway:
    build:
      context: ./packages/backend/api-gateway
      dockerfile: Dockerfile
    depends_on:
      - auth-service
      - account-service
      - transfer-service
      - transaction-service
    ports:
      - "8080:8080"

  auth-service:
    build:
      context: ./packages/backend/auth
      dockerfile: Dockerfile
    depends_on:
      cassandra:
        condition: service_healthy
      kafka:
        condition: service_healthy
    environment:
      - CASSANDRA_CONTACT_POINTS=cassandra:9042
      - CASSANDRA_KEYSPACE=auth_keyspace
      - KAFKA_BROKERS=kafka:9092
      - JWT_PRIVATE_KEY=${JWT_PRIVATE_KEY}

  frontend:
    build:
      context: ./packages/frontend/app
      dockerfile: Dockerfile
    depends_on:
      - api-gateway
    ports:
      - "3000:80"

volumes:
  cassandra_data:
```

Rules:

- Services: `zookeeper`, `kafka`, `cassandra`, `api-gateway`, `auth-service`, `account-service`, `transfer-service`, `transaction-service`, `frontend`. `account-service`, `transfer-service`, and `transaction-service` follow the `auth-service` shape (build context `./packages/backend/<service>`, health-gated `depends_on`, their own keyspace variable).
- Use `depends_on` with `condition: service_healthy` so each service waits for Cassandra, not just the container start.
- Every stateful service declares a `healthcheck`; startup order follows health, not hope.
- Named volumes for state (`cassandra_data`); `docker-compose down -v` is the explicit reset.
- Configuration comes from environment variables sourced from `.env`; no secrets baked into the Compose file.
- **Kafka runs in ZooKeeper mode** here: set `KAFKA_ZOOKEEPER_CONNECT` and do **not** also set `KAFKA_NODE_ID`/`KAFKA_PROCESS_ROLES`/`KAFKA_CLUSTER_ID` (KRaft and ZooKeeper settings must not be combined — a mixed config fails to start).
- **Local/CI single broker ⇒ RF=1** for every internal topic (`default`, `offsets`, transaction state log) and `min.insync.replicas=1`. Production keeps RF=3 / `min.insync.replicas=2` (see [`NFR-012`, `NFR-013`](../../01-requirements/non-functional-requirements.md#availability)).
- With `KAFKA_AUTO_CREATE_TOPICS_ENABLE=false`, the `transfers` and `transfers.dlq` topics are created by an explicit init step (`kafka-topics --create`), never by the first producer write.
- Pin image tags (`confluentinc/cp-kafka:7.4.0`, `confluentinc/cp-zookeeper:7.4.0`, `cassandra:4.1`) and pin digests (`@sha256:...`) for reproducibility; local Compose, CI service containers, and the docs must use the same tags.

## 4. Healthchecks & Startup

- Liveness: the process answers `/healthz`. Readiness: dependencies (Cassandra, Kafka) are reachable.
- Cassandra can take a minute to become ready; retry with a bounded wait rather than a fixed sleep.
- The API must not accept traffic before migrations have run (see [`cassandra.md`](../backend/cassandra.md)).
- Healthcheck failures trigger container restarts (Docker restart policy `unless-stopped`).

## 5. Security

- No secrets in images, Compose, or build args — pass them as runtime environment variables from a secret store.
- Images are scanned (Trivy) in CI; a high-severity finding fails the build.
- Minimise the attack surface: distroless/static runtime, no shell, no package manager.
- Pin digests for production pulls; tags are a convenience for local development.
- Non-root user in every runtime stage.

## 6. Image Variants

| Stage | Image | Purpose |
|-------|-------|---------|
| Build | `golang:1.22-alpine` | Compile Go binaries |
| Build | `node:22-alpine` | Build frontend assets |
| Runtime | `gcr.io/distroless/static-debian12` | Run the Go gateway and services |
| Runtime | `nginx:alpine` | Serve static frontend |
| Data | `cassandra:4.1` | Database |
| Data | `confluentinc/cp-kafka:7.4.0` | Message broker |

## 7. Code Style & Rules

- Multi-stage builds; pinned tags; non-root runtime.
- `.dockerignore` in every image context; dependency layers cached first.
- One image and one process per service; each service builds its own binary.
- Compose declares healthchecks and health-gated `depends_on`.
- Configuration and secrets are runtime-only; never baked into an image.
- Images are scanned in CI; high-severity findings block the merge.
- Related: [`kubernetes.md`](./kubernetes.md) for how the images are orchestrated, [`terraform.md`](./terraform.md) for cloud deployment, [`github-actions.md`](./github-actions.md) for the CI pipeline, [`11-deployment.md`](../../11-deployment.md) for the full topology.
