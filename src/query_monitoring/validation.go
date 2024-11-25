package query_monitoring

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/newrelic/nri-postgresql/src/args"
	"log"
	"strconv"
	"strings"
)

func PrintQueryOutput(argList args.ArgumentList) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		argList.Hostname, argList.Port, argList.Username, argList.Password, argList.Database)
	//fmt.Println("-------", connStr)

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	// Query the PostgreSQL version
	var version string
	err = db.QueryRow("SHOW server_version").Scan(&version)
	if err != nil {
		log.Fatal("Error querying PostgreSQL version: ", err)
	}

	// Parse the version number
	majorVersion, err := strconv.Atoi(strings.Split(version, ".")[0])
	if err != nil {
		log.Fatal("Error parsing PostgreSQL version: ", err)
	}

	// Check if the version is greater than or equal to 12
	if majorVersion >= 12 {
		// Execute the SQL query to check the extension
		rows, err := db.Query("SELECT count(*) FROM pg_extension WHERE extname = 'pg_stat_statements'")
		if err != nil {
			log.Fatal("Error executing query: ", err)
		}
		defer rows.Close()

		// Variable to hold the count
		var count int

		// Iterate over the rows and scan the count
		for rows.Next() {
			if err := rows.Scan(&count); err != nil {
				log.Fatal("Error scanning rows: ", err)
			}
		}

		// Check for any errors encountered during iteration
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		// Print the appropriate message based on the count
		if count > 0 {
			fmt.Println("Extension enabled, count:", count)
		} else {
			fmt.Println("Extension disabled")
		}
	} else {
		fmt.Println("PostgreSQL version is less than 12, upgrade your version to use pg_stat_statements")
	}
}
