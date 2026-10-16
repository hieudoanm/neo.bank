# NEO Bank — Tools

Cross-cutting tooling for repository scripts (not application code): Ruff for Python lint/format.

Back to the [tech stack decision record](../../08-tech-stack.md) · [shared backend rules](../backend/README.md#2-shared-rules).

## 1. Documents

| Document         | Covers                        |
| ---------------- | ----------------------------- |
| [`ruff.md`](./ruff.md) | Python lint and format rules |

## 2. Toolchain Rules

1. **Ruff replaces** `flake8`, `isort`, `pydocstyle`, `pyupgrade` with one binary.
2. **`ruff check`** lints; **`ruff format --check`** verifies formatting; both run in CI.
3. Only Python file in scope: `docs/07-ui-ux-design/prototypes/screenshot.py`.

## 3. Definition of Done (Tools)

1. `ruff check docs/07-ui-ux-design/prototypes/` exits with code 0.
2. `ruff format --check docs/07-ui-ux-design/prototypes/` exits with code 0.

## 4. Cross-references

- Ruff practices: [`./ruff.md`](./ruff.md)
- Screenshot script: [`../../07-ui-ux-design/prototypes/screenshot.py`](../../07-ui-ux-design/prototypes/screenshot.py)

---
↑ Part of the [NEO Bank tech stack decision record](../../08-tech-stack.md).
