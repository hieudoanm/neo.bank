CREATE KEYSPACE IF NOT EXISTS transaction_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE transaction_keyspace;

CREATE TABLE IF NOT EXISTS transactions (
  account_id UUID,
  created_at TIMESTAMP,
  transaction_id UUID,
  transfer_id UUID,
  type TEXT,
  amount_cents BIGINT,
  counterparty_account_id UUID,
  status TEXT,
  PRIMARY KEY ((account_id), created_at, transaction_id)
) WITH CLUSTERING ORDER BY (created_at DESC, transaction_id DESC);

CREATE TABLE IF NOT EXISTS transfer_movements (
  transfer_id UUID,
  movement TEXT,
  transaction_id UUID,
  account_id UUID,
  created_at TIMESTAMP,
  PRIMARY KEY ((transfer_id), movement)
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
