# NEO Bank — Zookeeper Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Purpose

Zookeeper is the coordination service for Apache Kafka. In NEO Bank it is used in two contexts:

- **Local development:** Docker Compose runs a Zookeeper node alongside Kafka (`zookeeper` service) so the stack is reproducible without cloud credentials.
- **Production:** Managed Kafka (AWS MSK) or Kafka on GKE (Strimzi) runs Zookeeper internally or uses KRaft mode; NEO Bank does not operate Zookeeper directly in production.

This document covers how Zookeeper is configured and operated in the local Docker Compose topology.

## 2. Role in the Stack

```
┌─────────┐     ┌─────────┐
│   API   │────▶│  Kafka  │─────┐
└─────────┘     └────┬────┘     │
                     │          ▼
              ┌──────▼──────┐  ┌─────────┐
              │ Zookeeper   │◀─┤ Consumer│
              │ (coordination)│  └─────────┘
              └─────────────┘
```

Zookeeper is responsible for:
- Kafka broker registration and membership
- Topic and partition metadata
- Consumer group offsets (in older Kafka versions)
- ACLs and configuration (when used)
- Leader election for Kafka partitions

NEO Bank does not use Zookeeper directly from application code — only the Kafka broker and client libraries interact with it.

## 3. Deployment Mode

| Environment | Mode | Notes |
|-------------|------|-------|
| Local (Docker Compose) | Standalone Zookeeper (1 node) | Single node is acceptable for development; no quorum needed |
| Production | Managed by MSK / KRaft | NEO Bank does not run Zookeeper nodes directly |

**Rules:**
- Local Zookeeper runs as a single container (`zookeeper` service in `docker-compose.yaml`).
- Do not add application-level dependencies on Zookeeper.
- If KRaft mode is adopted in production, Zookeeper is removed entirely from the topology.

## 4. Configuration (Local Docker Compose)

```yaml
# docker-compose.yaml (excerpt)
zookeeper:
  image: confluentinc/cp-zookeeper:7.4.0
  environment:
    ZOOKEEPER_CLIENT_PORT: 2181
    ZOOKEEPER_TICK_TIME: 2000
    ZOOKEEPER_INIT_LIMIT: 10
    ZOOKEEPER_SYNC_LIMIT: 5
    ZOOKEEPER_MAX_CLIENT_CNNXNS: 1000
    ZOOKEEPER_SNAP_RETAIN_COUNT: 3
    ZOOKEEPER_PURGE_INTERVAL: 1
  volumes:
    - zookeeper-data:/var/lib/zookeeper/data
    - zookeeper-log:/var/lib/zookeeper/log
  ports:
    - "2181:2181"
```

**Configuration rules:**

| Setting | Value | Reason |
|---------|-------|--------|
| `ZOOKEEPER_CLIENT_PORT` | `2181` | Standard client port |
| `ZOOKEEPER_TICK_TIME` | `2000` | 2s heartbeat; baseline for timeouts |
| `ZOOKEEPER_INIT_LIMIT` | `10` | 20s initial sync limit |
| `ZOOKEEPER_SYNC_LIMIT` | `5` | 10s follower sync limit |
| `ZOOKEEPER_MAX_CLIENT_CNNXNS` | `1000` | Allow enough concurrent Kafka connections |
| `ZOOKEEPER_SNAP_RETAIN_COUNT` | `3` | Keep 3 recent snapshots |
| `ZOOKEEPER_PURGE_INTERVAL` | `1` | Purge old snapshots hourly |

## 5. Quorum Rules

Zookeeper requires a majority (quorum) to be available:

| Node count | Fault tolerance | Quorum needed |
|------------|----------------|---------------|
| 1 | 0 | 1 (local dev only) |
| 3 | 1 | 2 |
| 5 | 2 | 3 |

**Rules:**
- Local development uses a single Zookeeper node — acceptable because there is no production data at risk.
- Production never uses a single Zookeeper node (managed services handle quorum internally).
- Never scale Zookeeper to an even number of nodes (quorum math gets worse, not better).

