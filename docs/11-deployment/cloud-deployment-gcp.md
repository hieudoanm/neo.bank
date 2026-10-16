# Cloud Deployment (GCP)

## Architecture

**Components (PLANNED):**

- **GKE (Kubernetes)** — runs `api-gateway`, `auth-service`, `account-service`, `transfer-service`, `transaction-service`, and `frontend` workloads.
- **Cassandra on GKE** — Apache Cassandra deployed to GKE (StatefulSet/operator, e.g., K8ssandra or official manifests) with per-service keyspaces (`auth_keyspace`, `account_keyspace`, `transfer_keyspace`, `transaction_keyspace`).
- **Kafka on GKE (Strimzi)** — Apache Kafka deployed via the Strimzi operator on GKE. Topics: `transfers` and DLQ `transfers.dlq`.
- **Cloud Load Balancing** — traffic routing to GKE Ingress/Services.
- **Cloud Monitoring** — logging and metrics.
- **Cloud Storage (GCS)** — Terraform state.
- **Kubernetes manifests** — under `infrastructure/kubernetes/` (see [`08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md)).

All components are **PLANNED**; the repository currently contains no application code or deployed infrastructure.

## Terraform Configuration (GCP)

**main.tf (PLANNED):**

```hcl
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "neo-bank-terraform-state-<env>-<suffix>" # Must be globally unique
    prefix = "gcp/terraform"
  }
}

provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
}
```

**variables.tf (PLANNED):**

```hcl
variable "gcp_project" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "gcp_zone" {
  description = "GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "environment" {
  description = "Environment name (dev/prod)"
  type        = string
  default     = "dev"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "neo-bank"
}

variable "cassandra_replicas" {
  description = "Number of Cassandra replicas (dev: 1, prod: 3)"
  type        = number
  default     = 1
}

variable "kafka_brokers" {
  description = "Kafka broker count (dev: 3 small, prod: 3+)"
  type        = number
  default     = 3
}
```

**Required variables:** `gcp_project` must be supplied (no default). `environment` selects the dev/prod sizing; `cassandra_replicas` and `kafka_brokers` default to the dev profile (1 Cassandra replica, 3 small Kafka brokers) and should be raised for production (Cassandra RF=3, Kafka RF≥3).

**gke.tf (PLANNED):**

```hcl
resource "google_container_cluster" "main" {
  name     = "${var.app_name}-${var.environment}"
  location = var.gcp_region
  
  remove_default_node_pool = true
  initial_node_count       = 1
  network                  = google_compute_network.main.name
  subnetwork               = google_compute_subnetwork.main.name
  workload_identity_config {
    workload_pool = "${var.gcp_project}.svc.id.goog"
  }
}

resource "google_container_node_pool" "main" {
  name       = "${var.app_name}-nodes"
  cluster    = google_container_cluster.main.name
  location   = var.gcp_region
  node_count = var.environment == "prod" ? 3 : 1

  node_config {
    machine_type = var.environment == "prod" ? "e2-standard-4" : "e2-standard-2"
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }
}

# Workloads (api-gateway, auth-service, account-service, transfer-service,
# transaction-service, frontend) live in infrastructure/kubernetes/ and
# are applied with kubectl / a GitOps controller after the cluster exists.
# Same manifests as AWS; only the overlay differs.
```

## Data Layer (Cassandra on GKE + Kafka on GKE)

- **Cassandra on GKE (PLANNED):** Deploy Apache Cassandra as a StatefulSet (or via an operator such as K8ssandra) in the GKE cluster. Each service owns its own keyspace:
  - `auth_keyspace` (auth-service)
  - `account_keyspace` (account-service)
  - `transfer_keyspace` (transfer-service)
  - `transaction_keyspace` (transaction-service)
- **Kafka on GKE (Strimzi, PLANNED):** Deploy Kafka using the Strimzi operator. Configure topics `transfers` and `transfers.dlq` with appropriate retention and replication (dev: 1 replica, prod: 3 replicas).

## Authentication & IAM (Workload Identity Federation)

GitHub Actions uses **GCP Workload Identity Federation** to authenticate to GCP for Terraform and deployments, avoiding long-lived credentials. CI configures identity federation with a GCP workload identity pool/provider linked to GitHub repository conditions. Kubernetes workloads use **GKE Workload Identity** to access GCP resources without node-wide credentials.

## Deployment Commands (PLANNED)

```bash
# Initialize Terraform
cd infrastructure/gcp
terraform init

# Plan deployment (save the reviewed plan; never apply without reviewing it)
terraform plan -out=tfplan

# Apply the exact plan that was reviewed (CI applies the reviewed tfplan behind
# a protected production environment; never -auto-approve)
terraform apply tfplan

# Deploy workloads to the cluster (same manifests as AWS; overlay differs)
gcloud container clusters get-credentials neo-bank-production --region us-central1
kubectl apply -k infrastructure/kubernetes/overlays/production

# Destroy deployment — interactive confirmation only; never automated on shared environments
terraform destroy
```

## Instance Guidance

| Environment | GKE Nodes | Cassandra | Kafka (Strimzi) |
|-------------|-----------|-----------|-----------------|
| **dev**     | 1 node (e2-standard-2) | 1 replica (dev) | 3 brokers (small), single AZ acceptable |
| **prod**    | 3 nodes (e2-standard-4+) | 3 replicas (RF=3) | 3 brokers minimum (RF>=3), multi-AZ |

GCS bucket names for Terraform state must be globally unique; use placeholders like `neo-bank-terraform-state-<env>-<suffix>`.
