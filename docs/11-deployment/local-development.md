# Local Development

## Prerequisites

- Docker and Docker Compose (v2.0+)
- Go 1.22+ (for backend development)
- Node.js 18+ LTS (for frontend development)
- pnpm (v8+)
- GNU Make (optional, for convenience commands)
- Git (v2.30+)

## Repository Structure

```text
neo.bank/
├── packages/
│   ├── frontend/            # React + TypeScript frontend (Vite)
│   ├── api-gateway/         # Go API gateway (routing, auth enforcement, rate limiting)
│   ├── auth-service/        # Go Auth Service
│   ├── account-service/     # Go Account Service
│   ├── transfer-service/    # Go Transfer Service — saga orchestrator
│   └── transaction-service/ # Go Transaction Service — read model
├── infrastructure/          # Terraform for AWS/GCP, Kubernetes manifests
├── docker-compose.yaml      # Local development services (PLANNED)
├── Makefile                 # Convenience commands (PLANNED)
└── docs/                    # Documentation
```

## Docker Compose Services

The local topology is the microservices stack. All of the following are **PLANNED** as this repository has no application code yet.

### Core Services
- **cassandra**: Apache Cassandra 4.1 (primary database)
  - Port: 9042 (CQL native transport)
  - Configuration: Single-node cluster for development
  - Persistence: Named volume `cassandra_data`

- **zookeeper**: Apache Zookeeper 3.8 (Kafka coordination)
  - Port: 2181 (client connections)
  - Image: confluentinc/cp-zookeeper:7.4.0

- **kafka**: Apache Kafka 3.5 (event streaming)
  - Port: 9092 (plaintext broker)
  - Dependencies: zookeeper
  - Configuration: Single broker, `auto.create.topics.enable=false`

- **api-gateway**: Go API gateway (routing, auth enforcement, rate limiting)
  - Port: 8080 (HTTP API, `/api/v1`, health endpoints at root)
  - Build context: `./packages/api-gateway` (PLANNED)
  - Dockerfile: `Dockerfile` (PLANNED)
  - Dependencies: auth-service, account-service, transfer-service, transaction-service, kafka (as applicable)
  - Environment: service URLs, Kafka brokers, JWT keys, log level

- **auth-service**: Go Auth Service (owns `auth_keyspace`)
  - Port: 8081 (PLANNED)
  - Build context: `./packages/auth-service` (PLANNED)
  - Dockerfile: `Dockerfile` (PLANNED)
  - Dependencies: cassandra, kafka
  - Environment: Cassandra connection, Kafka brokers, JWT keys, log level

- **account-service**: Go Account Service (owns `account_keyspace`)
  - Port: 8082 (PLANNED)
  - Build context: `./packages/account-service` (PLANNED)
  - Dockerfile: `Dockerfile` (PLANNED)
  - Dependencies: cassandra, kafka
  - Environment: Cassandra connection, Kafka brokers, log level

- **transfer-service**: Go Transfer Service — saga orchestrator (owns `transfer_keyspace`)
  - Port: 8083 (PLANNED)
  - Build context: `./packages/transfer-service` (PLANNED)
  - Dockerfile: `Dockerfile` (PLANNED)
  - Dependencies: cassandra, kafka
  - Environment: Cassandra connection, Kafka brokers, Kafka topics (`transfers`, `transfers.dlq`), log level

- **transaction-service**: Go Transaction Service — read model (owns `transaction_keyspace`)
  - Port: 8084 (PLANNED)
  - Build context: `./packages/transaction-service` (PLANNED)
  - Dockerfile: `Dockerfile` (PLANNED)
  - Dependencies: cassandra, kafka
  - Environment: Cassandra connection, Kafka brokers, log level

- **frontend**: React frontend development server
  - Port: 3000 (HTTP)
  - Build context: `./packages/frontend` (PLANNED)
  - Dockerfile: `Dockerfile` (PLANNED)
  - Dependencies: api-gateway
  - Environment: `VITE_API_URL` pointing to API gateway

## Environment Variables (.env.example)

### Authentication & Security
```env
# JWT Keys (2048-bit RSA key pair)
JWT_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----
JWT_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...
-----END PUBLIC KEY-----
```

### Database Configuration
```env
# Cassandra connection (one keyspace per service; set per service)
CASSANDRA_CONTACT_POINTS=localhost:9042
# auth-service -> auth_keyspace, account-service -> account_keyspace,
# transfer-service -> transfer_keyspace, transaction-service -> transaction_keyspace
CASSANDRA_KEYSPACE=<service>_keyspace
```

### Event Streaming
```env
# Kafka configuration
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_TRANSFERS=transfers
KAFKA_TOPIC_TRANSFERS_DLQ=transfers.dlq
```

### API Configuration
```env
API_PORT=8080
LOG_LEVEL=debug
CORS_ORIGINS=http://localhost:3000
```

