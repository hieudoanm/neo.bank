# NEO Bank — Test Maintenance

## Purpose

Defines how tests are reviewed, maintained, and improved over time as the NEO Bank codebase evolves.

## 1. Regular Review

### Monthly Review Cycle

| Activity | Frequency | Owner |
|----------|-----------|-------|
| Coverage review | Monthly | Tech lead |
| Flaky test audit | Weekly | On-call engineer |
| Obsolete test removal | Monthly | Team |
| Fixture updates | As data model changes | Feature developer |
| Test debt triage | Sprint planning | Team |

### What to Review

- **Coverage trends** — Is coverage increasing or decreasing?
- **Test duration** — Are tests getting slower?
- **Flaky tests** — Which tests fail intermittently?
- **Obsolete tests** — Tests for removed features or changed behavior.
- **Missing tests** — New features without test coverage.
- **Fixture accuracy** — Do fixtures match the current data model?

## 2. Test Debt

Test debt is tracked alongside feature debt:

| Debt Type | Example | Priority |
|-----------|---------|----------|
| Missing unit test | New service method without tests | High |
| Flaky integration test | Cassandra timing-dependent test | High |
| Obsolete test | Test for removed API endpoint | Medium |
| Slow E2E test | Test taking > 2 minutes | Medium |
| Missing negative test | Authorisation not tested negatively | High |
| Outdated fixture | Fixture doesn't match new schema | Medium |

### Tracking

- Test debt items are added to the project board.
- Test debt is prioritized alongside feature debt.
- Time is allocated for test improvement in each sprint (recommended: 10-15%).
- Debt items have an owner and a due date.

## 3. When to Add Tests

Tests are added when:

- A new feature is implemented (unit + integration + E2E for critical paths).
- A bug is fixed (regression test to prevent recurrence).
- A security vulnerability is found (security test).
- A performance issue is identified (performance test).
- The data model changes (migration test + fixture update).
- An edge case is discovered in production.

## 4. When to Remove or Update Tests

Tests are removed or updated when:

- The feature they test is removed.
- The API contract changes (request/response format).
- The data model changes (schema migration).
- A test is proven flaky and cannot be fixed quickly.
- A test duplicates another test's coverage.

### Removal Process

1. Identify obsolete test.
2. Verify no other test depends on it.
3. Remove the test and any associated fixtures.
4. Update coverage expectations if needed.
5. Document the removal in the PR.

## 5. Test Review Checklist

Before merging a PR with test changes:

- [ ] New features have unit tests.
- [ ] New features have integration tests (if they touch Cassandra/Kafka).
- [ ] Critical user journeys have E2E tests.
- [ ] Negative tests exist for authorisation boundaries.
- [ ] Idempotency is tested for mutating operations.
- [ ] Fixtures are synthetic (no real data).
- [ ] Tests are deterministic (no random values).
- [ ] Coverage does not decrease.
- [ ] No flaky tests introduced.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Quality gates: [`04-quality-gates.md`](./04-quality-gates.md)
- Definition of Done: [`11-definition-of-done.md`](./11-definition-of-done.md)
- Delivery rules: `AGENTS.md`
- Planned test: `test_maintenance` (verifies test debt is tracked)