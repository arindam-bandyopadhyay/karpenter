#!/usr/bin/env bash
# Karpenter Provider OCI
#
# Copyright (c) 2026 Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_MD="${ROOT_DIR}/docs/reference/helm-chart.md"

mkdir -p "$(dirname "${OUT_MD}")"

if ! command -v helm-docs >/dev/null 2>&1; then
  echo "helm-docs not found. Run: bash hack/toolchain.sh" >&2
  exit 1
fi

pushd "${ROOT_DIR}" >/dev/null
helm-docs \
  --chart-search-root . \
  --chart-to-generate chart \
  --output-file ../docs/reference/helm-chart.md
popd >/dev/null

echo "Wrote ${OUT_MD}" >&2
