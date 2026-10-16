# Security

## Purpose

Defines the security operations for NEO Bank in production: secrets management, network security, TLS, and access control. These controls protect financial data and banking operations.

## 1. Secrets Management

### AWS

- AWS Secrets Manager for JWT signing keys, database credentials, and API secrets.
- IAM roles for service-to-service authentication (no long-lived access keys).
- Secrets rotated on a 90-day schedule or immediately on suspected compromise.
- Applications access secrets via the AWS Secrets Manager SDK or Secrets Store CSI Driver.

### GCP

- Secret Manager for JWT signing keys, database credentials, and API secrets.
- Service accounts for service-to-service authentication (workload identity).
- Secrets rotated on a 90-day schedule or immediately on suspected compromise.
- Applications access secrets via the Secret Manager SDK or Secrets Store CSI Driver.

### Rules

- Secrets never stored in the repository, logs, or frontend bundles.
- `.env.example` holds placeholders only; `.env*` files are git-ignored.
- Local development uses `.env` files; cloud uses managed secret stores.
- Rotation does not require a rebuild; workloads reload on restart or via sidecar injection.
- CI/CD uses OIDC federation only — GitHub Actions assumes an AWS IAM role and uses GCP workload identity federation. No long-lived static cloud keys in CI, repositories, or secrets stores.

## 2. Network Security

### AWS

- VPC with private subnets for EKS nodes and pods (no public ingress to workloads).
- Security groups with least-privilege rules (ingress only from ALB, egress to managed services).
- Network policies enforce default-deny pod-to-pod communication.
- Ingress with TLS termination at the ALB; HTTP redirected to HTTPS.

### GCP

- VPC with private subnets for GKE nodes and pods (no public ingress to workloads).
- Firewall rules and network policies with least-privilege rules.
- Cloud Load Balancing with TLS termination; HTTP redirected to HTTPS.

### Rules

- All workloads in private subnets; no public IP addresses on pods.
- Ingress only through the load balancer; direct pod access blocked.
- East-west traffic between API and consumer restricted to necessary ports.
- Database and message broker accessible only from within the VPC.

## 3. TLS

- Enforce TLS 1.2+ for all HTTP traffic (internal and external).
- Use managed certificates (AWS ACM, GCP Certificate Manager).
- Rotate certificates automatically before expiry.
- HSTS enabled on all public endpoints.
- No plaintext HTTP in production; all traffic encrypted in transit.

## 4. Access Control

- **Deny by default.** Authenticate before authorising on every protected endpoint.
- `404` preferred over `403` for unauthorised access (prevents account enumeration).
- RBAC with least-privilege roles; no cluster-admin for workloads.
- Service accounts scoped to specific namespaces and resources.
- API tokens and credentials scoped to minimum required permissions.

## 5. Security Monitoring

- Failed authentication attempts logged and alerted (rate limit at 5 failures / 15 min).
- Unauthorised access attempts (404/401) logged with correlation IDs.
- Trivy image and filesystem scans in CI; high-severity findings block merge.
- Dependency vulnerability scanning (go mod, pnpm audit) in CI.
- Regular security patching for base images and dependencies.

## 6. Incident Response

- Security incidents classified by severity (critical, high, medium, low).
- Critical incidents (data breach, credential leak) trigger immediate response.
- Incident response includes: containment, eradication, recovery, post-mortem.
- Credentials rotated immediately on suspected compromise.
- Affected users notified per applicable data protection requirements.

## Cross-references

- Security rules: `../../AGENTS.md`
- Kubernetes security: [`docs/08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md#10-security)
- CI security: [`docs/08-tech-stack/infrastructure/github-actions.md`](../08-tech-stack/infrastructure/github-actions.md#5-security--permissions)
- Operations index: [`12-operations.md`](../12-operations.md)
- Planned test: `test_operations_security` (verifies security controls are documented)