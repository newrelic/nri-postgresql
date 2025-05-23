package performancemetrics

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) []datamodels.IndividualQueryInfo {
	isEligible, err := validations.CheckIndividualQueryMetricsFetchEligibility(enabledExtensions)
	if err != nil {
		log.Error("Error executing query for eligibility check: %v", err)
		return []datamodels.IndividualQueryInfo{}
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_monitor' is not enabled or unsupported version.")
		return []datamodels.IndividualQueryInfo{}
	}
	log.Debug("Extension 'pg_stat_monitor' enabled.")
	individualQueryMetricsInterface, individualQueriesList := getIndividualQueryMetrics(conn, slowRunningQueries, cp)
	if len(individualQueryMetricsInterface) == 0 {
		log.Debug("No individual queries found.")
		return []datamodels.IndividualQueryInfo{}
	}
	err = commonutils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting individual queries: %v", err)
		return []datamodels.IndividualQueryInfo{}
	}
	log.Debug("Successfully ingested individual query metrics for databases")
	return individualQueriesList
}

func getIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, cp *commonparameters.CommonParameters) ([]interface{}, []datamodels.IndividualQueryInfo) {
	if len(slowRunningQueries) == 0 {
		log.Debug("No slow running queries found.")
		return []interface{}{}, []datamodels.IndividualQueryInfo{}
	}
	var individualQueryInfoList []datamodels.IndividualQueryInfo
	var individualQueryMetricsListInterface []interface{}
	versionSpecificIndividualQuery, err := commonutils.FetchVersionSpecificIndividualQueries(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return []interface{}{}, []datamodels.IndividualQueryInfo{}
	}
	for _, slowRunningMetric := range slowRunningQueries {
		if slowRunningMetric.QueryID == nil {
			continue
		}
		query := fmt.Sprintf(versionSpecificIndividualQuery, *slowRunningMetric.QueryID, cp.Databases, cp.QueryMonitoringResponseTimeThreshold, min(cp.QueryMonitoringCountThreshold, commonutils.MaxIndividualQueryCountThreshold))
		rows, err := conn.Queryx(query)
		if err != nil {
			log.Debug("Error executing query in individual query: %v", err)
			return []interface{}{}, []datamodels.IndividualQueryInfo{}
		}
		defer rows.Close()
		individualQuerySamplesList := processRows(rows)
		for _, individualQuery := range individualQuerySamplesList {
			individualQueryInfoList = append(individualQueryInfoList, setIndividualQueriesInfo(individualQuery))
			individualQuery.QueryText = slowRunningMetric.QueryText
			individualQueryMetricsListInterface = append(individualQueryMetricsListInterface, individualQuery)
		}
	}
	log.Debug("Fetched %d individual query metrics", len(individualQueryInfoList))
	return individualQueryMetricsListInterface, individualQueryInfoList
}

func setIndividualQueriesInfo(individualQueryMetrics datamodels.IndividualQueryMetrics) datamodels.IndividualQueryInfo {
	return datamodels.IndividualQueryInfo{
		DatabaseName:  individualQueryMetrics.DatabaseName,
		QueryID:       individualQueryMetrics.QueryID,
		PlanID:        individualQueryMetrics.PlanID,
		RealQueryText: individualQueryMetrics.QueryText,
	}
}

func processRows(rows *sqlx.Rows) []datamodels.IndividualQueryMetrics {
	var individualQueryMetricsList []datamodels.IndividualQueryMetrics
	for rows.Next() {
		var model datamodels.IndividualQueryMetrics
		if scanErr := rows.StructScan(&model); scanErr != nil {
			log.Error("Could not scan row: %v", scanErr)
			continue
		}
		if model.QueryID == nil || model.DatabaseName == nil {
			log.Error("QueryID or DatabaseName is nil")
			continue
		}
		individualQueryMetric := model
		generatedPlanID, err := commonutils.GeneratePlanID()
		if err != nil {
			log.Error("Error generating plan ID: %v", err)
			continue
		}
		individualQueryMetric.PlanID = &generatedPlanID
		individualQueryMetricsList = append(individualQueryMetricsList, individualQueryMetric)
	}
	log.Debug("Processed %d rows into individual query metrics", len(individualQueryMetricsList))
	return individualQueryMetricsList
}
