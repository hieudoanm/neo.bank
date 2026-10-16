# Cloud Deployment (AWS)

## Architecture

**Components (PLANNED):**

- **EKS (Kubernetes)** — runs the `api-gateway`, `auth-service`, `account-service`, `transfer-service`, `transaction-service`, and `frontend` workloads.
- **MSK (Managed Streaming for Apache Kafka)** — Kafka cluster for the `transfers` topic and the `transfers.dlq` dead-letter topic.
- **Amazon Keyspaces (managed Cassandra)** — one keyspace per service: `auth_keyspace`, `account_keyspace`, `transfer_keyspace`, `transaction_keyspace`. Each service owns its own keyspace; there is no shared keyspace.
- **Application Load Balancer / Ingress** — traffic routing to the API gateway.
- **CloudWatch** — logging and metrics.
- **S3** — Terraform state (must use a globally unique bucket name).
- **Kubernetes manifests** under `infrastructure/kubernetes/` (see [`08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md)).

All AWS infrastructure and workloads described on this page are **PLANNED**; the repository contains no application code yet. Refer to the root `docs/` for canonical architecture and requirements.

## Terraform Structure

The GCP infrastructure tree lives in [Cloud Deployment (GCP)](./cloud-deployment-gcp.md).

```text
infrastructure/
├── aws/
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── vpc.tf
│   ├── eks.tf
│   ├── msk.tf
│   ├── keyspaces.tf
│   ├── alb.tf
│   └── cloudwatch.tf
└── kubernetes/
    ├── base/
    └── overlays/
```

## Terraform Configuration (AWS)

**main.tf (PLANNED):**

```hcl
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "neo-bank-terraform-state-<account-id>-<region>" # must be globally unique
    key            = "aws/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "neo-bank-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region
}
```

**variables.tf (PLANNED):**

```hcl
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "neo-bank"
}
```

**eks.tf (PLANNED):**

```hcl
resource "aws_eks_cluster" "main" {
  name     = "${var.app_name}-${var.environment}"
  role_arn = aws_iam_role.eks_cluster.arn
  version  = "1.29"

  vpc_config {
    subnet_ids              = aws_subnet.private[*].id
    endpoint_private_access = true
    endpoint_public_access  = false
  }
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.app_name}-nodes"
  node_role_arn   = aws_iam_role.eks_node.arn
  subnet_ids      = aws_subnet.private[*].id

  instance_types = ["m6i.large"]
  scaling_config {
    desired_size = 3
    min_size     = 2
    max_size     = 6
  }
}
```

**msk.tf (PLANNED):**

```hcl
resource "aws_msk_cluster" "cluster" {
  cluster_name           = "${var.app_name}-${var.environment}"
  kafka_version          = "3.5.0"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    client_subnets  = aws_subnet.private[*].id
    security_groups = [aws_security_group.msk.id]
  }

  configuration_info {
    arn      = aws_msk_configuration.config.arn
    revision = aws_msk_configuration.config.revision
  }
}

resource "aws_msk_configuration" "config" {
  kafka_versions = ["3.5.0"]
  name           = "${var.app_name}-config"

  server_properties = <<PROPERTIES
auto.create.topics.enable=false
default.replication.factor=3
min.insync.replicas=2
PROPERTIES
}
```

## Keyspaces (Amazon Keyspaces)

Each service owns its own Cassandra keyspace, as per the architecture. Keyspaces are provisioned per environment with per-service naming: `auth_keyspace`, `account_keyspace`, `transfer_keyspace`, and `transaction_keyspace`. No service reads or writes another service's keyspace.

**keyspaces.tf (PLANNED):**

```hcl
locals {
  service_keyspaces = [
    "auth_keyspace",
    "account_keyspace",
    "transfer_keyspace",
    "transaction_keyspace",
  ]
}

