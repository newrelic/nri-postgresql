package performance_metrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performanceDbConnection.PGSQLConnection) ([]datamodels.SlowRunningQuery, []interface{}, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var slowQueriesInterface []interface{}
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
		slowQueriesInterface = append(slowQueriesInterface, slowQuery)
	}

	for _, query := range slowQueries {
		log.Info("Slow Query: %+v", query)
	}
	return slowQueries, slowQueriesInterface, nil
}

func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection) []datamodels.SlowRunningQuery {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return nil
	}

	log.Info("Extension 'pg_stat_statements' enabled.")
	slowQueries, slowQueriesInterface, err := GetSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueries) == 0 {
		log.Info("No slow-running queries found.")
		return nil
	}
	log.Info("Populate-slow running: %+v", slowQueries)
	common_utils.IngestMetric(slowQueriesInterface, instanceEntity, "PostgresSlowQueries")
	return slowQueries

}
