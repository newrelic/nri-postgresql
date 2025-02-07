package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
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
	waitEventsList, err := performancemetrics.GetWaitEventMetrics(conn, cp)

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
	waitEventsList, err := performancemetrics.GetWaitEventMetrics(conn, cp)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
