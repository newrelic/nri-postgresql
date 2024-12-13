package validations

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
)

func CheckPgWaitSamplingExtensionEnabled(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	// Execute the SQL query to check the extension
	rows, err := conn.Queryx("SELECT count(*) FROM pg_extension WHERE extname = 'pg_wait_sampling'")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false, err
	}
	defer rows.Close()
	// Variable to hold the count
	var isEnable bool
	var count int
	// Iterate over the rows and scan the count
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("Error scanning rows: ", err.Error())
		}
	}
	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		log.Error(err.Error())
	}
	// Print the appropriate message based on the count
	if count > 0 {
		isEnable = true
	} else {
		isEnable = false
	}
	return isEnable, nil
}
