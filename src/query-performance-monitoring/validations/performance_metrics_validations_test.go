package validations

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCheckBlockingSessionMetricsFetchEligibilityExtensionNotRequired(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(12)
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilityUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilityUnSupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_monitor")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_monitor")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	isExtensionEnabledTest, _ := CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibilityUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckWaitEventMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibility(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(15)
	validationWait := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_wait_sampling")
	validationStat := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")

	testCases := []struct {
		waitCount int
		statCount int
		expected  bool
	}{
		{1, 1, true},  // Success
		{0, 1, false}, // Fail V1
		{1, 0, false}, // Fail V2
	}

	for _, tc := range testCases {
		mock.ExpectQuery(regexp.QuoteMeta(validationWait)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.waitCount))
		mock.ExpectQuery(regexp.QuoteMeta(validationStat)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.statCount))
		isExtensionEnabledTest, _ := CheckWaitEventMetricsFetchEligibility(conn, version)
		assert.Equal(t, isExtensionEnabledTest, tc.expected)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestCheckSlowQueryMetricsFetchEligibilityUnSupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	isExtensionEnabledTest, _ := CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}
