# NEO Bank — Implementation Plan — Phase 11: Cloud Infrastructure

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement Terraform for AWS and GCP cloud deployment. Phase 11 provisions production infrastructure and deploys the complete MVP to the cloud.

**Week:** 8+

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Terraform (AWS) | EKS, MSK, Keyspaces, ALB, CloudWatch, S3 |
| Terraform (GCP) | GKE, Kafka on GKE (Strimzi), Cassandra on GKE, Cloud Load Balancing, Cloud Monitoring, GCS |
| Kubernetes manifests | `api`, `consumer`, `frontend` Deployments, Services, Ingress |
| Deployment documentation | [`docs/11-deployment.md`](../11-deployment.md) complete |
| Staging deployment | Validated deployment to staging environment |

## Tasks

- [ ] Design AWS architecture (EKS, MSK, Keyspaces, ALB, VPC, subnets).
- [ ] Design GCP architecture (GKE, Kafka on GKE (Strimzi), Cassandra on GKE, Cloud Load Balancing).
- [ ] Implement Terraform for AWS (modules, remote state, secrets).
- [ ] Implement Terraform for GCP (modules, remote state, secrets).
- [ ] Implement Kubernetes manifests (readiness/liveness probes wired to `GET /healthz` and `GET /readyz`, resources, non-root security).
- [ ] Write deployment documentation (`11-deployment.md`).
- [ ] Test deployment to staging environment (validate end-to-end).
- [ ] Configure OIDC-based cloud credentials for CI deploys.
- [ ] Set up monitoring and alerting in cloud (CloudWatch / Cloud Monitoring).

## Acceptance Criteria

- Terraform applies successfully for AWS (no errors, idempotent).
- Terraform applies successfully for GCP (no errors, idempotent).
- Application deploys and runs on cloud infrastructure (all workloads healthy).
- Health checks pass in the cloud environment.
- Deployment documentation complete (CI, Docker Compose, Kubernetes, Terraform).
- Staging environment deployment validated end-to-end.
- Kubernetes readiness/liveness probes use the Phase 9 health endpoints (`GET /healthz`, `GET /readyz`).

## Exit Criterion

Phase 11 is optional. It is either explicitly deferred (MVP accepted at Phase 10) or considered complete only when Terraform applies cleanly for AWS and GCP, the staging deployment passes end-to-end validation, and all cloud workloads are healthy.

## Dependencies

Phase 10 (complete, documented MVP ready for cloud deployment).

## Testing

- Terraform validation: `terraform validate`, `terraform fmt -check`, `terraform plan`.
- Kubernetes dry-run: `kubectl apply --dry-run=server` for all manifests.
- Staging validation: end-to-end test against deployed staging environment.
- Health check verification: all workloads pass readiness/liveness probes.

## Risks & Considerations

- **Cloud provider differences:** AWS and GCP have different managed service equivalents (Amazon Keyspaces (AWS) vs Cassandra on GKE (GCP), MSK vs Kafka on GKE (Strimzi)).
- **Remote state:** Terraform state encrypted and locked; never committed to the repo.
- **Secrets:** never commit secrets; use cloud secret managers with OIDC access.
- **Cost:** cloud infrastructure incurs ongoing costs; tear down staging when not needed.

## Cross-references

- Terraform: [`docs/08-tech-stack/infrastructure/terraform.md`](../08-tech-stack/infrastructure/terraform.md)
- Kubernetes: [`docs/08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md)
- Deployment: [`docs/11-deployment.md`](../11-deployment.md)
- Operations: [`docs/12-operations.md`](../12-operations.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_11_cloud_infrastructure` (verifies cloud deployment deliverables)