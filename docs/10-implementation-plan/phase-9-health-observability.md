# NEO Bank — Implementation Plan — Phase 9: Health and Observability

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement health checks, structured logging, and metrics collection. Phase 9 makes the system observable and operable in production.

**Week:** 6-7

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Health endpoints | `GET /healthz`, `GET /readyz` |
| Structured logging | JSON logs with correlation IDs |
| Metrics | Prometheus metrics (latency, error rate, throughput, Kafka lag) |
| Alerting | Alerting rules for error rate, latency, Kafka lag |
| Documentation | Observability guidelines updated |

## Tasks

- [ ] Implement `GET /healthz` endpoint (liveness, process health).
- [ ] Implement `GET /readyz` endpoint (readiness, Cassandra + Kafka connectivity).
- [ ] Implement structured logging (JSON, correlation IDs on every request).
- [ ] Add correlation IDs to all requests (middleware-generated or propagated).
- [ ] Implement metrics collection (Prometheus client, custom metrics).
- [ ] Configure alerting rules (error rate > 1%, p99 latency > 2s, Kafka lag).
- [ ] Update documentation with observability guidelines.
- [ ] Verify structured logs include correlation IDs end-to-end.

## Acceptance Criteria

- Health endpoints return correct status (healthy/unhealthy, ready/not ready).
- Logs are structured JSON with correlation IDs on every line.
- Metrics are collected and exportable via Prometheus.
- Alerting rules configured for error rate, latency, and Kafka lag.
- Health endpoints are documented for use by Kubernetes readiness/liveness probes; probe wiring itself is deferred to Phase 11 (see [`phase-11-cloud-infrastructure.md`](./phase-11-cloud-infrastructure.md)).
- Documentation updated with observability runbook.
- Tests pass (unit, integration for health endpoints).

## Dependencies

Phase 8 (complete system to observe and healthcheck).

## Testing

- Unit tests: health check logic, metrics emission, log formatting.
- Integration tests: health endpoints with real Cassandra/Kafka dependencies.
- Observability tests: correlation ID propagation through request lifecycle.
- Alert tests: alert rules fire when thresholds exceeded.

## Risks & Considerations

- **Correlation IDs:** must propagate across service boundaries (API → consumer).
- **Log volume:** structured JSON is verbose; ensure log rotation and retention.
- **Metrics cardinality:** avoid high-cardinality labels (e.g., user IDs as labels).
- **Readiness vs liveness:** readiness checks dependencies; liveness checks process health only.

## Cross-references

- Monitoring: [`docs/12-operations/monitoring-alerting.md`](../12-operations/monitoring-alerting.md)
- Kubernetes probes: [`docs/08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md#3-deployments--probes)
- Backend observability: [`docs/08-tech-stack/backend/observability.md`](../08-tech-stack/backend/observability.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_9_health_observability` (verifies observability deliverables)