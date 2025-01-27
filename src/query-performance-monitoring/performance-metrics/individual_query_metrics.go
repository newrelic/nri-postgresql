package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	globalvariables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) []datamodels.IndividualQueryMetrics {
	isEligible, err := validations.CheckIndividualQueryMetricsFetchEligibility(conn, gv.Version)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_monitor' is not enabled or unsupported version.")
		return nil
	}
	log.Debug("Extension 'pg_stat_monitor' enabled.")
	individualQueryMetricsInterface, individualQueriesForExecPlan := GetIndividualQueryMetrics(conn, slowRunningQueries, gv)
	if len(individualQueryMetricsInterface) == 0 {
		log.Debug("No individual queries found.")
		return nil
	}
	commonutils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, gv)
	return individualQueriesForExecPlan
}

func GetIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, gv *globalvariables.GlobalVariables) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	if len(slowRunningQueries) == 0 {
		log.Debug("No slow running queries found.")
		return nil, nil
	}
	var individualQueryMetricsForExecPlanList []datamodels.IndividualQueryMetrics
	var individualQueryMetricsListInterface []interface{}
	anonymizedQueriesByDB := processForAnonymizeQueryMap(slowRunningQueries)
	versionSpecificIndividualQuery, err := commonutils.FetchVersionSpecificIndividualQueries(gv.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, nil
	}

	for _, slowRunningMetric := range slowRunningQueries {
		if slowRunningMetric.QueryID == nil {
			continue
		}
		getIndividualQueriesSamples(conn, slowRunningMetric, gv, anonymizedQueriesByDB, &individualQueryMetricsForExecPlanList, &individualQueryMetricsListInterface, versionSpecificIndividualQuery)
	}
	return individualQueryMetricsListInterface, individualQueryMetricsForExecPlanList
}

func getIndividualQueriesSamples(conn *performancedbconnection.PGSQLConnection, slowRunningQueries datamodels.SlowRunningQueryMetrics, gv *globalvariables.GlobalVariables, anonymizedQueriesByDB map[string]map[string]string, individualQueryMetricsForExecPlanList *[]datamodels.IndividualQueryMetrics, individualQueryMetricsListInterface *[]interface{}, versionSpecificIndividualQuery string) {
	query := fmt.Sprintf(versionSpecificIndividualQuery, *slowRunningQueries.QueryID, gv.DatabaseString, gv.QueryResponseTimeThreshold, min(gv.QueryCountThreshold, commonutils.MaxIndividualQueryThreshold))
	if query == "" {
		log.Debug("Error constructing individual query")
		return
	}
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Debug("Error executing query in individual query: %v", err)
		return
	}
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
		individualQueryMetric.QueryText = &anonymizedQueryText
		generatedPlanID := commonutils.GeneratePlanID(*model.QueryID)
		individualQueryMetric.PlanID = generatedPlanID
		model.PlanID = generatedPlanID
		model.RealQueryText = model.QueryText
		model.QueryText = &anonymizedQueryText

		*individualQueryMetricsForExecPlanList = append(*individualQueryMetricsForExecPlanList, model)
		*individualQueryMetricsListInterface = append(*individualQueryMetricsListInterface, individualQueryMetric)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return
	}

}

func processForAnonymizeQueryMap(slowRunningMetricList []datamodels.SlowRunningQueryMetrics) map[string]map[string]string {
	anonymizeQueryMapByDB := make(map[string]map[string]string)

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
