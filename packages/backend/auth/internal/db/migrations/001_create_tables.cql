CREATE KEYSPACE IF NOT EXISTS auth_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE auth_keyspace;

CREATE TABLE IF NOT EXISTS users (
  user_id UUID,
  email TEXT,
  password_hash TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS users_by_email (
  email TEXT,
  user_id UUID,
  PRIMARY KEY (email)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  token_hash TEXT,
  user_id UUID,
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
  event_type TEXT,
  actor_id UUID,
  resource_id UUID,
  ip_address TEXT,
  user_agent TEXT,
  details TEXT,
  PRIMARY KEY ((service, day), timestamp, event_id)
);
