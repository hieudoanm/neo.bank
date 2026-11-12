CREATE KEYSPACE IF NOT EXISTS transfer_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE transfer_keyspace;

CREATE TABLE IF NOT EXISTS transfers (
  transfer_id UUID PRIMARY KEY,
  source_account_id UUID,
  destination_account_id UUID,
  amount_cents BIGINT,
  status TEXT,
  idempotency_key TEXT,
  correlation_id TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  failure_reason TEXT
);

CREATE TABLE IF NOT EXISTS transfers_by_account (
  account_id UUID,
  transfer_id UUID,
  direction TEXT,
  counterparty_account_id UUID,
  amount_cents BIGINT,
  status TEXT,
  correlation_id TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  failure_reason TEXT,
  PRIMARY KEY ((account_id), transfer_id)
) WITH CLUSTERING ORDER BY (transfer_id DESC);

CREATE TABLE IF NOT EXISTS idempotency_keys (
  idempotency_key TEXT PRIMARY KEY,
  operation_type TEXT,
  resource_id UUID,
  request_hash TEXT,
  response_status INT,
  response_body TEXT,
  expires_at TIMESTAMP,
  created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_events (
  service TEXT,
  day DATE,
  timestamp TIMESTAMP,
  event_id UUID,
  actor_user_id UUID,
  action TEXT,
  resource_type TEXT,
  resource_id UUID,
  correlation_id TEXT,
  metadata MAP<TEXT, TEXT>,
  PRIMARY KEY ((service, day), timestamp, event_id)
) WITH CLUSTERING ORDER BY (timestamp DESC, event_id DESC);
