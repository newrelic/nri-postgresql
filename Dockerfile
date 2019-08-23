FROM golang:1.10 as builder-postgresql
RUN go get -d github.com/newrelic/nri-postgresql/... && \
    cd /go/src/github.com/newrelic/nri-postgresql && \
    make && \
    strip ./bin/nr-postgresql

FROM newrelic/infrastructure:latest
ENV NRIA_IS_FORWARD_ONLY true
ENV NRIA_K8S_INTEGRATION true
COPY --from=builder-postgresql /go/src/github.com/newrelic/nri-postgresql/bin/nr-postgresql /nri-sidecar/newrelic-infra/newrelic-integrations/bin/nr-postgresql
COPY --from=builder-postgresql /go/src/github.com/newrelic/nri-postgresql/postgresql-definition.yml /nri-sidecar/newrelic-infra/newrelic-integrations/definition.yaml
USER 1000
