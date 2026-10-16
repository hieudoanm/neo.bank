#!/usr/bin/env bash
#
# Scan the repository docs and write the graph artefacts to docs/.graphdown/.
#
# Usage:
#   ./scripts/graphdown/graphdown-docs.sh            # generate all outputs
#   ./scripts/graphdown/graphdown-docs.sh --check    # fail on broken links
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG_DIR="${SCRIPT_DIR}"
REPO_ROOT="$(cd "${PKG_DIR}/../.." && pwd)"
DOCS_DIR="${REPO_ROOT}/docs"

if [ -z "${PYTHONPATH:-}" ]; then
  export PYTHONPATH="${PKG_DIR}/src:${PYTHONPATH:-}"
else
  export PYTHONPATH="${PKG_DIR}/src:${PYTHONPATH}"
fi

FORMATS="json,dot,md,svg,html"

if command -v uv >/dev/null 2>&1; then
  cd "${PKG_DIR}"
  uv run graphdown scan "${DOCS_DIR}" --format "${FORMATS}" "$@"
else
  python3 -m graphdown.cli scan "${DOCS_DIR}" --format "${FORMATS}" "$@"
fi
