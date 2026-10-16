# NEO Bank — Infrastructure Best Practices

This is the entry point for the infrastructure best-practice documents. Each concern has its own file; this README is the index and the shared rules that apply across all of them.

The **why** behind these choices (trade-offs and rejected alternatives) lives in the [tech stack decision record](../../08-tech-stack.md). These documents are the **how**; the operational runbook is [`11-deployment.md`](../../11-deployment.md).

## 1. Documents

| Document                                     | Covers                                                       |
| -------------------------------------------- | ------------------------------------------------------------ |
| [`docker.md`](./docker.md)                   | Dockerfiles, Compose topology, healthchecks, images          |
| [`kubernetes.md`](./kubernetes.md)           | Workloads, probes, scaling, secrets, rollouts, hardening     |
| [`terraform.md`](./terraform.md)             | Layout, remote state, modules, secrets, plan/apply           |
| [`github-actions.md`](./github-actions.md)   | CI/CD workflows, service containers, caching, deploy gates   |

## 2. Shared Rules

These apply to every file in this folder:

1. **Reproducibility by default.** Pinned image tags, pinned provider/tool versions, and infrastructure as code. No manual provisioning.
2. **Same images everywhere.** Local Compose, CI, and the cluster run the identical build output; only configuration differs.
3. **Stateless workloads.** Application containers hold no state; all state lives in Cassandra and Kafka. Kubernetes orchestrates, managed services store.
4. **No secrets in the repository.** Secrets come from `.env` locally and a managed secret store in the cloud; `.env.example` holds placeholders only.
5. **Least privilege.** Minimal IAM/RBAC roles, private subnets, network policies, and security rules scoped to actual need.
6. **Fail fast.** Healthchecks and readiness probes gate startup and traffic; a bad deploy is caught before it serves a request.
7. **Idempotent and reversible.** Re-running a pipeline is safe; every deploy has a documented rollback.
8. **Everything declared, nothing clicked.** Any change to infrastructure is reviewed as code before it is applied.
9. **Observable.** Structured logs and metrics flow from every service; alerts are defined, not improvised (see [`11-deployment.md`](../../11-deployment.md)).
10. **Cloud is optional.** Local development uses Docker Compose; Terraform provisions cloud resources only for production/pre-production.

## 3. Toolchain

| Area            | Tool                                                        |
| --------------- | ----------------------------------------------------------- |
| Containers      | Docker, Docker Compose                                      |
| Orchestration   | Kubernetes (`kubectl`, Kustomize) on EKS / GKE              |
| IaC             | Terraform (AWS provider, Google provider)                   |
| Cloud (AWS)     | EKS, MSK (Kafka), Keyspaces (Cassandra), ALB, CloudWatch, S3 |
| Cloud (GCP)     | GKE, Cassandra on GKE, Kafka on GKE, Cloud Monitoring, GCS  |
| CI/CD           | GitHub Actions                                              |
| Image scanning  | Trivy                                                       |

## 4. Repository Layout

The repository layout is defined in [`11-deployment.md`](../../11-deployment.md):

```text
neo.bank/
├── packages/
│   ├── backend/
│   │   ├── api-gateway/     # Go gateway: routing, auth enforcement, rate limiting
│   │   ├── auth/            # Go service: auth_keyspace
│   │   ├── accounts/        # Go service: account_keyspace
│   │   ├── transfer/        # Go service: saga orchestrator, transfer_keyspace
│   │   └── transactions/    # Go service: read model, transaction_keyspace
│   └── frontend/
│       └── app/             # React + TypeScript SPA
├── infrastructure/
│   ├── aws/              # Terraform: EKS, MSK, Keyspaces, ALB
│   ├── gcp/              # Terraform: GKE, Cassandra, Kafka
│   └── kubernetes/       # Kustomize base + overlays per environment
├── docker-compose.yaml   # Local development topology
├── Makefile              # Convenience commands
├── .github/workflows/    # CI/CD pipelines
└── docs/                 # Normative project documentation
```

## 5. Environment Tiers

| Tier       | Platform      | Purpose                          | Infrastructure |
|------------|---------------|----------------------------------|----------------|
| Local      | Docker Compose | Developer ergonomics             | Compose, no cloud |
| CI         | GitHub Actions | Quality gates and tests          | Service containers |
| Staging    | Kubernetes     | Pre-production validation        | EKS/GKE + managed services |
| Production | Kubernetes     | Customer-facing workloads        | EKS/GKE + managed services |

Local development never depends on cloud infrastructure. Staging and production are provisioned via Terraform and deployed to Kubernetes.

## 6. Definition of Done (Infrastructure)

1. The full topology starts locally with one command (`docker-compose up -d`) and passes healthchecks.
2. Images are multi-stage, pinned, non-root, and reproducibly built.
3. Terraform is formatted (`terraform fmt -check`), validated (`terraform validate`), and planned in CI before apply.
4. Remote state is encrypted and locked; no secrets or state are committed.
5. CI runs lint, typecheck, unit, integration, and E2E gates on real Cassandra and Kafka service containers.
6. Deploys are gated on green CI and use least-privilege, OIDC-based cloud credentials.
7. Workloads run on Kubernetes with readiness/liveness probes, resource limits, and non-root security contexts; images are pinned by digest.
8. Rollback is documented and tested (previous image / `rollout undo`, Cassandra snapshot).
9. NetworkPolicies enforce least-privilege pod-to-pod communication (default deny).
10. Monitoring and alerting are configured before the workload accepts traffic.

---

↑ Part of the [NEO Bank tech stack decision record](../../08-tech-stack.md).