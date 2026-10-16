# CI/CD Pipeline

## 3.1 GitHub Actions Workflows

The NEO Bank project uses GitHub Actions for continuous integration and deployment. There are three main workflows:
1. CI workflow (`.github/workflows/ci.yml`) - runs on pull requests and pushes to main/develop
2. AWS Deployment workflow (`.github/workflows/deploy-aws.yml`) - deploys to AWS on main branch pushes
3. GCP Deployment workflow (`.github/workflows/deploy-gcp.yml`) - deploys to GCP on main branch pushes

### CI Workflow (`.github/workflows/ci.yml`)

#### Backend Job
Runs on `ubuntu-latest` with the following steps (PLANNED - no code exists yet):

1. **Checkout** (`actions/checkout@v4`)
   - Retrieves the repository source code
   - Fetch depth: 1 (for efficiency)

2. **Setup Go** (`actions/setup-go@v5`)
   - Installs Go 1.22.x
   - Configures GOPATH and GOBIN

3. **Run golangci-lint** (`golangci/golangci-lint-action@v3`)
   - Executes static analysis with default configuration

4. **Run go vet** (`go vet ./...`)
   - Static analysis to identify potential issues

5. **Run go build** (`go build ./...`)
   - Verifies all packages compile successfully

6. **Check gofmt** (`gofmt -l . | wc -l` / fail if non-zero)
   - Ensures code is properly formatted

7. **Run Tests with Coverage**
   ```bash
   go test ./... -coverprofile=coverage.out
   go tool cover -func=coverage.out
   ```
   - Runs all unit tests across backend services
   - Requires **80%+ minimum backend coverage**

8. **Upload Coverage** (`codecov/codecov-action@v3`)
   - Sends coverage data to Codecov.io
   - Enables branch coverage tracking
   - Shows coverage changes in PR comments

9. **Run Trivy Security Scan** (`aquasecurity/trivy-action@master`)
   - Scans the filesystem for dependencies (`scan-type: 'fs'`) and the built container image (`scan-type: 'image'`)
   - Checks OS packages and application dependencies
   - Outputs results in SARIF format for GitHub integration
   - **Fails on Critical/High severity vulnerabilities** (both dependency and container scans)

10. **Upload Trivy Results** (`github/codeql-action/upload-sarif@v2`)
   - Makes security scan results visible in GitHub UI
   - Enables code scanning alerts
   - Integrates with GitHub's security tab

#### Frontend Job
Similar structure with Node.js-specific steps:

1. **Checkout** (`actions/checkout@v4`)
2. **Setup pnpm** (`pnpm/action-setup@v2`, version 8)
3. **Setup Node** (`actions/setup-node@v4`, version 18.x with cache)
4. **Install Dependencies** (`pnpm install --frozen-lockfile`)
5. **Run Linter** (`pnpm lint`)
6. **Run Format Check** (`pnpm format:check`)
7. **Run Type Check** (`tsc --noEmit`)
8. **Run Tests** (`pnpm test -- --coverage`)
9. **Build** (`pnpm build`)

#### Integration Tests Job
Runs tests requiring external services. The CI workflow must **start real Cassandra and Kafka services** (e.g., via GitHub Actions service containers for `cassandra:4.1` and `confluentinc/cp-kafka:7.4.0` with `zookeeper`) as defined in the testing strategy.

1. **Checkout** and **Setup Go** (same as backend)
2. **Start Cassandra and Kafka** (real service containers) and wait for readiness (polling loop with a bounded timeout)
3. **Run Migrations** (PLANNED: `go run cmd/migrate/main.go` per service)
4. **Run Integration Tests** (`go test ./tests/integration/...`)
   - Assumes duplicate message delivery and verifies consumer idempotency against the real Kafka broker

#### End-to-End Tests Job
Tests complete user workflows:

1. **Checkout**, **Setup pnpm**, **Setup Node** (same as frontend)
2. **Install Dependencies** (`pnpm install --frozen-lockfile`)
3. **Install Playwright** (`pnpm exec playwright install --with-deps`)
   - Installs Chromium, Firefox, and WebKit browsers
   - Includes system dependencies for headless operation
4. **Run E2E Tests** (`pnpm test:e2e`)
   - Tests critical user journeys: registration → login → transfer → history
   - Runs against Docker Compose environment
