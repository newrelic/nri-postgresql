package metrics

import (
	"github.com/blang/semver/v4"
	"github.com/newrelic/nri-postgresql/src/collection"
)

// generateDBLoadDefinitions returns version-specific DBLoad query definitions
func generateDBLoadDefinitions(databases collection.DatabaseList, version *semver.Version) []*QueryDefinition {
	queryDefinitions := make([]*QueryDefinition, 0, 1)
	if len(databases) == 0 {
		return queryDefinitions
	}

	// PostgreSQL 10+ added backend_type column
	// PostgreSQL 9.6 added wait_event columns (replaced 'waiting' boolean)
	// PostgreSQL 9.2 added state column and renamed current_query to query
	v10 := semver.MustParse("10.0.0")
	v96 := semver.MustParse("9.6.0")

	if version.GE(v10) {
		// PostgreSQL 10+ (has backend_type, wait_event, and state)
		queryDefinitions = append(queryDefinitions, dbLoadDefinition10.insertDatabaseNames(databases))
	} else if version.GE(v96) {
		// PostgreSQL 9.6-9.x (has wait_event and state, but NOT backend_type)
		queryDefinitions = append(queryDefinitions, dbLoadDefinition96.insertDatabaseNames(databases))
	} else {
		// PostgreSQL 9.2-9.5 (has state column and 'waiting' boolean, but NOT wait_event)
		queryDefinitions = append(queryDefinitions, dbLoadDefinitionLegacy.insertDatabaseNames(databases))
	}

	return queryDefinitions
}

// dbLoadDefinition10 is for PostgreSQL 10+ with backend_type column
// This is the most accurate query using backend_type to filter client sessions
var dbLoadDefinition10 = &QueryDefinition{
	query: `SELECT
		datname AS database,
		COUNT(*) FILTER (WHERE state = 'active' AND wait_event IS NULL) AS cpu_load,
		COUNT(*) FILTER (WHERE state = 'active' AND wait_event IS NOT NULL) AS wait_load,
		COUNT(*) FILTER (WHERE state = 'active') AS db_load
	FROM pg_stat_activity
	WHERE backend_type = 'client backend'
		AND datname IN (%DATABASES%)
		AND pid != pg_backend_pid()
	GROUP BY datname;`,

	dataModels: []struct {
		databaseBase
		CPULoad  *float64 `db:"cpu_load"  metric_name:"db.load.cpu"  source_type:"gauge"`
		WaitLoad *float64 `db:"wait_load" metric_name:"db.load.wait" source_type:"gauge"`
		DBLoad   *float64 `db:"db_load"   metric_name:"db.load"      source_type:"gauge"`
	}{},
}

// dbLoadDefinition96 is for PostgreSQL 9.6-9.x which has wait_event but NOT backend_type
// Uses state='active' to filter active sessions (excludes idle and background processes)
var dbLoadDefinition96 = &QueryDefinition{
	query: `SELECT
		datname AS database,
		COUNT(*) FILTER (WHERE state = 'active' AND wait_event IS NULL) AS cpu_load,
		COUNT(*) FILTER (WHERE state = 'active' AND wait_event IS NOT NULL) AS wait_load,
		COUNT(*) FILTER (WHERE state = 'active') AS db_load
	FROM pg_stat_activity
	WHERE datname IN (%DATABASES%)
		AND pid != pg_backend_pid()
	GROUP BY datname;`,

	dataModels: []struct {
		databaseBase
		CPULoad  *float64 `db:"cpu_load"  metric_name:"db.load.cpu"  source_type:"gauge"`
		WaitLoad *float64 `db:"wait_load" metric_name:"db.load.wait" source_type:"gauge"`
		DBLoad   *float64 `db:"db_load"   metric_name:"db.load"      source_type:"gauge"`
	}{},
}

// dbLoadDefinitionLegacy is for PostgreSQL 9.2-9.5 which uses 'waiting' boolean and 'query' column
// Uses state='active' to filter active sessions (newer 9.2+ approach)
var dbLoadDefinitionLegacy = &QueryDefinition{
	query: `SELECT
		datname AS database,
		COUNT(*) FILTER (WHERE state = 'active' AND NOT waiting) AS cpu_load,
		COUNT(*) FILTER (WHERE state = 'active' AND waiting) AS wait_load,
		COUNT(*) FILTER (WHERE state = 'active') AS db_load
	FROM pg_stat_activity
	WHERE datname IN (%DATABASES%)
		AND pid != pg_backend_pid()
	GROUP BY datname;`,

	dataModels: []struct {
		databaseBase
		CPULoad  *float64 `db:"cpu_load"  metric_name:"db.load.cpu"  source_type:"gauge"`
		WaitLoad *float64 `db:"wait_load" metric_name:"db.load.wait" source_type:"gauge"`
		DBLoad   *float64 `db:"db_load"   metric_name:"db.load"      source_type:"gauge"`
	}{},
}
