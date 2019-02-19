FROM golang:1.10 as builder-postgresql
RUN go get -d github.com/newrelic/nri-postgresql/... && \
    cd /go/src/github.com/newrelic/nri-postgresql && \
    make && \
    strip ./bin/nr-postgresql

FROM newrelic/infrastructure:latest
COPY --from=builder-postgresql /go/src/github.com/newrelic/nri-postgresql/bin/nr-postgresql /var/db/newrelic-infra/newrelic-integrations/bin/nr-postgresql
COPY --from=builder-postgresql /go/src/github.com/newrelic/nri-postgresql/postgresql-definition.yml /var/db/newrelic-infra/newrelic-integrations/definition.yaml
