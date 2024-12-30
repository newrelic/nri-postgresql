package validations

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
)

func isExtensionEnabled(conn *performancedbconnection.PGSQLConnection, extensionName string) (bool, error) {
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

func CheckSlowQueryMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_statements")
}

func CheckWaitEventMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection) (bool, error) {
	pgWaitExtension, waitErr := isExtensionEnabled(conn, "pg_wait_sampling")
	if waitErr != nil {
		return false, waitErr
	}
	pgStatExtension, statErr := isExtensionEnabled(conn, "pg_stat_statements")
	if statErr != nil {
		return false, statErr
	}
	return pgWaitExtension && pgStatExtension, nil
}

func CheckBlockingSessionMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_statements")
}

func CheckIndividualQueryMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_monitor")
}
