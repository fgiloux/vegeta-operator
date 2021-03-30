FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV VEGETA_VERSION=12.8.4
ENV USER_UID=1001

LABEL \
  org.opencontainers.image.title="Vegeta" \
  org.opencontainers.image.description="Container image for the Vegeta HTTP load testing tool. Inspired by https://github.com/peter-evans/vegeta-docker" \
  org.opencontainers.image.authors="Frederic Giloux" \
  org.opencontainers.image.url="quay.io/fgiloux/vegeta" \
  org.opencontainers.image.licenses="APLv2" \
  app.tag="vegeta:$VEGETA_VERSION" \
  name="Vegeta" \
  description="Container image for the Vegeta HTTP load testing tool. Inspired by https://github.com/peter-evans/vegeta-docker" \
  summary="" \
  vendor="" \
  maintainer="Frederic Giloux" \
  release="0.1.0" \
  version="12.8.4" \
  url="https://github.com/fgiloux/vegeta.git" \
  io.k8s.description="Container image for the Vegeta HTTP load testing tool. Inspired by https://github.com/peter-evans/vegeta-docker" \
  io.k8s.display-name="Vegeta" \
  io.openshift.tags="vegeta,perftest"

COPY s3 /bin/s3

RUN set -ex \
 && microdnf install tar gzip ca-certificates \
#  && microdnf update 
 && microdnf clean all \
 && curl -s -L "https://github.com/tsenart/vegeta/releases/download/v$VEGETA_VERSION/vegeta_${VEGETA_VERSION}_linux_amd64.tar.gz" -o /tmp/vegeta.tar.gz \
 && cd bin \
 && tar xzf /tmp/vegeta.tar.gz \
 && rm /tmp/vegeta.tar.gz \
 && mkdir -p /results /opt/config/credentials \
 && chown -R ${USER_UID} /results /opt/config

VOLUME /results /opt/config

CMD [ "/bin/vegeta", "-help" ]

WORKDIR /results

USER ${USER_UID}
