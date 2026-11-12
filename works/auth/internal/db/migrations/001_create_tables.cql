CREATE KEYSPACE IF NOT EXISTS auth_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE auth_keyspace;

CREATE TABLE IF NOT EXISTS users (
  user_id UUID,
  email TEXT,
  account_id UUID,
  password_hash TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS users_by_email (
  email TEXT,
  user_id UUID,
  created_at TIMESTAMP,
  PRIMARY KEY (email)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  token_hash TEXT,
  user_id UUID,
  account_id UUID,
  email TEXT,
  expires_at TIMESTAMP,
  revoked_at TIMESTAMP,
  created_at TIMESTAMP,
  PRIMARY KEY (token_hash)
);

CREATE TABLE IF NOT EXISTS refresh_tokens_by_user (
  user_id UUID,
  token_hash TEXT,
  expires_at TIMESTAMP,
  revoked_at TIMESTAMP,
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, token_hash)
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
