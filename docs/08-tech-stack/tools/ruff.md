# NEO Bank — Ruff Best Practices

_Tools · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [tools index](./README.md)._

## 1. Setup

Ruff lints and formats Python scripts (e.g. prototype screenshot tooling) with a single binary.

```toml
# pyproject.toml
[tool.ruff]
line-length = 100
target-version = "py311"
src = ["docs/07-ui-ux-design/prototypes"]

[tool.ruff.lint]
select = ["E", "F", "I", "B", "C4", "UP", "DTZ", "PLC", "TRY"]
fixable = ["F", "I"]
unfixable = []

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
```

Rules:

- Ruff replaces `flake8`, `isort`, `pydocstyle`, and `pyupgrade` with one tool.
- Lint (`ruff check`) and format (`ruff format --check`) run in CI for any Python files.
- Import sorting (`I`) is auto-fixed; unused imports (`F401`) are errors.
- Line length is 100; double quotes for strings; spaces for indentation.

## 2. Python Scripts

The only Python in the repo is the prototype screenshot script (`docs/07-ui-ux-design/prototypes/screenshot.py`). Ruff keeps it consistent:

- Type hints on all public functions.
- `pathlib` for path operations; no `os.path`.
- F-strings for interpolation; no `%` or `.format()`.
- `if __name__ == "__main__"` guard for entry points.

## 3. CI

Ruff runs in CI on every pull request:

```yaml
- run: ruff check docs/07-ui-ux-design/prototypes/
- run: ruff format --check docs/07-ui-ux-design/prototypes/
```

Exit non-zero on any lint or format violation.

## 4. Definition of Done (Ruff)

1. `ruff check` exits with code 0 for all Python files.
2. `ruff format --check` exits with code 0.
3. New Python files pass lint and format on first commit.

## 5. Cross-references

- Screenshot script: [`../../07-ui-ux-design/prototypes/screenshot.py`](../../07-ui-ux-design/prototypes/screenshot.py).
- Screenshot generation: [`../../07-ui-ux-design/prototypes/screenshot.sh`](../../07-ui-ux-design/prototypes/screenshot.sh).
- Prototype tooling: [`../../07-ui-ux-design/prototypes/README.md`](../../07-ui-ux-design/prototypes/README.md).

---
↑ Part of the [NEO Bank tech stack decision record](../../08-tech-stack.md).
