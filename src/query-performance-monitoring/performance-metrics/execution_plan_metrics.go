package performance_metrics

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

var supportedStatements = map[string]bool{"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true, "WITH": true}

func PopulateExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, pgIntegration *integration.Integration, args args.ArgumentList) {

	if results == nil || len(results) == 0 {
		log.Info("No individual queries found.")
		return
	}

	executionDetailsList := GetExecutionPlanMetrics(results, args)
	log.Info("ExecutionPlanList len:", len(executionDetailsList))
	common_utils.IngestMetric(executionDetailsList, "PostgresExecutionPlanMetrics", pgIntegration, args)
}

func GetExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, args args.ArgumentList) []interface{} {

	var executionPlanMetricsList []interface{}

	var groupIndividualQueriesByDatabase = GroupQueriesByDatabase(results)

	for dbName, individualQueriesList := range groupIndividualQueriesByDatabase {
		dbConn, err := performanceDbConnection.OpenDB(args, dbName)
		if err != nil {
			log.Error("Error opening database connection: %v", err)
			continue
		}
		processExecutionPlanOfQueries(individualQueriesList, dbConn, &executionPlanMetricsList)
		dbConn.Close()
	}

	return executionPlanMetricsList

}

func processExecutionPlanOfQueries(individualQueriesList []datamodels.IndividualQueryMetrics, dbConn *performanceDbConnection.PGSQLConnection, executionPlanMetricsList *[]interface{}) {

	for _, individualQuery := range individualQueriesList {

		//queryText := strings.TrimSpace(*individualQuery.QueryText)
		//upperQueryText := strings.ToUpper(queryText)
		//log.Info("Query Text: %s", strings.Split(upperQueryText, " ")[0])
		//if !supportedStatements[strings.Split(upperQueryText, " ")[0]] {
		//	log.Info("Skipping unsupported query for EXPLAIN: %s", queryText)
		//	continue
		//}

		query := "EXPLAIN (FORMAT JSON) " + *individualQuery.RealQueryText
		rows, err := dbConn.Queryx(query)
		if err != nil {
			log.Info("Error executing query: %v", err)
			continue
		}
		defer rows.Close()
		if !rows.Next() {
			log.Info("Execution plan not found for queryId", *individualQuery.QueryId)
			continue
		}
		var execPlanJSON string
		if err := rows.Scan(&execPlanJSON); err != nil {
			log.Error("Error scanning row: ", err.Error())
			continue
		}

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}
		level := 0
		fetchNestedExecutionPlanDetails(individualQuery, &level, execPlan[0]["Plan"].(map[string]interface{}), executionPlanMetricsList)
	}
}

func GroupQueriesByDatabase(results []datamodels.IndividualQueryMetrics) map[string][]datamodels.IndividualQueryMetrics {
	databaseMap := make(map[string][]datamodels.IndividualQueryMetrics)

	for _, query := range results {
		dbName := *query.DatabaseName
		databaseMap[dbName] = append(databaseMap[dbName], query)
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
	execPlanMetrics.QueryText = *individualQuery.QueryText
	execPlanMetrics.QueryId = *individualQuery.QueryId
	execPlanMetrics.DatabaseName = *individualQuery.DatabaseName
	execPlanMetrics.Level = *level
	*level = *level + 1
	if individualQuery.PlanId != nil {
		execPlanMetrics.PlanId = *individualQuery.PlanId
	} else {
		execPlanMetrics.PlanId = 999
	}

	*executionPlanMetricsList = append(*executionPlanMetricsList, execPlanMetrics)

	if nestedPlans, ok := execPlan["Plans"].([]interface{}); ok {
		for _, nestedPlan := range nestedPlans {
			if nestedPlanMap, ok := nestedPlan.(map[string]interface{}); ok {
				fetchNestedExecutionPlanDetails(individualQuery, level, nestedPlanMap, executionPlanMetricsList)
			}
		}
	}
}