## 6. Data & Snapshots

```text
/var/lib/zookeeper/data/   # in-memory snapshot (myid, version-2)
/var/lib/zookeeper/log/    # transaction log
```

**Rules:**
- Data and log directories are separate volumes for performance.
- Snapshots are retained (`snapRetainCount=3`); old snapshots are purged automatically (`autopurge.purgeInterval=1`).
- Do not edit snapshot files manually; they are Zookeeper's internal state.
- Back up Zookeeper data only when recreating the Kafka cluster; Kafka topic metadata is reconstructible from topics themselves.

## 7. Health Checks

```bash
# Liveness
echo ruok | nc localhost 2181
# Expected: imok

# Monitoring (4lw commands must be enabled)
echo mntr | nc localhost 2181
# Returns: zookeeper_version, zookeeper_state, zookeeper_num_alive_connections, etc.
```

**Rules:**
- `ruok` → `imok` means the Zookeeper process is alive.
- `mntr` returns detailed statistics (latency, connections, outstanding requests).
- 4-letter word commands (`ruok`, `mntr`, `stat`, `cons`) are enabled in local dev but should be restricted or disabled in production.
- Docker healthcheck uses `ruok` with a short timeout.

## 8. Security

| Context | Auth | TLS | ACLs |
|---------|------|-----|------|
| Local dev | None | No | No |
| Production (managed) | Managed by provider | Managed by provider | Managed by provider |

**Rules:**
- Local Zookeeper listens only on the Docker network, not exposed to the host beyond the Compose port mapping.
- No application credentials are stored in Zookeeper.
- In production, Zookeeper access is controlled by the managed service's VPC and security groups, not by application configuration.

## 9. Observability

Export these metrics from Zookeeper (via `mntr` or JMX):

- `zookeeper_num_alive_connections` — active client connections
- `zookeeper_outstanding_requests` — requests waiting to be processed
- `zookeeper_avg_latency` — average request latency (ms)
- `zookeeper_max_latency` — max request latency (ms)
- `zookeeper_packets_received` / `zookeeper_packets_sent` — traffic
- `zookeeper_state` — `standalone`, `leader`, or `follower`

**Rules:**
- Alert when `avg_latency` > 100ms for 5 minutes.
- Alert when `outstanding_requests` > 100 for 5 minutes.
- Zookeeper metrics are secondary; Kafka consumer lag (`kafka_consumer_lag`) is the primary health signal.

## 10. Operational Procedures

### Starting Zookeeper
```bash
docker-compose up -d zookeeper
docker-compose logs zookeeper   # wait for "binding to port 0.0.0.0/0.0.0.0:2181"
```

### Stopping Zookeeper
```bash
docker-compose stop zookeeper
```
- Kafka will lose its coordination service; stop Kafka first if you need to preserve offsets.

### Data Persistence
- Zookeeper data is persisted in named Docker volumes (`zookeeper-data`, `zookeeper-log`).
- `docker-compose down` does **not** remove volumes; data survives restarts.
- `docker-compose down -v` removes volumes and clears all Kafka cluster metadata.

### Upgrading Zookeeper
- Match the Zookeeper image version to the Kafka image version (e.g., Confluent Platform 7.5.0 for both).
- Rolling restart: stop Zookeeper, update image, start Zookeeper, wait for `imok`, then restart Kafka.

## 11. Code Style & Rules

- **No direct Zookeeper client** in application code; all interaction goes through the Kafka client library.
- **Zookeeper is infrastructure, not application state** — never store business data in Zookeeper nodes.
- **Local-only topology** — Zookeeper appears only in `docker-compose.yaml`; production uses managed services.
- **Odd node counts only** — if Zookeeper is ever run manually, use 3 or 5 nodes, never 2 or 4.
- **Health via `ruok`** — the container healthcheck uses the 4-letter word command.

---
↑ Part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md).