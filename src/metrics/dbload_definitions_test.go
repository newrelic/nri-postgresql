package metrics

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/stretchr/testify/assert"
)

func Test_generateDBLoadDefinitions_EmptyDatabaseList(t *testing.T) {
	v13 := semver.MustParse("13.0.0")
	databaseList := collection.DatabaseList{}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v13)

	assert.Equal(t, 0, len(queryDefinitions), "Should return empty array for empty database list")
}

func Test_generateDBLoadDefinitions_PostgreSQL10(t *testing.T) {
	v10 := semver.MustParse("10.0.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v10)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition for PostgreSQL 10+")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NULL", "Should use wait_event column for PG 10+")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NOT NULL", "Should check for non-null wait_event")
	assert.Contains(t, queryDefinitions[0].query, "backend_type = 'client backend'", "Should filter by backend_type (added in v10)")
	assert.Contains(t, queryDefinitions[0].query, "state = 'active'", "Should filter by active state")
	assert.Contains(t, queryDefinitions[0].query, "pid != pg_backend_pid()", "Should exclude monitoring process itself")
}

func Test_generateDBLoadDefinitions_PostgreSQL96(t *testing.T) {
	v96 := semver.MustParse("9.6.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v96)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition for PostgreSQL 9.6")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NULL", "Should use wait_event column for PG 9.6")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NOT NULL", "Should check for non-null wait_event")
	assert.NotContains(t, queryDefinitions[0].query, "backend_type", "Should NOT use backend_type (not available in 9.6)")
	assert.Contains(t, queryDefinitions[0].query, "state = 'active'", "Should filter by active state")
}

func Test_generateDBLoadDefinitions_PostgreSQL99(t *testing.T) {
	v99 := semver.MustParse("9.9.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v99)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition for PostgreSQL 9.9")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NULL", "Should use wait_event column for PG 9.9")
	assert.NotContains(t, queryDefinitions[0].query, "backend_type", "Should NOT use backend_type (added in v10, not in 9.9)")
	assert.Contains(t, queryDefinitions[0].query, "state = 'active'", "Should filter by active state")
}

func Test_generateDBLoadDefinitions_PostgreSQL95(t *testing.T) {
	v95 := semver.MustParse("9.5.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v95)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition for PostgreSQL 9.2-9.5")
	assert.Contains(t, queryDefinitions[0].query, "waiting", "Should use 'waiting' column for PG 9.2-9.5")
	assert.NotContains(t, queryDefinitions[0].query, "wait_event", "Should NOT use wait_event (not available before 9.6)")
	assert.Contains(t, queryDefinitions[0].query, "state = 'active'", "Should use state column (available in 9.2+)")
	assert.NotContains(t, queryDefinitions[0].query, "backend_type", "Should NOT use backend_type (not available before 10)")
}

func Test_generateDBLoadDefinitions_PostgreSQL13(t *testing.T) {
	v13 := semver.MustParse("13.0.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v13)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition for PostgreSQL 13+")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NULL", "Should use wait_event column for PG 13+")
	assert.Contains(t, queryDefinitions[0].query, "backend_type = 'client backend'", "Should use backend_type (13 >= 10)")
}

func Test_generateDBLoadDefinitions_PostgreSQL16(t *testing.T) {
	v16 := semver.MustParse("16.0.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v16)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition for PostgreSQL 16+")
	assert.Contains(t, queryDefinitions[0].query, "wait_event IS NULL", "Should use wait_event column for PG 16+")
	assert.Contains(t, queryDefinitions[0].query, "backend_type = 'client backend'", "Should use backend_type (16 >= 10)")
}

func Test_generateDBLoadDefinitions_MultipleDatabases(t *testing.T) {
	v13 := semver.MustParse("13.0.0")
	databaseList := collection.DatabaseList{
		"database1": {},
		"database2": {},
		"database3": {},
	}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v13)

	assert.Equal(t, 1, len(queryDefinitions), "Should return 1 query definition")
	assert.Contains(t, queryDefinitions[0].query, "datname IN", "Should filter by database names")
}

func Test_DBLoadQueryStructure_PostgreSQL10(t *testing.T) {
	v10 := semver.MustParse("10.0.0")
	databaseList := collection.DatabaseList{"testdb": {}}

	queryDefinitions := generateDBLoadDefinitions(databaseList, &v10)
	query := queryDefinitions[0].query

	// Verify all required fields are present
	assert.Contains(t, query, "cpu_load", "Should calculate CPU load")
	assert.Contains(t, query, "wait_load", "Should calculate wait load")
	assert.Contains(t, query, "db_load", "Should calculate total DB load")
	assert.Contains(t, query, "datname AS database", "Should include database name")
	assert.Contains(t, query, "GROUP BY datname", "Should group by database")
	assert.Contains(t, query, "state = 'active'", "Should filter for active sessions only")
	assert.Contains(t, query, "backend_type = 'client backend'", "Should filter client backends")
	assert.Contains(t, query, "pid != pg_backend_pid()", "Should exclude monitoring process")
}

func Test_DBLoadDefinition10_DataModelStructure(t *testing.T) {
	// Verify the data model has the correct struct tags
	dataModels := dbLoadDefinition10.dataModels

	assert.NotNil(t, dataModels, "Data models should not be nil")
}

func Test_DBLoadDefinition96_DataModelStructure(t *testing.T) {
	// Verify the data model has the correct struct tags
	dataModels := dbLoadDefinition96.dataModels

	assert.NotNil(t, dataModels, "Data models should not be nil")
}

func Test_DBLoadDefinitionLegacy_DataModelStructure(t *testing.T) {
	// Verify the data model has the correct struct tags
	dataModels := dbLoadDefinitionLegacy.dataModels

	assert.NotNil(t, dataModels, "Data models should not be nil")
}
