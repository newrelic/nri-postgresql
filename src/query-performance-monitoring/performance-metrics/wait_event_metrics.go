package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList, databaseNames string) {
	isExtensionEnabled, err := validations.CheckWaitEventMetricsFetchEligibility(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Debug("Extension 'pg_wait_sampling' or 'pg_stat_statement' is not enabled.")
		return
	}
	waitEventMetricsList, err := GetWaitEventMetrics(conn, args, databaseNames)
	if err != nil {
		log.Error("Error fetching wait event queries: %v", err)
		return
	}

	if len(waitEventMetricsList) == 0 {
		log.Debug("No wait event queries found.")
		return
	}
	commonutils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, args)
}

func GetWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList, databaseNames string) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	var query = fmt.Sprintf(queries.WaitEvents, databaseNames, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var waitEvent datamodels.WaitEventMetrics
		if waitScanErr := rows.StructScan(&waitEvent); waitScanErr != nil {
			return nil, err
		}
		waitEventMetricsList = append(waitEventMetricsList, waitEvent)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, closeErr
	}
	return waitEventMetricsList, nil
}
