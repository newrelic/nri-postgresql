package performancemetrics

import (
	"fmt"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) error {
	var isEligible = validations.CheckWaitEventMetricsFetchEligibility(enabledExtensions)
	if !isEligible {
		log.Debug("Extension 'pg_wait_sampling' or 'pg_stat_statement' is not enabled or unsupported version.")
		return commonutils.ErrNotEligible
	}
	waitEventMetricsList, waitEventErr := getWaitEventMetrics(conn, cp)
	if waitEventErr != nil {
		log.Error("Error fetching wait event queries: %v", waitEventErr)
		return commonutils.ErrUnExpectedError
	}
	if len(waitEventMetricsList) == 0 {
		log.Debug("No wait event queries found.")
		return nil
	}
	err := commonutils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting wait event queries: %v", err)
		return err
	}
	return nil
}

func getWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	var query = fmt.Sprintf(queries.WaitEvents, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var waitEvent datamodels.WaitEventMetrics
		if waitScanErr := rows.StructScan(&waitEvent); waitScanErr != nil {
			return nil, err
		}

		waitEventMetricsList = append(waitEventMetricsList, waitEvent)
	}
	return waitEventMetricsList, nil
}

func PopulateWaitEventMetricsPgStat(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool, slowQueries []datamodels.SlowRunningQueryMetrics) error {
	waitEventMetricsList, waitEventErr := getWaitEventMetricsPgStat(conn, cp)
	if waitEventErr != nil {
		log.Error("Error fetching wait event queries: %v", waitEventErr)
		return commonutils.ErrUnExpectedError
	}
	if len(waitEventMetricsList) == 0 {
		log.Debug("No wait event queries found.")
		return nil
	}
	filteredWaitEvents := getFilteredWaitEvents(waitEventMetricsList, slowQueries)
	err := commonutils.IngestMetric(filteredWaitEvents, "PostgresWaitEvents", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting wait event queries: %v", err)
		return err
	}
	return nil
}

func getWaitEventMetricsPgStat(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]datamodels.WaitEventMetrics, error) {
	var waitEventMetricsList []datamodels.WaitEventMetrics
	var query = fmt.Sprintf(queries.WaitEventsFromPgStatActivity, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var waitEvent datamodels.WaitEventMetrics
		if waitScanErr := rows.StructScan(&waitEvent); waitScanErr != nil {
			return nil, err
		}

		waitEventMetricsList = append(waitEventMetricsList, waitEvent)
	}
	return waitEventMetricsList, nil
}

func getFilteredWaitEvents(waitEventMetrics []datamodels.WaitEventMetrics, slowQueryMetrics []datamodels.SlowRunningQueryMetrics) []interface{} {
	filteredWaitEventMetricInterface := make([]interface{}, 0)
	slowQueryTextMap := make(map[string]datamodels.SlowRunningQueryMetrics)
	for _, metric := range slowQueryMetrics {
		slowQueryTextMap[commonutils.AnonymizeAndNormalize(*metric.QueryText)] = metric
	}
	for _, waitEventMetric := range waitEventMetrics {
		normalizedWaitEventQueryText := commonutils.AnonymizeAndNormalize(*waitEventMetric.QueryText)
		if _, exists := slowQueryTextMap[normalizedWaitEventQueryText]; exists {
			waitEventMetric.QueryText = slowQueryTextMap[normalizedWaitEventQueryText].QueryText
			waitEventMetric.QueryID = slowQueryTextMap[normalizedWaitEventQueryText].QueryID
			filteredWaitEventMetricInterface = append(filteredWaitEventMetricInterface, waitEventMetric)
		}
	}
	return filteredWaitEventMetricInterface
}
