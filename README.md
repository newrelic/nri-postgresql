# New Relic Infrastructure Integration for PostgreSQL


The New Relic Infrastructure Integration for PostgreSQL captures critical performance metrics and inventory reported by PostgreSQL instances. Data on the level of mongos, mongod, config server, database, and collection is collected. Additionally, the integration can be configured to collect metrics on PgBouncer.

Inventory data for the configuration of each mongod, mongos, and config server is collected.

## Requirements

A user with the necessary permissions must be present on the cluster and each mongod for all metrics to be collected. See the documentation for details.

## Installation

- download an archive file for the PostgreSQL Integration
- extract `postgresql-definition.yml` and `/bin` directory into `/var/db/newrelic-infra/newrelic-integrations`
- add execute permissions for the binary file `nr-postgresql` (if required)
- extract `postgresql-config.yml.sample` into `/etc/newrelic-infra/integrations.d`

## Usage

This is the description about how to run the PostgreSQL Integration with New Relic Infrastructure agent, so it is required to have the agent installed (see [agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

In order to use the PostgreSQL Integration it is required to configure `postgresql-config.yml.sample` file. Firstly, rename the file to `postgresql-config.yml`. Then, depending on your needs, specify all instances that you want to monitor. Once this is done, restart the Infrastructure agent.

You can view your data in Insights by creating your own custom NRQL queries. To do so, use the **PostgreSQLInstanceSample**, **PostgresSQLDatabaseSample**, **PostgreSQLTableSample**,**PostgreSQLIndexSample**, and **PgBouncerSample** event types.

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
