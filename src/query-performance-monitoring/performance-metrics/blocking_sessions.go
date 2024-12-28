package performance_metrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetBlockingMetrics(conn *performanceDbConnection.PGSQLConnection, args args.ArgumentList) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	query := fmt.Sprintf(queries.BlockingQueries, args.QueryCountThreshold)
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

func PopulateBlockingMetrics(conn *performanceDbConnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList) {
	isExtensionEnabled, err := validations.CheckBlockingSessionMetricsFetchEligibility(conn)
	if err != nil {
		log.Error("Error validating eligibility for BlockingSessions: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Ineligible to collect Blocking session metrics")
		return
	}
	log.Info("Extension for PopulateBlockingMetrics enabled.")
	blockingQueriesMetricsList, err := GetBlockingMetrics(conn, args)
	if err != nil {
		log.Error("Error fetching Blocking queries: %v", err)
		return
	}

	if len(blockingQueriesMetricsList) == 0 {
		log.Info("No Blocking queries found.")
		return
	}
	common_utils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, args)

}
