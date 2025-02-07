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

type queryInfoMap map[string]string
type databaseQueryInfoMap map[string]queryInfoMap

func PopulateIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) []datamodels.IndividualQueryMetrics {
	isEligible, err := validations.CheckIndividualQueryMetricsFetchEligibility(enabledExtensions)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_monitor' is not enabled or unsupported version.")
		return nil
	}
	log.Debug("Extension 'pg_stat_monitor' enabled.")
	individualQueryMetricsInterface, individualQueriesList := GetIndividualQueryMetrics(conn, slowRunningQueries, cp)
	if len(individualQueryMetricsInterface) == 0 {
		log.Debug("No individual queries found.")
		return nil
	}
	err = commonutils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting individual queries: %v", err)
		return nil
	}
	return individualQueriesList
}

func GetIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, cp *commonparameters.CommonParameters) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	if len(slowRunningQueries) == 0 {
		log.Debug("No slow running queries found.")
		return nil, nil
	}
	var individualQueryMetricsList []datamodels.IndividualQueryMetrics
	var individualQueryMetricsListInterface []interface{}
	anonymizedQueriesByDB := processForAnonymizeQueryMap(slowRunningQueries)
	versionSpecificIndividualQuery, err := commonutils.FetchVersionSpecificIndividualQueries(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, nil
	}
	for _, slowRunningMetric := range slowRunningQueries {
		if slowRunningMetric.QueryID == nil {
			continue
		}
		query := fmt.Sprintf(versionSpecificIndividualQuery, *slowRunningMetric.QueryID, cp.Databases, cp.QueryMonitoringResponseTimeThreshold, min(cp.QueryMonitoringCountThreshold, commonutils.MaxIndividualQueryCountThreshold))
		rows, err := conn.Queryx(query)
		if err != nil {
			log.Debug("Error executing query in individual query: %v", err)
			return nil, nil
		}
		defer rows.Close()
		individualQuerySamplesList := processRows(rows, anonymizedQueriesByDB)
		for _, individualQuery := range individualQuerySamplesList {
			individualQueryMetricsList = append(individualQueryMetricsList, individualQuery)
			individualQueryMetricsListInterface = append(individualQueryMetricsListInterface, individualQuery)
		}
	}
	return individualQueryMetricsListInterface, individualQueryMetricsList
}

func processRows(rows *sqlx.Rows, anonymizedQueriesByDB databaseQueryInfoMap) []datamodels.IndividualQueryMetrics {
	var individualQueryMetricsList []datamodels.IndividualQueryMetrics
	for rows.Next() {
		var model datamodels.IndividualQueryMetrics
		if scanErr := rows.StructScan(&model); scanErr != nil {
			log.Error("Could not scan row: ", scanErr)
			continue
		}
		if model.QueryID == nil || model.DatabaseName == nil {
			log.Error("QueryID or DatabaseName is nil")
			continue
		}
		individualQueryMetric := model
		anonymizedQueryText := anonymizedQueriesByDB[*model.DatabaseName][*model.QueryID]
		queryText := *model.QueryText
		individualQueryMetric.RealQueryText = &queryText
		individualQueryMetric.QueryText = &anonymizedQueryText
		generatedPlanID, err := commonutils.GeneratePlanID()
		if err != nil {
			log.Error("Error generating plan ID: %v", err)
			continue
		}
		individualQueryMetric.PlanID = &generatedPlanID
		individualQueryMetricsList = append(individualQueryMetricsList, individualQueryMetric)
	}
	return individualQueryMetricsList
}

func processForAnonymizeQueryMap(slowRunningMetricList []datamodels.SlowRunningQueryMetrics) databaseQueryInfoMap {
	anonymizeQueryMapByDB := make(databaseQueryInfoMap)
	for _, metric := range slowRunningMetricList {
		if metric.DatabaseName == nil || metric.QueryID == nil || metric.QueryText == nil {
			continue
		}
		dbName := *metric.DatabaseName
		queryID := *metric.QueryID
		anonymizedQuery := *metric.QueryText

		if _, exists := anonymizeQueryMapByDB[dbName]; !exists {
			anonymizeQueryMapByDB[dbName] = make(map[string]string)
		}
		anonymizeQueryMapByDB[dbName][queryID] = anonymizedQuery
	}
	return anonymizeQueryMapByDB
}
