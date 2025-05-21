package performancemetrics

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func runSlowQueryTest(t *testing.T, query string, version uint64, expectedLength int) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	cp := common_parameters.SetCommonParameters(args, version, databaseName)

	query = fmt.Sprintf(query, "testdb", args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}).AddRow(
		"newrelic_value", "queryid1", "SELECT 1", "testdb", "public", 10,
		15.0, 5, 2, "SELECT", "2023-01-01T00:00:00Z",
	))
	slowQueryList, _, err := getSlowRunningMetrics(conn, cp)
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
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	expectedQuery := queries.SlowQueriesForV13AndAbove
	query := fmt.Sprintf(expectedQuery, "testdb", args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}))
	slowQueryList, _, err := getSlowRunningMetrics(conn, cp)

	assert.NoError(t, err)
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSlowRunningMetricsUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(11)
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	slowQueryList, _, err := getSlowRunningMetrics(conn, cp)
	assert.EqualError(t, err, commonutils.ErrUnsupportedVersion.Error())
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFilteredIndividualAndSlowMetrics_MatchingQueries(t *testing.T) {
	individualQueries := []string{"SELECT * FROM users", "SELECT * FROM orders"}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{QueryText: stringPointer("SELECT * FROM users")},
		{QueryText: stringPointer("SELECT * FROM products")},
	}

	filteredMetrics, _ := getFilteredSlowMetrics(individualQueries, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 1)
	assert.Equal(t, "SELECT * FROM users", *filteredMetrics[0].QueryText)
}

func TestGetFilteredIndividualAndSlowMetrics_NoMatchingQueries(t *testing.T) {
	individualQueries := []string{"SELECT * FROM customers"}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{QueryText: stringPointer("SELECT * FROM users")},
		{QueryText: stringPointer("SELECT * FROM products")},
	}

	filteredMetrics, filteredMetricsInterface := getFilteredSlowMetrics(individualQueries, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 0)
	assert.Len(t, filteredMetricsInterface, 0)
}

func TestGetFilteredIndividualAndSlowMetrics_EmptyInputs(t *testing.T) {
	individualQueries := []string{}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{}

	filteredMetrics, filteredMetricsInterface := getFilteredSlowMetrics(individualQueries, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 0)
	assert.Len(t, filteredMetricsInterface, 0)
}

func TestGetFilteredIndividualAndSlowMetrics_DuplicateQueries(t *testing.T) {
	individualQueries := []string{"SELECT * FROM users", "SELECT * FROM users"}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{QueryText: stringPointer("SELECT * FROM users")},
	}

	filteredMetrics, filteredMetricsInterface := getFilteredSlowMetrics(individualQueries, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 1)
	assert.Len(t, filteredMetricsInterface, 1)
	assert.Equal(t, "SELECT * FROM users", *filteredMetrics[0].QueryText)
}

func stringPointer(s string) *string {
	return &s
}
