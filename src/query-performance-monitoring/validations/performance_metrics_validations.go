package validations

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
)

func FetchAllExtensions(conn *performancedbconnection.PGSQLConnection) (map[string]bool, error) {
	rows, err := conn.Queryx("SELECT extname FROM pg_extension")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return nil, err
	}
	defer rows.Close()
	var enabledExtensions = make(map[string]bool)
	for rows.Next() {
		var extname string
		if err := rows.Scan(&extname); err != nil {
			log.Error("Error scanning rows: ", err.Error())
			return nil, err
		}
		enabledExtensions[extname] = true
	}
	return enabledExtensions, nil
}

func CheckSlowQueryMetricsFetchEligibility(enabledExtensions map[string]bool) bool {
	return enabledExtensions[commonutils.PgStatStatementExtension]
}

func CheckWaitEventMetricsFetchEligibility(enabledExtensions map[string]bool) bool {
	return enabledExtensions[commonutils.PgStatStatementExtension] && enabledExtensions[commonutils.PgWaitSamplingExtension]
}

func CheckBlockingSessionMetricsFetchEligibility(enabledExtensions map[string]bool, version uint64) bool {
	// Version 12 and 13 do not require the pg_stat_statements extension
	if version == commonutils.PostgresVersion12 || version == commonutils.PostgresVersion13 {
		return true
	}
	return enabledExtensions[commonutils.PgStatStatementExtension]
}

func CheckIndividualQueryMetricsFetchEligibility(enabledExtensions map[string]bool) bool {
	return enabledExtensions[commonutils.PgStatMonitorExtension]
}

func CheckPostgresVersionSupportForQueryMonitoring(version uint64) bool {
	return version >= commonutils.PostgresVersion12
}
