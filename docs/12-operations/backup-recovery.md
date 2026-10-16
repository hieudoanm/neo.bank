# Backup and Recovery

## Purpose

Defines the backup and recovery strategy for NEO Bank's data stores: Cassandra and Kafka. Backups ensure data can be restored after corruption, accidental deletion, or regional failure.

## 1. Cassandra Backup

### Local (Docker Compose)

```bash
# Take a snapshot of every service keyspace (one nodetool call per keyspace)
docker-compose exec cassandra nodetool snapshot auth_keyspace
docker-compose exec cassandra nodetool snapshot account_keyspace
docker-compose exec cassandra nodetool snapshot transfer_keyspace
docker-compose exec cassandra nodetool snapshot transaction_keyspace

# Copy snapshot files
docker cp cassandra:/var/lib/cassandra/data/auth_keyspace*/snapshots ./backup
# ...repeat for account_keyspace, transfer_keyspace, transaction_keyspace

# Restore snapshot
docker cp ./backup cassandra:/var/lib/cassandra/data/auth_keyspace*/snapshots
docker-compose exec cassandra nodetool refresh auth_keyspace
# ...repeat per keyspace
```

### AWS Amazon Keyspaces

- Enable continuous point-in-time recovery (PITR) — supports the 1-hour RPO.
- Automatic daily snapshots (35-day retention).
- Manual snapshots before schema changes.
- Restore to a new keyspace/table from PITR or snapshot.

### GCP Cassandra on GKE

- Cassandra runs on GKE (`gcp` Terraform stack); take hourly snapshots via a scheduled `nodetool snapshot` job.
- Export snapshots to Cloud Storage for long-term retention and cross-region durability.
- Restore snapshots to a new Cassandra cluster in the DR region.

## 2. Kafka Backup

### Local

- Kafka topics are ephemeral; transfer state is in Cassandra.
- If Kafka fails, consumer offsets reset to `earliest` and the retained `transfers` events are replayed into the transaction projection.
- Kafka (not Cassandra) is the source of truth for event replay: **do not** try to re-derive Kafka events from Cassandra. Cassandra holds current state (`transfers`, `transactions`); the append-only `transfers` topic holds the events, retained for 7 days.

### AWS MSK

- Enable data encryption at rest.
- Use multi-AZ deployment for durability.
- Kafka topics replicated across AZs.

### GCP Kafka

- Kafka is the event backbone on GCP (managed Kafka or Kafka on GKE).
- Enable message retention (7 days default) for `transfers`.
- Route failed messages to the dead-letter topic `transfers.dlq` for inspection and replay.

## 3. Recovery Objectives

The backup schedule is designed to meet a single coherent RPO/RTO pair, defined in full in [`disaster-recovery.md`](./disaster-recovery.md):

| Objective | Target | How the schedule supports it |
|-----------|--------|------------------------------|
| Recovery Point Objective (RPO) | 1 hour | Continuous PITR on Amazon Keyspaces and hourly Cassandra snapshots on GKE bound data loss to ≤ 1 hour. |
| Recovery Time Objective (RTO) | 4 hours | Pre-provisioned DR region and versioned migrations allow restore + redeploy + DNS cutover within 4 hours. |

## 4. Recovery Procedure

1. **Identify failure** — Determine which data store is affected and the point of failure.
2. **Select backup source** — Choose the most recent valid snapshot or PITR point (≤ 1 hour old to meet the RPO).
3. **Restore data** — Apply snapshot or PITR to restore the data store.
4. **Verify integrity** — Run checksums and spot checks against known data; confirm no negative balances and no ledger imbalance.
5. **Resume operations** — Restart consumers and API; verify end-to-end flows.
6. **Document recovery** — Record what was recovered, when, and from which backup.

## 5. Backup Schedule

| Data store | Frequency | Retention | Method |
|------------|-----------|-----------|--------|
| Cassandra (local) | Hourly snapshots | 24 hours | `nodetool snapshot` |
| Cassandra (Amazon Keyspaces) | Continuous PITR + daily snapshots | 35 days | AWS managed |
| Cassandra (GKE) | Hourly snapshots | 7 days hot, then Cloud Storage | `nodetool snapshot` job + GCS export |
| Kafka topics (`transfers`, `transfers.dlq`) | Native retention | 7 days | Kafka retention |

## 6. Testing

- Restore tested quarterly (full restore from backup to staging).
- Backup integrity verified after each backup (checksums).
- Recovery procedure documented and updated after each test.

## Cross-references

- Disaster recovery: [`disaster-recovery.md`](./disaster-recovery.md)
- Rollback strategy: [`rollback-strategy.md`](./rollback-strategy.md)
- Cassandra backend: [`docs/08-tech-stack/backend/cassandra.md`](../08-tech-stack/backend/cassandra.md)
- Kafka backend: [`docs/08-tech-stack/backend/kafka.md`](../08-tech-stack/backend/kafka.md)
- Operations index: [`12-operations.md`](../12-operations.md)
- Planned test: `test_operations_backup` (verifies backup procedures are documented)