5. **Upload Playwright Report** (`actions/upload-artifact@v3`)
   - Preserves test videos, traces, and screenshots
   - Available as downloadable artifact in workflow run

### Quality Gates Enforced in CI

All jobs must pass for the workflow to succeed:

1. **Test Coverage**: Minimum **80% backend** line coverage, **70% frontend** line coverage
2. **Linting**: Zero warnings/errors from `golangci-lint`, `go vet`, `gofmt` check, and `pnpm lint` + `pnpm format:check`
3. **Type Checking**: No TypeScript compilation errors (`tsc --noEmit`)
4. **Build**: `go build ./...` must succeed for all backend packages
5. **Security Scanning**: No **Critical/High** severity vulnerabilities (container and dependency scans)
6. **Build Success**: Both backend and frontend must compile successfully
7. **Test Passing**: All unit, integration, and E2E tests must pass
8. **Terraform**: `terraform fmt -check`, `terraform validate`, and `terraform plan -out=tfplan` must pass for changed infra; apply uses the reviewed `tfplan` file (never `-auto-approve`)
9. **Production Deploy Approval**: Production deployments require **manual approval** before applying changes

### AWS Deployment Workflow (`.github/workflows/deploy-aws.yml`)

Triggers on pushes to the `main` branch. Authenticates with **OIDC** (no static credentials):

1. **Checkout** (`actions/checkout@v4`)
2. **Setup Terraform** (`hashicorp/setup-terraform@v2`, version 1.8+)
3. **Configure AWS credentials via OIDC** (`aws-actions/configure-aws-credentials@v4` with `role-to-assume`; requires `permissions: id-token: write`)
4. **Terraform Init** (`terraform init` in `infrastructure/aws/`)
5. **Terraform Format/Validate/Plan** (`terraform fmt -check`, `terraform validate`, `terraform plan -out=tfplan`), then upload `tfplan` as a workflow artifact
6. **Manual Approval** (GitHub production environment protection rule)
7. **Terraform Apply** (`terraform apply tfplan`) — applies the exact plan that was reviewed. Never `-auto-approve`
   - Provisions EKS cluster, MSK cluster, Amazon Keyspaces (one keyspace per service), VPC, IAM roles
8. **Deploy Workloads to EKS**
   ```bash
   aws eks update-kubeconfig --name neo-bank-production
   kubectl apply -k infrastructure/kubernetes/overlays/production
   kubectl rollout status deployment/api-gateway -n neo-bank
   ```

### GCP Deployment Workflow (`.github/workflows/deploy-gcp.yml`)

Similar structure with GCP-specific steps. Authenticates with **Workload Identity Federation**, avoiding long-lived credentials:

1. **Checkout** (`actions/checkout@v4`)
2. **Setup Terraform** (`hashicorp/setup-terraform@v2`)
3. **Authenticate to GCP** (`google-github-actions/auth@v2` with Workload Identity Federation; requires `permissions: id-token: write`)
4. **Terraform Init** (`terraform init` in `infrastructure/gcp/`)
5. **Terraform Format/Validate/Plan** (`terraform fmt -check`, `terraform validate`, `terraform plan -out=tfplan`), then upload `tfplan` as a workflow artifact
6. **Manual Approval** (GitHub production environment protection rule)
7. **Terraform Apply** (`terraform apply tfplan`) — applies the exact plan that was reviewed. Never `-auto-approve`
   - Provisions GKE cluster, Cassandra on GKE, Kafka on GKE (Strimzi), VPC, service accounts
8. **Deploy Workloads to GKE** (PLANNED)
   ```bash
   gcloud container clusters get-credentials neo-bank-production --region us-central1
   kubectl apply -k infrastructure/kubernetes/overlays/production
   kubectl rollout status deployment/api-gateway -n neo-bank
   kubectl rollout status deployment/auth-service -n neo-bank
   kubectl rollout status deployment/account-service -n neo-bank
   kubectl rollout status deployment/transfer-service -n neo-bank
   kubectl rollout status deployment/transaction-service -n neo-bank
   kubectl rollout status deployment/frontend -n neo-bank
   ```

## 3.2 Pipeline Triggers and Branch Strategy

