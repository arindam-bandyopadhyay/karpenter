#!/usr/bin/env bash
# Karpenter Provider OCI
#
# Copyright (c) 2026 Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

set -euo pipefail

K8S_VERSION="${K8S_VERSION:="1.29.x"}"
KUBEBUILDER_ASSETS="/usr/local/kubebuilder/bin"

main() {
    tools
    kubebuilder
}

tools() {
    go install github.com/google/addlicense@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
    go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
    go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
    go install github.com/onsi/ginkgo/v2/ginkgo@latest
    go install github.com/theunrepentantgeek/crddoc@v0.0.0-20260216183532-da03417affa6

    if ! echo "$PATH" | grep -q "${GOPATH:-undefined}/bin\|$HOME/go/bin"; then
        echo "Go workspace's \"bin\" directory is not in PATH. Run 'export PATH=\"\$PATH:\${GOPATH:-\$HOME/go}/bin\"'."
    fi
}

kubebuilder() {
    sudo mkdir -p "${KUBEBUILDER_ASSETS}"
    sudo chown "${USER}" "${KUBEBUILDER_ASSETS}"
    arch=$(go env GOARCH)
    ln -sf "$(setup-envtest use -p path "${K8S_VERSION}" --arch="${arch}" --bin-dir="${KUBEBUILDER_ASSETS}")"/* "${KUBEBUILDER_ASSETS}"
    find "$KUBEBUILDER_ASSETS"

    # Install latest binaries for 1.25.x (contains CEL fix)
    if [[ "${K8S_VERSION}" = "1.25.x" ]] && [[ "$OSTYPE" == "linux"* ]]; then
        for binary in 'kube-apiserver' 'kubectl'; do
            rm $KUBEBUILDER_ASSETS/$binary
            wget -P $KUBEBUILDER_ASSETS dl.k8s.io/v1.25.16/bin/linux/"${arch}"/${binary}
            chmod +x $KUBEBUILDER_ASSETS/$binary
        done
    fi
}

main "$@"
