# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

Unreleased section should follow [Release Toolkit](https://github.com/newrelic/release-toolkit#render-markdown-and-update-markdown)

## Unreleased

## v2.11.0 - 2023-06-02

### 🚀 Enhancements
- Upgrade Go version to 1.20

### ⛓️ Dependencies
- Updated github.com/stretchr/testify to v1.8.4 - [Changelog 🔗](https://github.com/stretchr/testify/releases/tag/v1.8.4)
- Updated github.com/jmoiron/sqlx to v1.3.5 - [Changelog 🔗](https://github.com/jmoiron/sqlx/releases/tag/v1.3.5)
- Updated github.com/lib/pq to v1.10.9 - [Changelog 🔗](https://github.com/lib/pq/releases/tag/v1.10.9)

## 2.10.5  (2022-10-05)
### Fixed
- In Tables with multiple indexes, only one was reported. Issue https://github.com/newrelic/nri-postgresql/issues/94
- When collecting metrics from multiple databases only indexes from the 1st database were reported

## 2.10.4  (2022-07-05)
### Changed
- Update Go to v1.18
- Bump dependencies
### Added
Added support for more distributions:
- RHEL(EL) 9
- Ubuntu 22.04

## 2.10.3  (2022-05-25)
### Changed
- Updated the custom query file for Postgres 13
- fix(ci/cd): removing snyk unused step
### Added
- add_postgresql_logs_example_yml

## 2.10.2 (2022-03-10)
### Changed
- Updated the custom query sample file `pg_stat_statements` query, disabling it by default.

## 2.10.1 (2022-02-08)
### Fixed
- Add cl_cancel_req to pgBouncer pool metrics.

## 2.10.0 (2021-11-16)
### Fixed
- Numeric custom metrics without metric type defined are now converted to Gauge type instead of string.

## 2.9.0 (2021-09-17)
### Added
- Adds `COLLECTION_IGNORE_DATABASE_LIST` configuration, that allows to exclude a list of database names for metrics collection.

## 2.8.0 (2021-08-27)
### Added

Moved default config.sample to [V4](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-newer-configuration-format/), added a dependency for infra-agent version 1.20.0.

Please notice that old [V3](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-standard-configuration-format/) configuration format is deprecated, but still supported.

## 2.7.2 (2021-06-17)
### Changed
- Add `db.maxconnections` metric that collects the maximum number of concurrent connections to the database server.

### Fixed
- Index size metric now calculated using `indexrelid`, instead of `indrelid`.

## 2.7.1 (2021-06-10)
### Changed
- Add ARM support.

## 2.7.0 (2021-05-10)
### Changed
- Update Go to v1.16.
- Migrate to Go Modules
- Update Infrastracture SDK to v3.6.7.
- Update other dependecies.

## 2.6.2 (2021-03-25)
### Fixed
- Semver Library was updated
- gopkg.in/yaml.v2 library has been updated due to a medium vulnerability
- Release pipeline has been moved to Github Actions

## 2.6.1 (2020-11-24)
### Fixed
- Removed ping from the database connection so it works with pgbouncer

## 2.6.0 (2020-11-05)
### Added
- Option `collect_bloat_metrics` which enables or disables the performance-intensive bloat query

## 2.5.3 (2020-09-25)
### Fixed
- Issue with converting custom query results to strings

## 2.5.0 (2020-08-26)
### Changed
- Updated the lib/pq library

## 2.4.3 (2020-07-30)
### Changed
- Allow partial failures when building collection list

## 2.4.2 (2020-07-29)
### Fixed
- Removed check for client-side certificate and key when enabling ssl. Server certificate and key are enough to create SSL connections

## 2.4.1 (2020-07-01)
### Fixed
- Stats collection for newer pgbouncer version

## 2.4.0 (2020-06-08)
### Added
- Support for custom query file

## 2.3.5 (2020-05-29)
### Fixed
- Bug causing missing tablespace metrics

## 2.3.4 (2020-01-06)
### Fixed
- Tablespace bloat collection for Postgres 12+

## 2.3.3 (2020-01-06)
### Added
- ALL setting for `collection_list`

## 2.3.1 (2020-01-06)
### Added
- Example of `custom_metrics_query`

## 2.3.0 (2020-01-03)
### Added
- `custom_metrics_query` argument to support collecting non-standard metrics

## 2.2.0 (2019-11-18)
### Changed
- Renamed the integration executable from nr-postgresql to nri-postgresql in order to be consistent with the package naming. **Important Note:** if you have any security module rules (eg. SELinux), alerts or automation that depends on the name of this binary, these will have to be updated.
-
## 2.1.4 - 2019-10-23
### Added
- Windows MSI resources

## 2.1.3 - 2019-07-30
### Added
- Lock metrics behind `collect_db_lock_metrics`

## 2.1.2 - 2019-07-23
- Removed unneeded nrjmx dependency

## 2.1.1 - 2019-06-10
### Fixed
- Segfault when collecting indexes with new collection list format

## 2.1.0 - 2019-05-23
### Added
- A collection list mode that allows collecting everything in a list of databases

## 2.0.0 - 2019-04-25
### Changes
- Prefix entity namespaces with pg-
- Update SDK
- Add identity attributes

## 1.1.0 - 2019-03-19
### Changes
- Add optional database connection param to allow collecting metrics from any database

## 1.0.4 - 2019-03-14
### Fixes
- Remove quote_ident that was causing failures on some systems

## 1.0.3 - 2019-02-11
### Fixes
- Doesn't panic on failed pgbouncer connection

## 1.0.2 - 2019-02-04
### Fixes
- Added special case for parsing Debian versions

## 1.0.1 - 2019-01-09
### Fixes
- Added special case for parsing Ubuntu versions

## 1.0.0 - 2018-11-29
### Changes
- Bumped version for GA release

## 0.2.3 - 2018-11-15
### Added
- Added host name as metadata for easier filtering

## 0.2.2 - 2018-11-14
### Fixed
- Fail gracefully if no rows are returned for a query

## 0.2.1 - 2018-10-23
### Added
- Description for `collection_list` now states it is required.

## 0.2.0 - 2018-10-23
### Changed
- Change casing of sample from PostgreSQL to Postgresql

## 0.1.2 - 2018-10-22
### Fixed
- Missing dependency

## 0.1.1 - 2018-10-22
### Fixed
- Issue in Makefile that was causing `make package` to fail.

## 0.1.0 - 2018-09-19
### Added
- Initial version: Includes Metrics and Inventory data
