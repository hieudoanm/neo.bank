# NEO Bank — Project Overview

## 1. Project Summary

**NEO Bank** is a full-stack banking application demonstrating core financial operations: account registration, authentication, balance inquiry, and fund transfers. The project showcases end-to-end software engineering across a React and TypeScript frontend, Go microservices with Cassandra and Kafka, containerised with Docker, orchestrated with Kubernetes (EKS on AWS, GKE on GCP), and deployed via Terraform on AWS and GCP.

NEO Bank is an independent portfolio project inspired by the engineering challenges of financial software. It must not copy proprietary software, branding, source code, or confidential workflows from any existing financial institution.

## 2. Problem Statement

Personal banking involves account lifecycle management, secure authentication, real-time balance visibility, and reliable fund transfers between accounts.

A useful banking platform needs to make it possible to:

- Register new bank accounts with secure credential management.
- Authenticate users and maintain session security.
- Display current account balance with transaction history context.
- Execute transfers between accounts with consistency guarantees (atomic per-account updates and saga compensation).
- Track transaction status, failures, and audit trails.
- Preserve data consistency across distributed systems.
- Control access to financial data with strict authorisation.
- Maintain reliable, secure, and auditable infrastructure.

NEO Bank brings these workflows together in one coherent application.

## 3. Product Vision

Build a polished, production-style banking platform that demonstrates how modern full-stack engineering can support core financial operations.

The product should be sufficiently complete that a reviewer can create an account, log in, view their balance, and transfer funds to another account without manually interacting with the database or backend.

The project should prioritise engineering quality, usability, security, and data consistency over the number of features.

## 4. Target Users

### 4.1 Account Holder

A person who owns a bank account and performs banking operations.

Primary needs:

- Register a new account with secure credentials.
- Log in and maintain authenticated sessions.
- View current balance and recent activity.
- Transfer funds to other accounts.
- Receive confirmation of completed transactions.

### 4.2 Developer or Operator

A person responsible for maintaining and deploying NEO Bank.

Primary needs:

- Run the application locally with Docker Compose.
- Diagnose failed requests and transaction processing.
- Apply database schema changes safely.
- Deploy and roll back application releases on AWS and GCP.
- Monitor application health, Kafka lag, and infrastructure costs.

## 5. Product Goals

### Goal 1: Deliver a complete banking workflow

A user can move from account creation to balance inquiry and fund transfers within the web application.

### Goal 2: Demonstrate full-stack ownership

The implementation covers frontend development, API design, distributed database modelling, event-driven architecture, infrastructure as code, testing, and multi-cloud deployment.

### Goal 3: Ensure financial data integrity

The application maintains balance consistency, transfer consistency (saga compensation across accounts), and audit trails for all financial operations.

### Goal 4: Establish secure authentication and authorisation

Users can only access their own accounts and perform authorised operations.

### Goal 5: Support event-driven transaction processing

Transfers are processed asynchronously via Kafka with idempotent consumers and exactly-once semantics.

### Goal 6: Provide a maintainable codebase

The repository has clear architectural boundaries, automated tests, reproducible setup instructions, and documented engineering decisions.

## 6. Non-Goals

The initial release will not:

- Process credit cards, loans, or investment products.
- Implement interest calculation or compounding.
- Support multi-currency accounts or foreign exchange.
- Integrate with external payment networks (SWIFT, ACH, card networks).
- Implement KYC/AML compliance workflows.
- Provide regulatory reporting or audit certifications.
- Support joint accounts, beneficiaries, or trust structures.
- Implement billing, subscriptions, or commercial pricing.
- Add services beyond the API gateway and the auth, account, transfer, and transaction services without a demonstrated need.

These features may be considered in future iterations, but they must not compromise the initial delivery.

## 7. Financial Data

The initial demonstration will use synthetic account data generated for testing and demonstration purposes.

All financial amounts are stored as integer minor units (cents) to avoid floating-point precision issues.

The application must preserve audit information for every transaction, including:

- Source and destination account identifiers.
- Transfer amount in minor units.
- Timestamp (ISO 8601 UTC).
- Transaction status (pending, processing, completed, failed, reversed).
- Idempotency key for duplicate detection.
- Correlation ID for tracing across services.

No real financial data or personally identifiable information should be used in the portfolio demonstration.

## 8. Primary User Workflow

The primary workflow is:

1. A user visits the application and registers a new account.
2. The system creates the account with initial balance (configurable default).
3. The user logs in with their credentials.
4. The user views their current balance.
5. The user initiates a transfer to another account by providing the destination account ID and amount.
6. NEO Bank validates the request (sufficient funds, valid destination, positive amount).
7. The transfer is published to Kafka for asynchronous processing.
8. The user sees the transfer status update from pending to completed.
9. Both source and destination balances are updated, each with its own atomic (LWT) statement; if the credit fails, the saga compensates by reversing the debit.
10. The user can review the transaction in their history.

Every stage should provide meaningful feedback, including loading states, validation errors, processing progress, and recovery instructions.

## 9. Technology Direction

The intended technology stack is:

- **Frontend:** React 18 with TypeScript, Vite, TanStack Query, React Router; styling with Tailwind CSS and DaisyUI.
- **Backend:** Go 1.22+ microservices with Gin router, Cassandra driver (gocql), Kafka client (segmentio/kafka-go).
- **Services:** API Gateway, Auth Service, Account Service, Transfer Service (saga orchestrator), and Transaction Service — each owning its own Cassandra keyspace (`auth_keyspace`, `account_keyspace`, `transfer_keyspace`, `transaction_keyspace`) and writing its own `audit_events`.
- **Database:** Apache Cassandra 4.x for account and transaction storage, one keyspace per service; no shared keyspace.
- **Message Queue:** Apache Kafka 3.x for transfer event processing (`transfers` topic, `transfers.dlq` dead-letter queue).
- **Containerisation:** Docker and Docker Compose for local development (gateway, four services, frontend, Cassandra, Kafka, ZooKeeper).
- **Orchestration:** Kubernetes is required — EKS on AWS, GKE on GCP.
- **Infrastructure:** Terraform for AWS (EKS, MSK, Keyspaces — one keyspace per service) and GCP (GKE, Cassandra on GKE, Kafka on GKE).
- **CI/CD:** GitHub Actions using OIDC (AWS role-to-assume, GCP workload identity federation); no static cloud credentials.
- **Testing:** Frontend unit and E2E tests (Vitest, Playwright), backend unit and integration tests, contract tests, against real Cassandra and Kafka.

Local commits use Husky and lint-staged to apply fast formatting and lint checks to staged source files. CI remains the authoritative full quality gate.

## 10. Security and Privacy Principles

Security must be implemented as part of the architecture, not added as a final polishing step.

The application must:

- Authenticate users before granting access to private resources.
- Enforce authorisation on the server for every account-scoped operation.
- Validate account ownership for every transfer operation.
- Use parameterised CQL queries (no string interpolation).
- Keep secrets out of source control and frontend bundles.
- Limit request sizes and validate all input.
- Use TLS for all service-to-service and client-to-server communication.
- Avoid logging credentials, access tokens, or full account numbers.
- Provide automated tests for unauthorised access attempts.
- Implement idempotency for all mutating operations.

NEO Bank is a portfolio demonstration and must not claim compliance with PCI DSS, SOX, GDPR, banking regulations, or other certifications without independently establishing the applicable requirements and evidence.

## 11. MVP Definition of Done

The canonical release definition of done lives in [`10-implementation-plan/definition-of-done.md`](10-implementation-plan/definition-of-done.md); this section summarises it and defers to it.

The minimum viable product is complete when:

- A developer can run the application using documented instructions (Docker Compose).
- A user can register a new account with email and password.
- A user can log in and receive a session token.
- A user can view their current balance.
- A user can transfer funds to another valid account.
- Transfers are processed asynchronously via Kafka with exactly-once semantics.
- Balance updates are atomic and consistent.
- Invalid transfers (insufficient funds, invalid account, negative amount) produce understandable errors.
- Automated tests cover the core workflow and security boundaries.
- CI runs the required quality checks (lint, typecheck, test).
- The architecture and operational limitations are documented.
- A cloud deployment can be reproduced using Terraform on AWS and GCP when credentials are available.

## 12. Success Criteria

The project should demonstrate the following outcomes:

| Area             | Success criterion                                                                 |
| ---------------- | --------------------------------------------------------------------------------- |
| User experience  | A reviewer can complete the primary workflow without developer assistance.        |
| Frontend         | Strictly typed React components with meaningful loading, error, and empty states. |
| Backend          | Documented REST endpoints with validated request and response models.             |
| Database         | Versioned CQL migrations, parameterised queries, and appropriate partitioning.    |
| Event processing | Idempotent Kafka consumers, bounded retries, and dead-letter handling.            |
| Security         | Automated negative tests for cross-account access and unauthorised operations.    |
| Testing          | Repeatable automated checks covering critical workflows.                          |
| DevOps           | Reproducible infrastructure and documented deployment procedures for AWS and GCP. |
| Portfolio        | A clear README, architecture diagram, demo, and documented trade-offs.            |

## 13. Guiding Principles

1. **Working software first:** deliver a complete vertical slice before expanding the scope.
2. **Financial integrity:** preserve audit trails and distinguish display formatting from stored precision.
3. **Secure by default:** deny access unless authentication and authorisation checks succeed.
4. **Explicit contracts:** use documented API models and database migrations.
5. **Reliable asynchronous processing:** assume messages may be delivered more than once or fail midway; design for idempotency.
6. **Simple architecture:** keep the decomposition small and cohesive — one API gateway plus the auth, account, transfer, and transaction services — and avoid distributed complexity beyond the documented saga and event flow.
7. **Reproducibility:** record the inputs, parameters, and versions needed to understand a transaction.
8. **Evidence over claims:** document only features that have actually been implemented and tested.
9. **Incremental delivery:** keep every implementation milestone runnable and reviewable.
10. **No financial advice:** present NEO Bank as a software demonstration, not a financial product.
