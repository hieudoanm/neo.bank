# NEO Bank — Backend Best Practices

This is the entry point for the backend best-practice documents. Each technology has its own file; this README is the index and the shared rules that apply across all of them.

The **why** behind these choices (trade-offs and rejected alternatives) lives in the [tech stack decision record](../../08-tech-stack.md). These documents are the **how**.

## 1. Documents

| Document                         | Covers                                                          |
| -------------------------------- | --------------------------------------------------------------- |
| [`go.md`](./go.md)                   | Structure, layers, errors, config, testing, observability    |
| [`gin.md`](./gin.md)             | Gin router, middleware, request binding, error responses        |
| [`jwt.md`](./jwt.md)             | JWT access/refresh tokens, Argon2id, key rotation               |
| [`cassandra.md`](./cassandra.md) | Query-first modelling, LWTs, migrations, sessions               |
| [`kafka.md`](./kafka.md)         | Topic design, event contract, idempotent consumers, DLQ         |
| [`zookeeper.md`](./zookeeper.md) | Kafka coordination service: config, quorum, health checks |

Python scripts are linted and formatted with Ruff; that tooling is documented with the other cross-cutting tools in [`../tools/README.md`](../tools/README.md).

## 2. Shared Rules

These apply to every file in this folder:

1. **Financial operations never run in web request handlers.** The API validates input and publishes events to Kafka; Kafka consumers perform balance changes and record transactions.
2. **Amounts are integers in minor units (cents)**, never floating point.
3. **Balance updates use lightweight transactions (paxos)** and must never allow a negative balance.
4. **All queries are parameterised**; user input is never interpolated into CQL.
5. **Account-scoped access is verified by the application layer**, never by a raw client-supplied ID.
6. **Deny by default**; `404` is preferred over disclosing another account's resource.
7. **Every mutating operation enforces an idempotency key**; consumers are idempotent under duplicate delivery.
8. **Context** is passed through all call chains; **errors are wrapped** with `%w`.
9. **No global state** except configuration; connections are pooled and process-scoped.
10. **Observability**: structured JSON logs with correlation IDs and Prometheus metrics.

## 3. Toolchain

| Area      | Tool                                                       |
| --------- | ---------------------------------------------------------- |
| Language  | Go 1.22+                                                   |
| Router    | Gin                                                        |
| Database  | Apache Cassandra 4.x (gocql)                               |
| Messaging | Apache Kafka 3.x (segmentio/kafka-go)                      |
| Lint      | golangci-lint, go vet, staticcheck                         |
| Tests     | Go testing, testify (real Cassandra/Kafka via CI service containers) |

Run the backend checks through the root Makefile or CI pipeline. Never weaken a check to make CI pass.

## 4. Definition of Done (Backend)

1. Builds, vets, lints, and tests cleanly in CI.
2. Migrations apply cleanly; every query is parameterised and account-scoped.
3. No financial operation runs in a web request handler; consumers are idempotent.
4. All amounts are integer minor units; balance updates use lightweight transactions.
5. Authentication uses JWT (RS256) with Argon2id password hashing.
6. Every mutating operation enforces an idempotency key.
7. Duplicate delivery and redelivery are tested explicitly.

---

↑ Part of the [NEO Bank tech stack decision record](../../08-tech-stack.md).
