package performancemetrics

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetWaitEventMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	cp := common_parameters.SetCommonParameters(args, uint64(14), databaseName)

	var query = fmt.Sprintf(queries.WaitEvents, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 1000.0, "2023-01-01T00:00:00Z", "queryid1", "SELECT 1", "testdb",
	))
	waitEventsList, err := getWaitEventMetrics(conn, cp)

	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWaitEventEmptyMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	cp := common_parameters.SetCommonParameters(args, uint64(14), databaseName)

	var query = fmt.Sprintf(queries.WaitEvents, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}))
	waitEventsList, err := getWaitEventMetrics(conn, cp)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWaitEventMetricsFromPgStatActivity(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10, Hostname: "testhost.rds.amazonaws.com"}
	databaseName := "testdb"

	cp := common_parameters.SetCommonParameters(args, uint64(14), databaseName)
	query := fmt.Sprintf(queries.WaitEventsFromPgStatActivity, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 500.0, "2023-01-01T00:00:00Z", "queryid2", "SELECT 2", "testdb",
	))
	waitEventsList, err := getWaitEventMetricsPgStat(conn, cp)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 1)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWaitEventMetricsPgStat_Success(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	cp := &common_parameters.CommonParameters{
		Databases:                     "testdb",
		QueryMonitoringCountThreshold: 10,
	}
	query := fmt.Sprintf(queries.WaitEventsFromPgStatActivity, cp.Databases, cp.QueryMonitoringCountThreshold)
	mockRows := sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 500.0, "2023-01-01T00:00:00Z", "queryid1", "SELECT 1", "testdb",
	)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(mockRows)

	waitEventMetrics, err := getWaitEventMetricsPgStat(conn, cp)

	assert.NoError(t, err)
	assert.Len(t, waitEventMetrics, 1)
	assert.Equal(t, "Locks:Lock", *waitEventMetrics[0].WaitEventName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFilteredWaitEvents(t *testing.T) {
	waitEventMetrics := []datamodels.WaitEventMetrics{
		{
			QueryText: stringPointer("SELECT 1"),
		},
	}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{
			QueryText: stringPointer("SELECT ?"),
			QueryID:   stringPointer("queryid1"),
		},
	}

	filteredMetrics := getFilteredWaitEvents(waitEventMetrics, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 1)
	filteredMetric := filteredMetrics[0].(datamodels.WaitEventMetrics)
	assert.Equal(t, "SELECT ?", *filteredMetric.QueryText)
	assert.Equal(t, "queryid1", *filteredMetric.QueryID)
}

func TestGetFilteredWaitEvents_NoMatch(t *testing.T) {
	waitEventMetrics := []datamodels.WaitEventMetrics{
		{
			QueryText: stringPointer("SELECT 2"),
		},
	}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{
			QueryText: stringPointer("SELECT 1 where a='b'"),
			QueryID:   stringPointer("queryid1"),
		},
	}

	filteredMetrics := getFilteredWaitEvents(waitEventMetrics, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 0)
}
