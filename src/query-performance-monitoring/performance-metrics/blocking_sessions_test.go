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

func TestGetBlockingMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	cp := common_parameters.SetCommonParameters(args, version, databaseName)

	expectedQuery := queries.BlockingQueriesForV12AndV13
	query := fmt.Sprintf(expectedQuery, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name",
		"blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start",
	}).AddRow(
		"newrelic_value", 123, "SELECT 1", 1233444, "2023-01-01 00:00:00", "testdb",
		456, "SELECT 2", 4566, "2023-01-01 00:00:00",
	))
	blockingQueriesMetricsList, err := performancemetrics.GetBlockingMetrics(conn, cp)
	assert.NoError(t, err)
	assert.Len(t, blockingQueriesMetricsList, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
