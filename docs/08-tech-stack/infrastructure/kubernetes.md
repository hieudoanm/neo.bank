# NEO Bank — Kubernetes Best Practices

_Infrastructure · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [infrastructure index](./README.md)._

Kubernetes orchestrates the NEO Bank workloads — the API gateway, the four Go services, and the frontend — using the same images built for Docker ([`docker.md`](./docker.md)). The cluster (EKS or GKE) is provisioned by Terraform ([`terraform.md`](./terraform.md)); the workloads are declared as manifests and applied by CI ([`github-actions.md`](./github-actions.md)).

## 1. Workloads

| Component            | Kind         | Notes                                                                        |
| -------------------- | ------------ | ---------------------------------------------------------------------------- |
| `api-gateway`        | `Deployment` | Stateless; routing, auth enforcement, rate limiting; behind a `Service` + `Ingress`. |
| `auth-service`       | `Deployment` | Stateless; owns `auth_keyspace`; scales horizontally.                         |
| `account-service`    | `Deployment` | Stateless; owns `account_keyspace`.                                          |
| `transfer-service`   | `Deployment` | Stateless; saga orchestrator; owns `transfer_keyspace`; bounded by Kafka partitions. |
| `transaction-service`| `Deployment` | Stateless; read model; owns `transaction_keyspace`.                          |
| `frontend`           | `Deployment` | Serves the built static assets (`dist/`); read-only.                         |

Rules:

- Everything is stateless; all state lives in Cassandra/Kafka (see [`cassandra.md`](../backend/cassandra.md), [`kafka.md`](../backend/kafka.md)).
- Cassandra and Kafka are **not** run as hand-rolled `StatefulSets`: use managed services (Keyspaces/MSK on AWS) or proven operators (Cassandra/Kafka on GKE on GCP), never bespoke stateful workloads.
- One container per Pod; no sidecars unless there is a documented reason.

## 2. Manifests & Tooling

Declarative manifests under `infrastructure/kubernetes/`, rendered per environment.

```text
infrastructure/kubernetes/
├── base/
│   ├── api-gateway.yaml
│   ├── auth-service.yaml
│   ├── account-service.yaml
│   ├── transfer-service.yaml
│   ├── transaction-service.yaml
│   ├── frontend.yaml
│   ├── service.yaml
│   └── kustomization.yaml
└── overlays/
    ├── staging/
    └── production/
```

Rules:

- `kubectl` + **Kustomize** for overlays; Helm only if a dependency genuinely needs templating.
- Base holds environment-agnostic manifests; overlays set replicas, resources, and hostnames.
- No `latest` tags and no `kubectl edit` in production — the repo is the source of truth (GitOps).
- All manifests are applied from CI or a GitOps controller, never by hand.

## 3. Deployments & Probes

```yaml
# base/api-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 2
  selector:
    matchLabels: { app: api-gateway }
  template:
    metadata:
      labels: { app: api-gateway }
    spec:
      containers:
        - name: api-gateway
          image: ghcr.io/neo-bank/api-gateway:${GIT_SHA}
          ports: [{ containerPort: 8080 }]
          readinessProbe:
            httpGet: { path: /readyz, port: 8080 }
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet: { path: /healthz, port: 8080 }
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests: { cpu: "100m", memory: "128Mi" }
            limits:   { cpu: "500m", memory: "512Mi" }
```

Rules:

- **Readiness** gates traffic; **liveness** restarts a wedged process. Never point liveness at a dependency check.
- Always set CPU/memory **requests and limits**; an unbounded Pod can starve the node.
- Use named ports and image tags pinned to the commit SHA.
- A `startupProbe` covers slow starts if the API legitimately needs longer than `initialDelaySeconds`.

## 4. Configuration & Secrets

```yaml
          envFrom:
            - configMapRef: { name: api-gateway-config }
          env:
            - name: JWT_PUBLIC_KEY_PEM
              valueFrom:
                secretKeyRef: { name: api-gateway-secrets, key: jwt_public_key }
```

Rules:

