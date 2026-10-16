# NEO Bank — Docker Best Practices

_Infrastructure · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [infrastructure index](./README.md)._

Docker provides the reproducible local topology and the image that runs in every environment — local Compose, CI, and the Kubernetes cluster ([`kubernetes.md`](./kubernetes.md)). The full service list and Compose file are specified in [`11-deployment.md`](../../11-deployment.md).

## 1. Images

Build small, deterministic, non-root images with a multi-stage build.

### Backend (Go)

Every service is an independent Go module with its own binary, so each owns a small Dockerfile. The example below is `auth-service`; `api-gateway`, `account-service`, `transfer-service`, and `transaction-service` follow the same shape.

```dockerfile
# packages/auth-service/Dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/auth-service ./cmd/auth-service
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
# packages/frontend/Dockerfile
FROM node:20-alpine AS build
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
# packages/auth-service/.dockerignore
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
    image: bitnami/zookeeper:3.9@sha256:...
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes

  kafka:
    image: bitnami/kafka:3.x@sha256:...
    environment:
      - KAFKA_CFG_NODE_ID=0
      - KAFKA_CFG_PROCESS_ROLES=controller,broker
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
    healthcheck:
      test: ["CMD", "kafka-topics.sh", "--bootstrap-server", "localhost:9092", "--list"]
      interval: 10s
      timeout: 5s
      retries: 12
    depends_on:
      - zookeeper

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
      context: ./packages/api-gateway
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
      context: ./packages/auth-service
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
      context: ./packages/frontend
      dockerfile: Dockerfile
    depends_on:
      - api-gateway
    ports:
      - "3000:80"

volumes:
  cassandra_data:
```

Rules:

- Services: `zookeeper`, `kafka`, `cassandra`, `api-gateway`, `auth-service`, `account-service`, `transfer-service`, `transaction-service`, `frontend`. `account-service`, `transfer-service`, and `transaction-service` follow the `auth-service` shape (build context `./packages/<service>`, health-gated `depends_on`, their own keyspace variable).
- Use `depends_on` with `condition: service_healthy` so each service waits for Cassandra, not just the container start.
- Every stateful service declares a `healthcheck`; startup order follows health, not hope.
- Named volumes for state (`cassandra_data`); `docker-compose down -v` is the explicit reset.
- Configuration comes from environment variables sourced from `.env`; no secrets baked into the Compose file.
- Pin image digests (`@sha256:...`) for reproducibility; tags are a convenience for local development.

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
| Build | `node:20-alpine` | Build frontend assets |
| Runtime | `gcr.io/distroless/static-debian12` | Run the Go gateway and services |
| Runtime | `nginx:alpine` | Serve static frontend |
| Data | `cassandra:4.1` | Database |
| Data | `bitnami/kafka:3.x` | Message broker |

## 7. Code Style & Rules

- Multi-stage builds; pinned tags; non-root runtime.
- `.dockerignore` in every image context; dependency layers cached first.
- One image and one process per service; each service builds its own binary.
- Compose declares healthchecks and health-gated `depends_on`.
- Configuration and secrets are runtime-only; never baked into an image.
- Images are scanned in CI; high-severity findings block the merge.
- Related: [`kubernetes.md`](./kubernetes.md) for how the images are orchestrated, [`terraform.md`](./terraform.md) for cloud deployment, [`github-actions.md`](./github-actions.md) for the CI pipeline, [`11-deployment.md`](../../11-deployment.md) for the full topology.
