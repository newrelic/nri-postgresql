package performancemetrics

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
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
	rowData := []driver.Value{
		"newrelic_value", int64(123), "SELECT 1", "1233444", "2023-01-01 00:00:00", "testdb",
		int64(456), "SELECT 2", "4566", "2023-01-01 00:00:00",
	}
	expectedRows := [][]driver.Value{
		rowData, rowData,
	}
	mockRows := sqlmock.NewRows([]string{
		"newrelic", "blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name",
		"blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start",
	}).AddRow(rowData...).AddRow(rowData...)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(mockRows)
	blockingQueriesMetricsList, err := getBlockingMetrics(conn, cp)
	compareMockRowsWithMetrics(t, expectedRows, blockingQueriesMetricsList)
	assert.NoError(t, err)
	assert.Len(t, blockingQueriesMetricsList, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func compareMockRowsWithMetrics(t *testing.T, expectedRows [][]driver.Value, blockingQueriesMetricsList []interface{}) {
	assert.Equal(t, 2, len(blockingQueriesMetricsList))
	for index := range blockingQueriesMetricsList {
		anonymizeQuery := commonutils.AnonymizeQueryText(expectedRows[index][2].(string))
		blockingSession := blockingQueriesMetricsList[index].(datamodels.BlockingSessionMetrics)
		assert.Equal(t, expectedRows[index][0], *blockingSession.Newrelic)
		assert.Equal(t, expectedRows[index][1], *blockingSession.BlockedPid)
		assert.Equal(t, anonymizeQuery, *blockingSession.BlockedQuery)
		assert.Equal(t, expectedRows[index][3], *blockingSession.BlockedQueryID)
		assert.Equal(t, expectedRows[index][4], *blockingSession.BlockedQueryStart)
		assert.Equal(t, expectedRows[index][5], *blockingSession.BlockedDatabase)
		assert.Equal(t, expectedRows[index][6], *blockingSession.BlockingPid)
		assert.Equal(t, anonymizeQuery, *blockingSession.BlockingQuery)
		assert.Equal(t, expectedRows[index][8], *blockingSession.BlockingQueryID)
		assert.Equal(t, expectedRows[index][9], *blockingSession.BlockingQueryStart)
	}
}

func TestGetBlockingMetricsErr(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	_, err := getBlockingMetrics(conn, cp)
	assert.EqualError(t, err, commonutils.ErrUnExpectedError.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockingMetricsPgStat_Success(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	cp := &common_parameters.CommonParameters{
		Databases:                     "testdb",
		QueryMonitoringCountThreshold: 10,
		Version:                       14,
	}
	query := fmt.Sprintf(queries.RDSPostgresBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	mockRows := sqlmock.NewRows([]string{
		"newrelic", "blocked_pid", "blocked_query", "blocked_query_start", "database_name",
		"blocking_pid", "blocking_query", "blocking_query_start",
	}).AddRow(
		"newrelic_value", 123, "SELECT 1", "2023-01-01 00:00:00", "testdb",
		456, "SELECT 2", "2023-01-01 00:00:00",
	).AddRow(
		"newrelic_value", 789, "SELECT 3", "2023-01-02 00:00:00", "testdb",
		101, "SELECT 4", "2023-01-02 00:00:00",
	)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(mockRows)

	blockingMetrics, err := getBlockingMetricsPgStat(conn, cp)

	assert.NoError(t, err)
	assert.Len(t, blockingMetrics, 2)
	assert.Equal(t, "SELECT 1", *blockingMetrics[0].BlockedQuery)
	assert.Equal(t, "SELECT 2", *blockingMetrics[0].BlockingQuery)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockingMetricsPgStat_Error(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	cp := &common_parameters.CommonParameters{
		Databases:                     "testdb",
		QueryMonitoringCountThreshold: 10,
		Version:                       14,
	}
	query := fmt.Sprintf(queries.RDSPostgresBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(commonutils.ErrUnExpectedError)

	blockingMetrics, err := getBlockingMetricsPgStat(conn, cp)

	assert.Error(t, err)
	assert.Nil(t, blockingMetrics)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFilteredBlockingSessions(t *testing.T) {
	blockingSessionMetrics := []datamodels.BlockingSessionMetrics{
		{
			BlockingQuery: stringPointer("SELECT 1"),
			BlockedQuery:  stringPointer("SELECT 2"),
		},
	}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{
			QueryText: stringPointer("SELECT 1"),
			QueryID:   stringPointer("123"),
		},
	}

	filteredMetrics := getFilteredBlockingSessions(blockingSessionMetrics, slowQueryMetrics)

	assert.Len(t, filteredMetrics, 1)
}

func TestGetFilteredBlockingSessions_NoMatch(t *testing.T) {
	blockingSessionMetrics := []datamodels.BlockingSessionMetrics{
		{
			BlockingQuery: stringPointer("SELECT 3"),
			BlockedQuery:  stringPointer("SELECT 4"),
		},
	}
	slowQueryMetrics := []datamodels.SlowRunningQueryMetrics{
		{
			QueryText: stringPointer("SELECT 1 where a='b'"),
			QueryID:   stringPointer("123"),
		},
	}
	filteredMetrics := getFilteredBlockingSessions(blockingSessionMetrics, slowQueryMetrics)
	assert.Len(t, filteredMetrics, 0)
}
