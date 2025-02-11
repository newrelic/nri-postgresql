package performancemetrics

import (
	"encoding/json"

	"github.com/go-viper/mapstructure/v2"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, connectionInfo performancedbconnection.Info) {
	if len(results) == 0 {
		log.Debug("No individual queries found.")
		return
	}
	executionDetailsList := getExecutionPlanMetrics(results, connectionInfo)
	err := commonutils.IngestMetric(executionDetailsList, "PostgresExecutionPlanMetrics", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting Execution Plan metrics: %v", err)
		return
	}
}

func getExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, connectionInfo performancedbconnection.Info) []interface{} {
	var executionPlanMetricsList []interface{}
	var groupIndividualQueriesByDatabase = groupQueriesByDatabase(results)
	for dbName, individualQueriesList := range groupIndividualQueriesByDatabase {
		dbConn, err := connectionInfo.NewConnection(dbName)
		if err != nil {
			log.Error("Error opening database connection: %v", err)
			continue
		}
		processExecutionPlanOfQueries(individualQueriesList, dbConn, &executionPlanMetricsList)
		dbConn.Close()
	}

	return executionPlanMetricsList
}

func processExecutionPlanOfQueries(individualQueriesList []datamodels.IndividualQueryMetrics, dbConn *performancedbconnection.PGSQLConnection, executionPlanMetricsList *[]interface{}) {
	for _, individualQuery := range individualQueriesList {
		if individualQuery.RealQueryText == nil || individualQuery.QueryID == nil || individualQuery.DatabaseName == nil {
			log.Error("QueryText, QueryID or Database Name is nil")
			continue
		}
		query := "EXPLAIN (FORMAT JSON) " + *individualQuery.RealQueryText
		rows, err := dbConn.Queryx(query)
		if err != nil {
			log.Debug("Error executing query: %v", err)
			continue
		}
		defer rows.Close()
		if !rows.Next() {
			log.Debug("Execution plan not found for queryId", *individualQuery.QueryID)
			continue
		}
		var execPlanJSON string
		if scanErr := rows.Scan(&execPlanJSON); scanErr != nil {
			log.Error("Error scanning row: ", scanErr.Error())
			continue
		}

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}
		validateAndFetchNestedExecPlan(execPlan, individualQuery, executionPlanMetricsList)
	}
}

func validateAndFetchNestedExecPlan(execPlan []map[string]interface{}, individualQuery datamodels.IndividualQueryMetrics, executionPlanMetricsList *[]interface{}) {
	level := 0
	if len(execPlan) > 0 {
		if plan, ok := execPlan[0]["Plan"].(map[string]interface{}); ok {
			fetchNestedExecutionPlanDetails(individualQuery, &level, plan, executionPlanMetricsList)
		} else {
			log.Debug("execPlan is not in correct datatype")
		}
	} else {
		log.Debug("execPlan is empty")
	}
}

func groupQueriesByDatabase(results []datamodels.IndividualQueryMetrics) map[string][]datamodels.IndividualQueryMetrics {
	databaseMap := make(map[string][]datamodels.IndividualQueryMetrics)
	for _, individualQueryMetric := range results {
		if individualQueryMetric.DatabaseName == nil {
			continue
		}
		dbName := *individualQueryMetric.DatabaseName
		databaseMap[dbName] = append(databaseMap[dbName], individualQueryMetric)
	}
	return databaseMap
}

func fetchNestedExecutionPlanDetails(individualQuery datamodels.IndividualQueryMetrics, level *int, execPlan map[string]interface{}, executionPlanMetricsList *[]interface{}) {
	var execPlanMetrics datamodels.QueryExecutionPlanMetrics
	err := mapstructure.Decode(execPlan, &execPlanMetrics)
	if err != nil {
		log.Error("Failed to decode execPlan to execPlanMetrics: %v", err)
		return
	}
	execPlanMetrics.QueryID = *individualQuery.QueryID
	execPlanMetrics.DatabaseName = *individualQuery.DatabaseName
	execPlanMetrics.Level = *level
	*level++
	execPlanMetrics.PlanID = *individualQuery.PlanID
	*executionPlanMetricsList = append(*executionPlanMetricsList, execPlanMetrics)
	if nestedPlans, ok := execPlan["Plans"].([]interface{}); ok {
		for _, nestedPlan := range nestedPlans {
			if nestedPlanMap, nestedOk := nestedPlan.(map[string]interface{}); nestedOk {
				fetchNestedExecutionPlanDetails(individualQuery, level, nestedPlanMap, executionPlanMetricsList)
			}
		}
	}
}
