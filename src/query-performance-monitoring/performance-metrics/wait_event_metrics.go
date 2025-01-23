package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	globalvariables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) error {
	var isEligible bool
	var eligibleCheckErr error
	isEligible, eligibleCheckErr = validations.CheckWaitEventMetricsFetchEligibility(conn, gv.Version)
	if eligibleCheckErr != nil {
		log.Error("Error executing query: %v", eligibleCheckErr)
		return commonutils.ErrUnExpectedError
	}
	if !isEligible {
		log.Debug("Extension 'pg_wait_sampling' or 'pg_stat_statement' is not enabled or unsupported version.")
		return commonutils.ErrNotEligible
	}
	waitEventMetricsList, waitEventErr := GetWaitEventMetrics(conn, gv)
	if waitEventErr != nil {
		log.Error("Error fetching wait event queries: %v", waitEventErr)
		return commonutils.ErrUnExpectedError
	}
	if len(waitEventMetricsList) == 0 {
		log.Debug("No wait event queries found.")
		return nil
	}
	commonutils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, gv)
	return nil
}

func GetWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, gv *globalvariables.GlobalVariables) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	var query = fmt.Sprintf(queries.WaitEvents, gv.DatabaseString, min(gv.QueryCountThreshold, commonutils.MaxQueryThreshold))
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
