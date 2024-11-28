package validations

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
)

//func IsExtensionEnabled(conn *connection.PGSQLConnection, extensionName string) {
//
//	//Execute the SQL query to check the extension
//	rows, err := conn.Queryx("SELECT count(*) FROM pg_extension WHERE extname = 'pg_stat_statements'")
//	if err != nil {
//		log.Error("Error executing query: ", err.Error())
//		return
//	}
//	defer func() {
//		if err := rows.Close(); err != nil {
//			log.Error("Error closing rows: ", err.Error())
//		}
//	}()
//
//	// Variable to hold the count
//	var count int
//
//	// Iterate over the rows and scan the count
//	for rows.Next() {
//		if err := rows.Scan(&count); err != nil {
//			log.Error("Error scanning rows: ", err.Error())
//		}
//	}
//
//	// Check for any errors encountered during iteration
//	if err := rows.Err(); err != nil {
//		log.Error(err.Error())
//	}
//
//	// Print the appropriate message based on the count
//	if count > 0 {
//
//		fmt.Println("Extension enabled, count:", count)
//	} else {
//		fmt.Println("Extension disabled")
//	}
//}

func IsExtensionEnabled(conn *connection.PGSQLConnection, extensionName string) bool {
	// Execute the SQL query to check the extension
	query := "SELECT count(*) FROM pg_extension WHERE extname = $1"
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error("Error closing rows: ", err.Error())
		}
	}()

	// Variable to hold the count
	var count int

	// Iterate over the rows and scan the count
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("Error scanning rows: ", err.Error())
			return false
		}
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		log.Error(err.Error())
		return false
	}

	// Return true if the extension is enabled
	return count > 0
}
