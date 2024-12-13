package validations

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
)

func CheckPgStatStatementsExtensionEnabled(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	rows, err := conn.Queryx("SELECT count(*) FROM pg_extension WHERE extname = 'pg_stat_statements'")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false, err
	}
	defer rows.Close()
	var isEnable bool
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("Error scanning rows: ", err.Error())
		}
	}
	if err := rows.Err(); err != nil {
		log.Error(err.Error())
	}
	if count > 0 {
		isEnable = true
	} else {
		isEnable = false
	}
	return isEnable, nil
}
