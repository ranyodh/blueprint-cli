#!/usr/bin/env bash

set -eux -o pipefail

: "${CHART_FILE?required}"
: "${CHART_NAME?required}"
: "${CHART_MKE_NAME:="$(basename "${CHART_FILE%%.yaml}")"}"
: "${CHART_PACKAGE:="${CHART_NAME%%-crd}"}"
: "${CHART_URL:="${CHART_REPO:="https://rke2-charts.rancher.io"}/assets/${CHART_PACKAGE}/${CHART_NAME}-${CHART_VERSION:="v0.0.0"}.tgz"}"
: "${CHART_TMP:=$(mktemp --suffix .tar.gz)}"
: "${CHART_HELM_DIR:=helm}"
: "${CHART_BUNDLE_DIR:=bundle}"

cleanup() {
  exit_code=$?
  trap - EXIT INT
  rm -rf ${CHART_TMP} ${CHART_TMP/tar.gz/tar}
  exit ${exit_code}
}
trap cleanup EXIT INT

mkdir -p ${CHART_HELM_DIR}/
mkdir -p ${CHART_BUNDLE_DIR}/

curl -fsSL "${CHART_URL}" -o "${CHART_TMP}"
tar -xvf ${CHART_TMP} -C ${CHART_HELM_DIR}/

# output raw chart after rename chart to match mke naming convention
mv ${CHART_HELM_DIR}/${CHART_NAME} ${CHART_HELM_DIR}/${CHART_MKE_NAME}

# output HelmChart resource
cat <<-EOF > "${CHART_FILE}"
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: "${CHART_NAME}"
  namespace: "${CHART_NAMESPACE:="kube-system"}"
  annotations:
    helm.cattle.io/chart-url: "${CHART_URL}"
spec:
  bootstrap: ${CHART_BOOTSTRAP:=false}
  chartContent: $(base64 -w0 < "${CHART_TMP}")
EOF
