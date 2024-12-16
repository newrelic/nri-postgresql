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

func GetBlockingMetrics(conn *performanceDbConnection.PGSQLConnection) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	var query = queries.BlockingQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if err := rows.StructScan(&blockingQueryMetric); err != nil {
			return nil, err
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}

	return blockingQueriesMetricsList, nil
}

func PopulateBlockingMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}
	log.Info("Extension 'pg_stat_statements' enabled.")
	blockingQueriesMetricsList, err := GetBlockingMetrics(conn)
	if err != nil {
		log.Error("Error fetching Blocking queries: %v", err)
		return
	}

	if len(blockingQueriesMetricsList) == 0 {
		log.Info("No Blocking queries found.")
		return
	}
	log.Info("Populate Blocking running: %+v", blockingQueriesMetricsList)
	common_utils.IngestMetric(blockingQueriesMetricsList, instanceEntity, "PostgresBlockingSessions")

}
