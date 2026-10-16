CREATE KEYSPACE IF NOT EXISTS account_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE account_keyspace;

CREATE TABLE IF NOT EXISTS accounts (
  account_id UUID,
  user_id UUID,
  balance_cents BIGINT,
  currency_code TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (account_id)
);

CREATE TABLE IF NOT EXISTS accounts_by_user (
  user_id UUID,
  account_id UUID,
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, account_id)
);

CREATE TABLE IF NOT EXISTS processed_commands (
  idempotency_key TEXT,
  transfer_id UUID,
  account_id UUID,
  event_type TEXT,
  result_event_type TEXT,
  balance_cents BIGINT,
  failure_reason TEXT,
  created_at TIMESTAMP,
  applied_at TIMESTAMP,
  PRIMARY KEY (idempotency_key)
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
