CREATE KEYSPACE IF NOT EXISTS transfer_keyspace WITH replication = {'class':'SimpleStrategy','replication_factor':1};
USE transfer_keyspace;
CREATE TABLE IF NOT EXISTS transfers (transfer_id UUID, idempotency_key UUID, source_account_id UUID, destination_account_id UUID, amount_cents BIGINT, currency_code TEXT, status TEXT, failure_reason TEXT, created_at TIMESTAMP, updated_at TIMESTAMP, PRIMARY KEY (transfer_id));
CREATE TABLE IF NOT EXISTS transfers_by_account (account_id UUID, transfer_id UUID, status TEXT, created_at TIMESTAMP, PRIMARY KEY (account_id, transfer_id));
CREATE TABLE IF NOT EXISTS idempotency_keys (idempotency_key UUID, transfer_id UUID, request_hash TEXT, created_at TIMESTAMP, PRIMARY KEY (idempotency_key));
CREATE TABLE IF NOT EXISTS audit_events (service TEXT, day DATE, timestamp TIMESTAMP, event_id UUID, event_type TEXT, actor_id UUID, resource_id UUID, ip_address TEXT, user_agent TEXT, details TEXT, PRIMARY KEY ((service,day),timestamp,event_id));
