package validations

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
)

func isExtensionEnabled(conn *performanceDbConnection.PGSQLConnection, extensionName string) (bool, error) {
	rows, err := conn.Queryx(fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", extensionName))
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("Error scanning rows: ", err.Error())
		}
	}
	if err := rows.Err(); err != nil {
		log.Error(err.Error())
	}

	return count > 0, nil
}

func CheckSlowQueryMetricsFetchEligibility(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_statements")
}

func CheckWaitEventMetricsFetchEligibility(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	pgWaitExtension, err := isExtensionEnabled(conn, "pg_wait_sampling")
	pgStatExtension, err := isExtensionEnabled(conn, "pg_stat_statements")
	if err != nil {
		return false, err
	}
	return pgWaitExtension && pgStatExtension, nil
}

func CheckBlockingSessionMetricsFetchEligibility(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_statements")
}

func CheckIndividualQueryMetricsFetchEligibility(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_monitor")
}
