# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

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
