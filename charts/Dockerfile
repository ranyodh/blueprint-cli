FROM rancher/hardened-build-base:v1.20.5b2 AS build
RUN set -x && \
    apk --no-cache add \
    bash \
    curl \
    file \
    git \
    libseccomp-dev \
    rsync \
    gcc \
    bsd-compat-headers \
    py-pip \
    py3-pip \
    pigz \
    tar \
    yq

FROM build AS charts
ARG CHART_REPO="https://rke2-charts.rancher.io"
ARG CACHEBUST="cachebust"
ARG CHART_BUNDLE_DIR=/bundle
ARG CHART_HELM_DIR=/helm
COPY build-chart.sh /build-chart.sh
RUN echo ${CACHEBUST}>/dev/null
RUN CHART_VERSION="1.13.200"   CHART_NAME=rke2-cilium              CHART_FILE=/bundle/mke-cilium.yaml         /build-chart.sh
RUN CHART_VERSION="v3.26.100"  CHART_NAME=rke2-calico-crd          CHART_FILE=/bundle/mke-calico-crd.yaml     /build-chart.sh
RUN CHART_VERSION="v3.26.100"  CHART_NAME=rke2-calico              CHART_FILE=/bundle/mke-calico.yaml         /build-chart.sh
RUN CHART_VERSION="4.6.100"    CHART_NAME=rke2-ingress-nginx        CHART_FILE=/bundle/mke-ingress-nginx.yaml  /build-chart.sh
RUN CHART_URL="https://github.com/Kong/charts/releases/download/ingress-0.4.0/ingress-0.4.0.tgz"  \
    CHART_NAME=ingress CHART_FILE=/bundle/mke-kong.yaml /build-chart.sh
RUN rm -vf *.sh *.md

FROM scratch AS charts-bundle
COPY --from=charts /bundle/ /

FROM scratch AS charts-helm
COPY --from=charts /helm/ /

