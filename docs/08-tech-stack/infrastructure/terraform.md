# NEO Bank — Terraform Best Practices

_Infrastructure · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [infrastructure index](./README.md)._

Terraform provisions the optional cloud layer on AWS and GCP. It is provider-neutral and never required for local development. The resources and full configuration live in [`11-deployment.md`](../../11-deployment.md).

## 1. Layout

One root module per cloud, with the shared shape kept identical.

```text
infrastructure/
├── aws/
│   ├── main.tf        # provider + backend configuration
│   ├── variables.tf   # input variables
│   ├── outputs.tf     # output values
│   ├── vpc.tf         # VPC, subnets, security groups
│   ├── eks.tf         # EKS cluster and node groups
│   ├── msk.tf         # Managed Streaming for Kafka
│   ├── keyspaces.tf   # Amazon Keyspaces (Cassandra-compatible)
│   ├── alb.tf         # Application Load Balancer
│   └── cloudwatch.tf  # Log groups, metric alarms
├── gcp/
│   ├── main.tf        # provider + backend configuration
│   ├── variables.tf   # input variables
│   ├── outputs.tf     # output values
│   ├── vpc.tf         # VPC, subnets, firewall rules
│   ├── gke.tf         # GKE cluster and node pools
│   ├── kafka.tf       # Kafka on GKE (Strimzi operator)
│   ├── cassandra.tf   # Cassandra on GKE
│   └── cloud_monitoring.tf  # Logs, metrics, alert policies
└── kubernetes/        # workload manifests, not Terraform (see kubernetes.md)
    ├── base/          # Kustomize base: Deployments, Services, NetworkPolicies
    └── overlays/      # Environment overlays: staging, production
```

Rules:

- One root module per cloud; no cross-cloud module that hides provider differences.
- Split by concern (`vpc`, `eks`/`gke`, `msk`/`kafka`, …), not one giant `main.tf`.
- Reusable patterns become modules; copy-paste between AWS and GCP is expected where providers differ.
- Workload manifests are Kubernetes (Kustomize), not Terraform — Terraform provisions the cluster and the managed data services it talks to.

## 2. Provider & Versions

```hcl
# infrastructure/aws/main.tf
terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.0" }
  }
}

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  host                   = var.eks_endpoint
  cluster_ca_certificate = var.eks_cluster_ca
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    arguments   = ["eks", "get-token", "--cluster-name", var.eks_cluster_name]
  }
}
```

Rules:

- Pin `required_version` and every provider version; never float.
- Commit `.terraform.lock.hcl` so provider selections are reproducible across runs and machines.
- Set the region (and GCP project) from a variable, not hard-coded.
- Kubernetes provider uses `exec` auth (AWS IAM / GCP gcloud) — no long-lived credentials in config.

## 3. Remote State

State is shared, encrypted, locked, and never in the repository.

### AWS (S3 + DynamoDB)

```hcl
terraform {
  backend "s3" {
    bucket         = "neo-bank-terraform-state"
    key            = "aws/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "neo-bank-terraform-locks"
  }
}
```

### GCP (GCS)

```hcl
terraform {
  backend "gcs" {
    bucket = "neo-bank-terraform-state"
    prefix = "gcp/terraform.tfstate"
  }
}
```

Rules:

- Remote backend per cloud (S3 + DynamoDB lock on AWS, GCS on GCP).
- Encryption and locking are mandatory; a shared state without a lock invites corruption.
- Never commit `terraform.tfstate` or `*.tfstate.backup`; add them to `.gitignore`.
- State can contain sensitive values — treat the bucket as confidential.
- Lock table prevents concurrent applies that corrupt state.

## 4. Variables & Outputs

- Every variable has a `description` and a `type`; defaults only where safe.
- Mark sensitive variables `sensitive = true` so they are not printed in plan output.
- Outputs expose only what another layer needs (for example, the ALB DNS name); avoid leaking internals.
- Validate inputs (`validation` blocks) so a bad value fails at plan time, not apply time.

```hcl
variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
  validation {
    condition     = contains(["us-east-1", "eu-west-1"], var.aws_region)
    error_message = "Region must be us-east-1 or eu-west-1."
  }
}

variable "jwt_private_key_arn" {
  description = "ARN of the JWT signing key in Secrets Manager"
  type        = string
  sensitive   = true
}
```

## 5. Secrets

- Secrets live in AWS Secrets Manager / GCP Secret Manager, referenced by ARN/ID — never in `.tf` files or state inputs.
- Grant least-privilege read access to the task/service role that needs each secret.
- Rotate signing keys and credentials out of band; Terraform references them, it does not own their values.
- Kubernetes Secrets reference the managed secret store via CSI driver or external-secrets operator.

## 6. Workflow

```bash
terraform fmt -check -recursive   # formatting gate
terraform init
terraform validate                # syntax and reference check
terraform plan -out=tfplan        # reviewable plan
terraform apply tfplan            # apply the reviewed plan
```

Rules:

- `fmt` and `validate` run in CI on every change; a diff in `fmt` fails the build.
- Always review a saved plan before apply; CI applies the exact plan that was reviewed.
- Apply is gated behind green CI and a protected environment; destroy is never automated on shared environments.
- Tag resources (`app`, `environment`, `owner`) for cost tracking and cleanup.
- Use workspaces or separate state files per environment (staging, production).

## 7. Safety

- Destructive changes (replacing stateful resources: databases, buckets) require explicit review and a backup step.
- `prevent_destroy` on stateful resources (Keyspaces, Cassandra on GKE, buckets) where accidental deletion is catastrophic.
- Blue-green for the API: apply the new task definition, shift traffic on health, keep the previous revision for rollback.
- Test `plan` against a sandbox workspace before touching production.
- Never run `terraform apply -auto-approve` in CI on production.

## 8. Code Style & Rules

- One root module per cloud; split by concern; pin versions; commit the lock file.
- Remote, encrypted, locked state; no state or secrets in the repository.
- Variables typed, described, and validated; sensitive ones marked.
- `fmt -check` and `validate` in CI; apply only a reviewed, saved plan.
- `prevent_destroy` on stateful resources; blue-green for the API.
- Cloud is optional — local development never depends on Terraform.
- Related: [`docker.md`](./docker.md) for images, [`github-actions.md`](./github-actions.md) for the apply pipeline, [`11-deployment.md`](../../11-deployment.md) for the architecture.
