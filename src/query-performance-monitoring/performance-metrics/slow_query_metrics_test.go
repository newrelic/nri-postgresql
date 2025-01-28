package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func runSlowQueryTest(t *testing.T, query string, version uint64, expectedLength int) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	gv := global_variables.SetGlobalVariables(args, version, databaseName)

	query = fmt.Sprintf(query, "testdb", min(args.QueryCountThreshold, commonutils.MaxQueryCountThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}).AddRow(
		"newrelic_value", "queryid1", "SELECT 1", "testdb", "public", 10,
		15.0, 5, 2, "SELECT", "2023-01-01T00:00:00Z",
	))
	slowQueryList, _, err := performancemetrics.GetSlowRunningMetrics(conn, gv)
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
	gv := global_variables.SetGlobalVariables(args, version, databaseName)
	expectedQuery := queries.SlowQueriesForV13AndAbove
	query := fmt.Sprintf(expectedQuery, "testdb", min(args.QueryCountThreshold, commonutils.MaxQueryCountThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}))
	slowQueryList, _, err := performancemetrics.GetSlowRunningMetrics(conn, gv)

	assert.NoError(t, err)
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSlowRunningMetricsUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(11)
	gv := global_variables.SetGlobalVariables(args, version, databaseName)
	slowQueryList, _, err := performancemetrics.GetSlowRunningMetrics(conn, gv)
	assert.EqualError(t, err, commonutils.ErrUnsupportedVersion.Error())
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
