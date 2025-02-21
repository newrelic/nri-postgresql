package performancemetrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
)

func fetchMetrics[T any](conn *performancedbconnection.PGSQLConnection, query string, logPrefix string) ([]T, []interface{}, error) {
	var metricsList []T
	var metricsListInterface []interface{}

	log.Debug("Executing query to fetch %s metrics", logPrefix)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query for %s, error: %v", logPrefix, err)
		return metricsList, metricsListInterface, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric T
		if scanErr := rows.StructScan(&metric); scanErr != nil {
			log.Error("Error scanning row into %s: %v", logPrefix, scanErr)
			return metricsList, metricsListInterface, scanErr
		}
		metricsList = append(metricsList, metric)
		metricsListInterface = append(metricsListInterface, metric)
	}
	log.Debug("Fetched %d %s metrics", len(metricsList), logPrefix)
	return metricsList, metricsListInterface, nil
}
