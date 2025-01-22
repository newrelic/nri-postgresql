package performancemetrics

import (
	"encoding/json"

	"github.com/go-viper/mapstructure/v2"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	globalvariables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
)

func PopulateExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) {
	if len(results) == 0 {
		log.Debug("No individual queries found.")
		return
	}
	executionDetailsList := GetExecutionPlanMetrics(results, gv)
	commonutils.IngestMetric(executionDetailsList, "PostgresExecutionPlanMetrics", pgIntegration, gv)
}

func GetExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, gv *globalvariables.GlobalVariables) []interface{} {
	var executionPlanMetricsList []interface{}
	var groupIndividualQueriesByDatabase = GroupQueriesByDatabase(results)
	for dbName, individualQueriesList := range groupIndividualQueriesByDatabase {
		connectionInfo := performancedbconnection.DefaultConnectionInfo(&gv.Arguments)
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
		query := "EXPLAIN (FORMAT JSON) " + *individualQuery.RealQueryText
		rows, err := dbConn.Queryx(query)
		if err != nil {
			log.Debug("Error executing query: %v", err)
			continue
		}
		if individualQuery.QueryText == nil || individualQuery.QueryID == nil || individualQuery.DatabaseName == nil {
			log.Error("QueryText, QueryID or Database Name is nil")
			continue
		}
		if !rows.Next() {
			log.Debug("Execution plan not found for queryId", *individualQuery.QueryID)
			continue
		}
		var execPlanJSON string
		if scanErr := rows.Scan(&execPlanJSON); scanErr != nil {
			log.Error("Error scanning row: ", scanErr.Error())
			continue
		}
		if closeErr := rows.Close(); closeErr != nil {
			log.Error("Error closing rows: %v", closeErr)
			continue
		}

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}

		level := 0
		FetchNestedExecutionPlanDetails(individualQuery, &level, execPlan[0]["Plan"].(map[string]interface{}), executionPlanMetricsList)
	}
}

func GroupQueriesByDatabase(results []datamodels.IndividualQueryMetrics) map[string][]datamodels.IndividualQueryMetrics {
	databaseMap := make(map[string][]datamodels.IndividualQueryMetrics)
	for _, query := range results {
		if query.DatabaseName == nil {
			continue
		}
		dbName := *query.DatabaseName
		databaseMap[dbName] = append(databaseMap[dbName], query)
	}
	return databaseMap
}

func FetchNestedExecutionPlanDetails(individualQuery datamodels.IndividualQueryMetrics, level *int, execPlan map[string]interface{}, executionPlanMetricsList *[]interface{}) {
	var execPlanMetrics datamodels.QueryExecutionPlanMetrics
	err := mapstructure.Decode(execPlan, &execPlanMetrics)
	if err != nil {
		log.Error("Failed to decode execPlan to execPlanMetrics: %v", err)
		return
	}

	execPlanMetrics.QueryText = *individualQuery.QueryText
	execPlanMetrics.QueryID = *individualQuery.QueryID
	execPlanMetrics.DatabaseName = *individualQuery.DatabaseName
	execPlanMetrics.Level = *level
	*level++
	execPlanMetrics.PlanID = *individualQuery.PlanID

	*executionPlanMetricsList = append(*executionPlanMetricsList, execPlanMetrics)

	if nestedPlans, ok := execPlan["Plans"].([]interface{}); ok {
		for _, nestedPlan := range nestedPlans {
			if nestedPlanMap, nestedOk := nestedPlan.(map[string]interface{}); nestedOk {
				FetchNestedExecutionPlanDetails(individualQuery, level, nestedPlanMap, executionPlanMetricsList)
			}
		}
	}
}