resource "aws_keyspaces_keyspace" "service" {
  for_each = toset(local.service_keyspaces)
  name     = each.value
}
```

**Keyspaces portability rules:**

- **Keyspace creation is Terraform, not CQL.** Do not run `CREATE KEYSPACE`: Keyspaces manages replication (`SingleRegionStrategy`, RF=3 across 3 AZs). Application migrations create **tables only**, inside the provisioned keyspace.
- **No `CREATE INDEX`.** Keyspaces rejects secondary indexes, so every non-primary-key access path is a dedicated lookup table (`accounts_by_user`, `users_by_email`, `transfers_by_account`) — see [`05-data-model.md`](../05-data-model.md#61-keyspace).
- **LWTs are supported** for `INSERT`/`UPDATE`/`DELETE` compare-and-set, so the conditional debit (`IF balance_cents >= ?`) and `processed_commands` claim (`IF NOT EXISTS`) are portable. Verify any new conditional against the [Keyspaces functional differences](https://docs.aws.amazon.com/keyspaces/latest/devguide/functional-differences.html) before relying on it.
- **No aggregate calls** (`COUNT`, `SUM`, `MIN`, `MAX`, `AVG`) and **no materialized views / triggers / UDFs** in queries — these are rejected by Keyspaces. Pagination uses the driver's paging state (decode the opaque state on the client); do not implement cursor pagination with `COUNT`.
- **Auth is IAM, not CQL.** Migrations and services authenticate to Keyspaces with IAM (SIGV4) credentials and least-privilege policies; there are no Cassandra `CREATE ROLE`/`GRANT` statements.
- Keyspaces DDL is **asynchronous** — wait for the table to reach `ACTIVE` before running migrations that depend on it.

## Authentication & IAM (OIDC)

CI/CD uses **OIDC with AWS IAM role-to-assume**, avoiding long-lived credentials. GitHub Actions authenticates to AWS via `aws-actions/configure-aws-credentials` using `role-to-assume` and a scoped trust policy restricted to this repository's workflows. No long-lived AWS access key credentials are stored as repository secrets.

The S3 Terraform state bucket name must be **globally unique**; use a placeholder such as `neo-bank-terraform-state-<account-id>-<region>` or an environment-specific unique suffix.

## Deployment Commands (PLANNED)

```bash
# Initialize Terraform
cd infrastructure/aws
terraform init

# Plan deployment (save the reviewed plan; never apply without reviewing it)
terraform plan -out=tfplan

# Apply the exact plan that was reviewed (interactive confirmation is bypassed
# because tfplan was reviewed; CI does this via a protected environment)
terraform apply tfplan

# Deploy workloads to the cluster
aws eks update-kubeconfig --name neo-bank-production
kubectl apply -k infrastructure/kubernetes/overlays/production

# Destroy deployment
terraform destroy
```

## Kubernetes Workload Deployment (AWS)

```bash
# Render and apply the manifests for an environment
kubectl apply -k infrastructure/kubernetes/overlays/production

# Watch the rollout; readiness gates traffic
kubectl rollout status deployment/api-gateway -n neo-bank

# Roll back if the new revision is unhealthy
kubectl rollout undo deployment/api-gateway -n neo-bank
```

Rules (see [`08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md)):

- Manifests are the source of truth; apply them from CI or a GitOps controller, never by hand in production.
- Cassandra migrations run as a `Job` **before** the new revision takes traffic.
- Images are pinned by digest; every workload defines readiness/liveness probes, resource requests and limits, and a non-root security context.
- The same manifests run on GKE; only the overlay differs.

## CI/CD for AWS

**GitHub Actions workflow:** `.github/workflows/deploy-aws.yml` (PLANNED)

```yaml
name: Deploy AWS

on:
  push:
    branches: [main]

permissions:
  id-token: write # required for OIDC
  contents: read

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v2
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_DEPLOY_ROLE_ARN }}
          aws-region: us-east-1
      - name: Terraform Plan
        run: |
          cd infrastructure/aws
          terraform init
          terraform fmt -check -recursive
          terraform validate
          terraform plan -out=tfplan
      - uses: actions/upload-artifact@v4
        with:
          name: tfplan
          path: infrastructure/aws/tfplan

  apply:
    needs: plan
    runs-on: ubuntu-latest
    environment: production # gates production with a manual approval
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v2
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_DEPLOY_ROLE_ARN }}
          aws-region: us-east-1
      - uses: actions/download-artifact@v4
        with:
          name: tfplan
          path: infrastructure/aws
      - name: Terraform Apply reviewed plan
        # Applies the exact plan produced by `plan` — never -auto-approve.
        run: |
          cd infrastructure/aws
          terraform init
          terraform apply tfplan
      - name: Deploy workloads to EKS
        run: |
          aws eks update-kubeconfig --name neo-bank-production
          kubectl apply -k infrastructure/kubernetes/overlays/production
          kubectl rollout status deployment/api-gateway -n neo-bank
```
