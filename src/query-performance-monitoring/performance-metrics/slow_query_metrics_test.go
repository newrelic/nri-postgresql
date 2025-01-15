package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/stretchr/testify/assert"
)

func TestPopulateSlowMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	validationQuery := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	expectedQuery := queries.SlowQueriesForV13AndAbove
	query := fmt.Sprintf(expectedQuery, "testdb", min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}).AddRow(
		"newrelic_value", "queryid1", "SELECT 1", "testdb", "public", 10,
		15.0, 5, 2, "SELECT", "2023-01-01T00:00:00Z",
	))

	performancemetrics.PopulateSlowRunningMetrics(conn, pgIntegration, args, databaseName, version)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateSlowMetricsInEligibility(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	validationQuery := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	slowqueryList := performancemetrics.PopulateSlowRunningMetrics(conn, pgIntegration, args, databaseName, version)

	assert.Len(t, slowqueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func runSlowQueryTest(t *testing.T, query string, version uint64, expectedLength int) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"

	query = fmt.Sprintf(query, "testdb", min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}).AddRow(
		"newrelic_value", "queryid1", "SELECT 1", "testdb", "public", 10,
		15.0, 5, 2, "SELECT", "2023-01-01T00:00:00Z",
	))
	slowQueryList, _, err := performancemetrics.GetSlowRunningMetrics(conn, args, databaseName, version)

	assert.NoError(t, err)
	assert.Len(t, slowQueryList, expectedLength)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSlowRunningMetrics(t *testing.T) {
	runSlowQueryTest(t, queries.SlowQueriesForV13AndAbove, 13, 1)
}

func TestGetSlowRunningMetricsV12(t *testing.T) {
	runSlowQueryTest(t, queries.SlowQueriesForV12, 12, 1)
}

func TestGetSlowRunningEmptyMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	expectedQuery := queries.SlowQueriesForV13AndAbove
	query := fmt.Sprintf(expectedQuery, "testdb", min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}))
	slowQueryList, _, err := performancemetrics.GetSlowRunningMetrics(conn, args, databaseName, version)

	assert.NoError(t, err)
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSlowRunningMetricsUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(11)
	slowQueryList, _, err := performancemetrics.GetSlowRunningMetrics(conn, args, databaseName, version)
	assert.EqualError(t, err, commonutils.ErrUnsupportedVersion.Error())
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