### Branch Protection Rules
- `main` branch: Requires PR reviews, status checks, linear history
- `develop` branch: Requires PR reviews and status checks
- Feature branches: Require PR reviews before merging to develop

### Workflow Triggers
- **CI Workflow**: 
  - `push` to `main` and `develop` branches
  - `pull_request` targeting `main` and `develop` branches
- **Deployment Workflows**:
  - `push` to `main` branch only (production deployments)
  - Manual workflow_dispatch option available for hotfixes

### Concurrent Workflow Limits
- Maximum 3 concurrent workflow runs per repository
- Queued runs execute in FIFO order
- Workflow concurrency groups prevent duplicate runs on same branch

## 3.3 Artifacts and Caching

### Dependency Caching
- **Go modules**: Cached via `actions/setup-go` (default behavior)
- **Node modules**: pnpm store caching via `pnpm/action-setup@v2`
- **Playwright browsers**: Cached between workflow runs (reduces install time)

### Artifact Retention
- **Test reports**: Retained for 90 days
- **Build artifacts**: Retained for 30 days
- **Deployment logs**: Retained for 365 days (compliance requirement)
- **Security reports**: Retained for 365 days (audit requirement)

### Artifact Types
- `playwright-report`: Videos, traces, screenshots from E2E tests
- `coverage-report`: HTML coverage reports from backend tests
- `lint-report`: SARIF files from golangci-lint and ESLint
- `trivy-results`: Security scan results in SARIF format
- `terraform-logs`: Output from terraform apply/destroy operations

## 3.4 Performance Optimization

### Matrix Testing Strategies
- Backend tests run in parallel using Go's built-in test parallelization
- Frontend tests utilize Vitest's built-in parallelization
- Integration tests run sequentially due to shared resource dependencies

### Early Failure Detection
- Linting runs first in each job (fast feedback)
- Unit tests run before integration tests
- Security scanning runs after build completion

### Resource Optimization
- Use of `actions/checkout@v4` with `fetch-depth: 1` for most jobs
- Selective caching based on file changes (using actions/cache when beneficial)
- Job-level timeout settings (default 1 hour, extendable for specific needs)

## 3.5 Monitoring and Notifications

### Workflow Status Badges
- README.md displays CI status badge from GitHub Actions
- Separate badges for backend, frontend, integration, and e2e tests
- Deployment workflows show last successful deployment time

### Failure Notifications
- GitHub automatically notifies PR authors of workflow failures
- Optional: Slack/MS Teams notifications via repository webhooks
- Email notifications for workflow failures (repository setting)

### Debugging Failed Workflows
- Re-run workflow with debug logging enabled
- Use `actions/upload-artifact` to preserve intermediate states

## 3.6 Best Practices and Guidelines

### Commit Message Conventions
- Conventional Commits format: `type(scope): description`
- Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert
- Scopes: backend, frontend, infra, docs, api, consumer, etc.

### Branch Naming
- `feature/*` for new features
- `bugfix/*` for bug fixes
- `hotfix/*` for urgent production fixes
- `release/*` for release preparation
- `docs/*` for documentation-only changes

### Workflow Maintenance
- Quarterly review of action versions for updates
- Annual audit of secrets and permissions
- Monthly cleanup of old workflow runs and artifacts
- Dependency updates managed through Dependabot or manual PRs

## 3.7 Troubleshooting Common CI/CD Issues

### Dependency Resolution Failures
- Check `go.sum` and `package-lock.json` for consistency
- Clear caches: delete `~/.cache/go-build` and `~/Library/Caches/pnpm`
- Verify network connectivity to proxy.golang.org and registry.npmjs.org

### Test Flakiness
- Increase timeout values for integration tests
- Use deterministic test data (fixed seeds for random generators)
- Ensure proper cleanup between test runs

### Resource Exhaustion
- Monitor GitHub Actions minutes usage (free tier limits)
- Optimize test suites to reduce execution time
- Consider self-hosted runners for resource-intensive workloads

### Secret Access Issues
- Verify secret names match exactly in workflow files
- Check repository vs organization secret visibility
- Ensure proper permissions for deploying to target environments

### Terraform Apply Failures
- Check AWS/GCP service quotas and limits
- Verify IAM permissions for the service account/user
- Review error messages for specific resource creation failures