### Frontend Configuration
```env
VITE_API_URL=http://localhost:8080/api/v1
```

## Development Workflow

**Note:** All commands below are **PLANNED** (no code exists yet in this repository). The `docker-compose.yaml` file does not exist yet.

### Starting Services
```bash
# Start all services in detached mode (PLANNED)
docker-compose up -d

# Verify services are running (PLANNED)
docker-compose ps
```

### Database Initialization
```bash
# Apply database migrations to Cassandra (per service, PLANNED)
# e.g. for each service with its own keyspace:
# docker-compose exec auth-service go run cmd/migrate/main.go
# docker-compose exec account-service go run cmd/migrate/main.go
# docker-compose exec transfer-service go run cmd/migrate/main.go
# docker-compose exec transaction-service go run cmd/migrate/main.go

# Check migration status (PLANNED)
# docker-compose exec <service> go run cmd/migrate/main.go -status
```

### Backend Development (Hot Reload)
```bash
# Run services outside Docker for faster iteration (PLANNED)
# cd packages/api-gateway && go run cmd/api-gateway/main.go
# cd packages/auth-service && go run cmd/auth-service/main.go
# cd packages/account-service && go run cmd/account-service/main.go
# cd packages/transfer-service && go run cmd/transfer-service/main.go
# cd packages/transaction-service && go run cmd/transaction-service/main.go

# Alternative: Use air for live reloading (PLANNED)
# air
```

### Frontend Development (Hot Module Replacement)
```bash
# Run frontend development server (PLANNED)
cd packages/frontend
pnpm dev
```

### Kafka Consumers
```bash
# Run Kafka consumers outside Docker (PLANNED)
# cd packages/transfer-service && go run cmd/consumer/main.go
# cd packages/transaction-service && go run cmd/consumer/main.go
```

### Stopping and Cleaning Up
```bash
# Stop all services (preserves volumes) (PLANNED)
docker-compose down

# Stop services and remove volumes (fresh start) (PLANNED)
docker-compose down -v

# Remove orphaned containers and networks (PLANNED)
docker-compose down --remove-orphans
```

## Makefile Commands (Optional Convenience)

The Makefile is **PLANNED**. When implemented, it may provide shortcuts for common development tasks. Examples:

```bash
# Start all services (PLANNED)
make up

# Stop all services (PLANNED)
make down

# Apply database migrations (PLANNED)
make migrate

# Run API gateway (PLANNED)
make api

# Run frontend development server (PLANNED)
make frontend

# Run Kafka consumers (PLANNED)
make consumer

# Run all backend tests (PLANNED)
make test

# Run all frontend tests (PLANNED)
make test-frontend

# Run backend linter (PLANNED)
make lint

# Run frontend linter (PLANNED)
make lint-frontend

# Run all tests (backend + frontend) (PLANNED)
make test-all
```

## Troubleshooting Common Issues

### Port Conflicts
If ports 9042, 9092, 8080, or 3000 are already in use:
- Stop conflicting services: `sudo lsof -i :9042`
- Or modify docker-compose.yaml ports temporarily

### Cassandra Connection Issues
- Wait 10-15 seconds after Cassandra starts for it to be ready
- Check logs: `docker-compose logs cassandra`
- Verify with: `docker-compose exec cassandra cqlsh`

### Kafka Connection Issues
- Ensure Zookeeper starts before Kafka
- Check logs: `docker-compose logs kafka`
- Verify with: `docker-compose exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092`

### Docker Desktop Resource Limits
- Increase memory allocation in Docker Desktop settings (minimum 4GB recommended)
- Adjust CPUs if experiencing slow build times

## Performance Considerations for Local Development

### Resource Usage
- Approximate memory usage: 2-3GB when all services running
- CPU usage: Minimal when idle, spikes during builds/tests
- Disk space: ~5GB for images, containers, and volumes

### Optimization Tips
- Use `:z` or `:Z` suffix on volume mounts for SELinux systems
- Consider using `docker compose` (v2) instead of `docker-compose`
- Prune unused images/containers periodically: `docker system prune -f`

## Security Notes for Development

### JWT Keys
- The example keys in .env.example are for development only
- Never commit actual private keys to version control
- In production, use proper secret management systems

### Database Access
- Development Cassandra instance has no authentication enabled
- Never expose development ports to public networks
- Use Docker network isolation for container communication

## Technology Versions

| Component | Version | Purpose |
|-----------|---------|---------|
| Docker Compose | v2.0+ | Orchestration |
| Cassandra | 4.1.x | Primary database |
| Kafka | 3.5.x | Event streaming |
| Go | 1.22.x | Backend language |
| Node.js | 18.x LTS | Frontend runtime |
| React | 18.x | Frontend library |
| Vite | 4.x | Frontend build tool |

