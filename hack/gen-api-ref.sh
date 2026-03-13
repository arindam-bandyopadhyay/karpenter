#!/usr/bin/env bash
# Karpenter Provider OCI
#
# Copyright (c) 2026 Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

set -euo pipefail

# Generates the OCI v1beta1 API reference as GitHub-flavored Markdown.
#
# Requires:
#   crddoc (install via hack/toolchain.sh)
#
# Output:
#   docs/src/reference/v1beta1-api-raw.md

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_MD="${ROOT_DIR}/docs/reference/v1beta1-api-raw.md"

mkdir -p "$(dirname "${OUT_MD}")"

if ! command -v crddoc >/dev/null 2>&1; then
  echo "crddoc not found. Run: bash hack/toolchain.sh" >&2
  exit 1
fi

template_dir="${ROOT_DIR}/hack/crddoc-template"

rm -rf "${template_dir}"
crddoc export templates --folder "${template_dir}"

pushd "${ROOT_DIR}/pkg/apis/v1beta1" >/dev/null
crddoc document crds --template "${template_dir}" --output "${OUT_MD}" .
popd >/dev/null

echo "Wrote ${OUT_MD}" >&2