- Non-secret configuration lives in a `ConfigMap`; secrets live in a `Secret` synced from the cloud secret manager (External Secrets Operator / Secrets Store CSI), never committed to the repo.
- Inject configuration as environment variables; mount files only when a tool requires it.
- Rotating a secret must not require a rebuild; the workload reloads on restart.
- Never log secret values or dump the environment.

## 5. Scaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata: { name: api-gateway }
spec:
  scaleTargetRef: { apiVersion: apps/v1, kind: Deployment, name: api-gateway }
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource: { name: cpu, target: { type: Utilization, averageUtilization: 70 } }
```

Rules:

- HPA on CPU by default; add a custom metric (queue lag, request rate) only when it is the real driver.
- The **Kafka-consuming workloads** scale up to the partition count and no further; more replicas just idle.
- Set `minReplicas >= 2` for the gateway and the services so a node drain never takes a component down.
- Configure a `PodDisruptionBudget` for the gateway so rolling maintenance keeps capacity.

## 6. Networking

- One `Service` per workload; expose only the `api-gateway` externally via `Ingress`.
- The frontend is served through the ingress or a CDN; the gateway reaches each service by its in-cluster DNS name.
- Enforce `NetworkPolicy` (default deny, then allow specific east-west paths).
- Terminate TLS at the ingress with managed certificates; redirect HTTP to HTTPS.
- Set sensible timeouts and a bounded request body size at the ingress.

## 7. Migrations & Jobs

- Run each service's Cassandra migrations as a `Job` (or an `initContainer`) **before** that service's new revision takes traffic.
- Migrations are idempotent and additive (see [`cassandra.md`](../backend/cassandra.md)); a failed Job blocks the rollout.
- Never run schema changes from an application container at startup.

## 8. Rollouts & Rollback

- `strategy: RollingUpdate` with `maxUnavailable: 0` and `maxSurge: 1` for zero-downtime deploys.
- Readiness gates the rollout; a failing probe stops the progression.
- Rollback is `kubectl rollout undo` (or re-applying the previous image SHA); the previous `ReplicaSet` is retained.
- Blue-green via two Deployments + a Service selector switch when a change needs atomic cutover.

## 9. Observability

- Log structured JSON to **stdout/stderr**; the platform collects it. Never log to files inside the container.
- Carry the `correlationId` and `requestId` through every log line (see [`11-deployment.md`](../../11-deployment.md)).
- Expose Prometheus metrics on a dedicated port; add `prometheus.io/scrape` annotations.
- Alert on the metrics in [`11-deployment.md`](../../11-deployment.md), including consumer lag.

## 10. Security

- Run as **non-root** with a `securityContext` (`runAsNonRoot: true`, `readOnlyRootFilesystem: true`, drop all capabilities).
- Least-privilege `RBAC`; workloads get no cluster-admin.
- Pin images by digest in production; scan them in CI (Trivy) and block high-severity findings.
- Adopt the restricted `PodSecurity` standard; no privileged containers, no host namespaces.
- Secrets are mounted/`env`-only, never baked into images.

```yaml
          securityContext:
            runAsNonRoot: true
            runAsUser: 65532
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities: { drop: ["ALL"] }
```

## 11. Code Style & Rules

- Declarative manifests + Kustomize overlays; the repo is the source of truth (GitOps).
- Stateless workloads only; state lives in managed Cassandra/Kafka.
- Readiness and liveness probes on every workload; requests and limits always set.
- Config in ConfigMaps, secrets from the cloud secret manager, never committed.
- HPA for the gateway and services; Kafka-consumer scaling bounded by partitions; `minReplicas >= 2`; use a PDB.
- Migrations run as a Job before rollout; rollback via `rollout undo` or the previous SHA.
- Log to stdout, expose metrics, run non-root, pin images by digest.
- Related: [`docker.md`](./docker.md) for images, [`terraform.md`](./terraform.md) for the cluster, [`github-actions.md`](./github-actions.md) for the apply pipeline, [`11-deployment.md`](../../11-deployment.md) for the deployment strategy.
