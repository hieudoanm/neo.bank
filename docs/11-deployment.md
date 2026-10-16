# Deployment

## Purpose

This document defines the deployment strategy for NEO Bank, including local development, CI/CD, and cloud deployment on AWS and GCP.

The deployment strategy must ensure:

- Reproducible local development with Docker Compose.
- Automated CI/CD with quality gates.
- Secure cloud deployment with Terraform.
- Kubernetes (EKS on AWS, GKE on GCP) as the workload orchestrator.
- Multi-cloud support (AWS and GCP).
- Operational best practices (health checks, monitoring, backup).

## Scope

This document covers:

- Local development environment setup and workflow
- CI/CD pipeline configuration with GitHub Actions
- Cloud infrastructure provisioning with Terraform for AWS and GCP
- Kubernetes workload deployment and management

## Deployment Sections

- [Local Development](./11-deployment/local-development.md)
- [CI/CD Pipeline](./11-deployment/ci-cd-pipeline.md)
- [Cloud Deployment (AWS)](./11-deployment/cloud-deployment-aws.md)
- [Cloud Deployment (GCP)](./11-deployment/cloud-deployment-gcp.md)
