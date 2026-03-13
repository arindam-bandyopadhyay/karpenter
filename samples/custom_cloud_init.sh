#!/usr/bin/env bash
# Karpenter Provider OCI
#
# Copyright (c) 2026 Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/


set -o errexit
set -o nounset
set -o pipefail

MD_URL="http://169.254.169.254/opc/v2/instance/metadata"
AUTH_HDR="Authorization: Bearer Oracle"

# Fetch a metadata key, returning empty on error/missing
fetch_md() {
  local key="$1"
  curl -sfL -H "${AUTH_HDR}" --connect-timeout 2 --max-time 5 "${MD_URL}/${key}" 2>/dev/null || true
}

CLUSTER_DNS="$(fetch_md kubedns_svc_ip)"
KUBELET_EXTRA_ARGS="$(fetch_md kubelet-extra-args)"
APISERVER_ENDPOINT="$(fetch_md apiserver_host)"
KUBELET_CA_CERT="$(fetch_md cluster_ca_cert)"

# Export only when present to avoid surprising consumers with empty values
[ -n "${CLUSTER_DNS}" ] && export CLUSTER_DNS
[ -n "${KUBELET_EXTRA_ARGS}" ] && export KUBELET_EXTRA_ARGS
[ -n "${APISERVER_ENDPOINT}" ] && export APISERVER_ENDPOINT
[ -n "${KUBELET_CA_CERT}" ] && export KUBELET_CA_CERT

# BEGIN OF CUSTOM SCRIPT BOOTSTRAP SCRIPT , REPLACE THIS SECTION WITH CUSTOM PRE BOOTSTRAP SCRIPT
#echo "pre bootstrap script"
#echo "CLUSTER_DNS: ${CLUSTER_DNS:-}"
#echo "KUBELET_EXTRA_ARGS: ${KUBELET_EXTRA_ARGS:-}"
#echo "APISERVER_ENDPOINT: ${APISERVER_ENDPOINT:-}"
#echo "KUBELET_CA_CERT: ${KUBELET_CA_CERT:-}"
# END OF CUSTOM SCRIPT BOOTSTRAP SCRIPT

bash /etc/oke/oke-install.sh

# BEGIN OF POST BOOTSTRAP SCRIPT, IF NEEDED
#echo "post bootstrap script"
#END OF POST BOOTSTRAP SCRIPT
