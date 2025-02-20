package performancemetrics

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetBlockingMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	cp := commonparameters.SetCommonParameters(args, version, databaseName)
	expectedQuery := queries.BlockingQueriesForV12AndV13
	query := fmt.Sprintf(expectedQuery, databaseName, args.QueryMonitoringCountThreshold)
	rowData := []driver.Value{
		int64(123), "SELECT 1", "1233444", "2023-01-01 00:00:00", "testdb",
		int64(456), "SELECT 2", "4566", "2023-01-01 00:00:00",
	}
	expectedRows := [][]driver.Value{
		rowData, rowData,
	}
	mockRows := sqlmock.NewRows([]string{
		"blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name",
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
		anonymizeQuery := commonutils.AnonymizeQueryText(expectedRows[index][1].(string))
		blockingSession := blockingQueriesMetricsList[index].(datamodels.BlockingSessionMetrics)
		assert.Equal(t, expectedRows[index][0], *blockingSession.BlockedPid)
		assert.Equal(t, anonymizeQuery, *blockingSession.BlockedQuery)
		assert.Equal(t, expectedRows[index][2], *blockingSession.BlockedQueryID)
		assert.Equal(t, expectedRows[index][3], *blockingSession.BlockedQueryStart)
		assert.Equal(t, expectedRows[index][4], *blockingSession.BlockedDatabase)
		assert.Equal(t, expectedRows[index][5], *blockingSession.BlockingPid)
		assert.Equal(t, anonymizeQuery, *blockingSession.BlockingQuery)
		assert.Equal(t, expectedRows[index][7], *blockingSession.BlockingQueryID)
		assert.Equal(t, expectedRows[index][8], *blockingSession.BlockingQueryStart)
	}
}

func TestGetBlockingMetricsErr(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	cp := commonparameters.SetCommonParameters(args, version, databaseName)
	_, err := getBlockingMetrics(conn, cp)
	assert.EqualError(t, err, commonutils.ErrUnExpectedError.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
