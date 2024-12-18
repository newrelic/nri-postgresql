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

func GetWaitEventMetrics(conn *performanceDbConnection.PGSQLConnection, args args.ArgumentList) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	query := fmt.Sprintf(queries.WaitEvents, args.QueryCountThreshold)
	fmt.Print("Query: ", query)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var waitEvent datamodels.WaitEventMetrics
		if err := rows.StructScan(&waitEvent); err != nil {
			return nil, err
		}
		waitEventMetricsList = append(waitEventMetricsList, waitEvent)
	}
	return waitEventMetricsList, nil
}

func PopulateWaitEventMetrics(conn *performanceDbConnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList) {
	isExtensionEnabled, err := validations.CheckPgWaitSamplingExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return
	}
	log.Info("Extension 'pg_wait_sampling' enabled.")
	waitEventMetricsList, err := GetWaitEventMetrics(conn, args)
	if err != nil {
		log.Error("Error fetching wait event queries: %v", err)
		return
	}

	if len(waitEventMetricsList) == 0 {
		log.Info("No wait event queries found.")
		return
	}
	log.Info("Populate wait event : %+v", waitEventMetricsList)

	common_utils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, args)

}
