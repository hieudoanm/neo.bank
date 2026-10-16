CREATE KEYSPACE IF NOT EXISTS transaction_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE transaction_keyspace;

CREATE TABLE IF NOT EXISTS transactions (
  account_id UUID,
  transaction_id UUID,
  created_at TIMESTAMP,
  transfer_id UUID,
  direction TEXT,
  amount_cents BIGINT,
  balance_after_cents BIGINT,
  description TEXT,
  PRIMARY KEY ((account_id), created_at, transaction_id)
);

CREATE TABLE IF NOT EXISTS audit_events (
  service TEXT,
  day DATE,
  timestamp TIMESTAMP,
  event_id UUID,
  event_type TEXT,
  actor_id UUID,
  resource_id UUID,
  ip_address TEXT,
  user_agent TEXT,
  details TEXT,
  PRIMARY KEY ((service, day), timestamp, event_id)
);
