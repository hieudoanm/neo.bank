CREATE KEYSPACE IF NOT EXISTS account_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE account_keyspace;

CREATE TABLE IF NOT EXISTS accounts (
  account_id UUID,
  user_id UUID,
  balance_cents BIGINT,
  currency_code TEXT,
  status TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (account_id)
);

CREATE TABLE IF NOT EXISTS accounts_by_user (
  user_id UUID,
  account_id UUID,
  PRIMARY KEY (user_id, account_id)
);

CREATE TABLE IF NOT EXISTS processed_commands (
  command_id UUID,
  processed_at TIMESTAMP,
  PRIMARY KEY (command_id)
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
