package performancemetrics

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func runSlowQueryTest(t *testing.T, query string, version uint64, expectedLength int) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(args, version, databaseName)

	query = fmt.Sprintf(query, "testdb", args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}).AddRow(
		"queryid1", "SELECT 1", "testdb", "public", 10,
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
	cp := commonparameters.SetCommonParameters(args, version, databaseName)
	expectedQuery := queries.SlowQueriesForV13AndAbove
	query := fmt.Sprintf(expectedQuery, "testdb", args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"query_id", "query_text", "database_name", "schema_name", "execution_count",
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
	cp := commonparameters.SetCommonParameters(args, version, databaseName)
	slowQueryList, _, err := getSlowRunningMetrics(conn, cp)
	assert.EqualError(t, err, commonutils.ErrUnsupportedVersion.Error())
	assert.Len(t, slowQueryList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInEligibility(t *testing.T) {
	conn, _ := connection.CreateMockSQL(t)
	argumentList := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	enabledExtensions := map[string]bool{"pg_stat_statements": false}
	cp := commonparameters.SetCommonParameters(argumentList, 13, "testdb")
	slowQueryMetricsList := PopulateSlowRunningMetrics(conn, nil, cp, enabledExtensions)
	assert.Len(t, slowQueryMetricsList, 0)
}

func TestPopulateSlowRunningMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	argumentList := args.ArgumentList{QueryMonitoringCountThreshold: 10, Hostname: "localhost", Port: "5432"}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(argumentList, 13, databaseName)
	enabledExtensions := map[string]bool{"pg_stat_statements": true}
	query := fmt.Sprintf(queries.SlowQueriesForV13AndAbove, "testdb", argumentList.QueryMonitoringCountThreshold)
	rowData := []driver.Value{
		"queryid1", "SELECT ?", "testdb", "public", int64(10),
		float64(15), float64(5), float64(2), "SELECT", "2023-01-01T00:00:00Z",
	}
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"query_id", "query_text", "database_name", "schema_name", "execution_count",
		"avg_elapsed_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp",
	}).AddRow(rowData...).AddRow(rowData...))
	expectedRows := [][]driver.Value{
		rowData, rowData,
	}
	slowQueryMetricsList := PopulateSlowRunningMetrics(conn, pgIntegration, cp, enabledExtensions)
	compareMockRowsWithSlowQueryMetrics(t, expectedRows, slowQueryMetricsList)
	assert.Len(t, slowQueryMetricsList, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func compareMockRowsWithSlowQueryMetrics(t *testing.T, expectedRows [][]driver.Value, slowQueryMetricList []datamodels.SlowRunningQueryMetrics) {
	for index := range slowQueryMetricList {
		metric := slowQueryMetricList[index]
		assert.Equal(t, expectedRows[index][0], *metric.QueryID)
		assert.Equal(t, expectedRows[index][1], *metric.QueryText)
		assert.Equal(t, expectedRows[index][2], *metric.DatabaseName)
		assert.Equal(t, expectedRows[index][3], *metric.SchemaName)
		assert.Equal(t, expectedRows[index][4], *metric.ExecutionCount)
		assert.Equal(t, expectedRows[index][5], *metric.AvgElapsedTimeMs)
		assert.Equal(t, expectedRows[index][6], *metric.AvgDiskReads)
		assert.Equal(t, expectedRows[index][7], *metric.AvgDiskWrites)
		assert.Equal(t, expectedRows[index][8], *metric.StatementType)
		assert.Equal(t, expectedRows[index][9], *metric.CollectionTimestamp)
	}
}
