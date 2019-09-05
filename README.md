# New Relic Infrastructure Integration for PostgreSQL


The New Relic Infrastructure Integration for PostgreSQL captures critical performance metrics and inventory reported by PostgreSQL instances. Data on the level of instance, database, and collection is collected. Additionally, the integration can be configured to collect metrics on PgBouncer.

Inventory data for the configuration of the instance is collected from the `pg_statistics` database.

See our [documentation web site](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/postgresql-monitoring-integration) for more details.

## Requirements

A user with the necessary permissions must be present on the database for all metrics to be collected. See the documentation for details on permissions.

## Installation

- download an archive file for the PostgreSQL Integration
- extract `postgresql-definition.yml` and `/bin` directory into `/var/db/newrelic-infra/newrelic-integrations`
- add execute permissions for the binary file `nr-postgresql` (if required)
- extract `postgresql-config.yml.sample` into `/etc/newrelic-infra/integrations.d`

## Usage

This is the description about how to run the PostgreSQL Integration with New Relic Infrastructure agent, so it is required to have the agent installed (see [agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

In order to use the PostgreSQL Integration it is required to configure `postgresql-config.yml.sample` file. Firstly, rename the file to `postgresql-config.yml`. Then, depending on your needs, specify all instances that you want to monitor. Once this is done, restart the Infrastructure agent.

You can view your data in Insights by creating your own custom NRQL queries. To do so, use the **PostgresqlInstanceSample**, **PostgressqlDatabaseSample**, **PostgresqlTableSample**,**PostgresqlIndexSample**, and **PgBouncerSample** event types.

### Database Lock Metrics

Collecting DB Lock Metrics requires that you first install the `tablefunc` extension on the `public` schema of the database you will be collecting lock metrics for. You can do so by:

1. Installing the postgresql contribs package for your particular OS; and then
2. Run the query `CREATE EXTENSION tablefunc;` against your database's public schema

Afterwards, simply enable db lock collection by setting `collect_db_lock_metrics: true` in your nri-postgresql config file.

## Compatibility

* Supported OS: No limitations
* Supported PostgreSQL version: 9.0+

## Integration Development usage

Assuming that you have source code you can build and run the PostgreSQL Integration locally.

* Go to directory of the PostgreSQL Integration and build it
```bash
$ make
```
* The command above will execute tests for the PostgreSQL Integration and build an executable file called `nr-postgresql` in `bin` directory.
```bash
$ ./bin/nr-postgresql
```
* If you want to know more about usage of `./nr-postgresql` check
```bash
$ ./bin/nr-postgresql --help
```

For managing external dependencies [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to specific version (if possible) into vendor directory.
