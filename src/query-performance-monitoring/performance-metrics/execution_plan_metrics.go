package performance_metrics

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, args args.ArgumentList, pgIntegration *integration.Integration) {

	if len(results) == 0 {
		log.Info("No individual queries found.")
		return
	}
	log.Info("Populate individual queries: %+v", results)

	executionDetailsList := GetExecutionPlanMetrics(results, args)

	log.Info("executionDetailsList", executionDetailsList)

	common_utils.IngestMetric(executionDetailsList, "PostgresExecutionPlanMetricsV5", pgIntegration)
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
		defer dbConn.Close()
		processExecutionPlanOfQueries(individualQueriesList, dbConn, &executionPlanMetricsList)
	}

	return executionPlanMetricsList

}

func processExecutionPlanOfQueries(individualQueriesList []datamodels.IndividualQueryMetrics, dbConn *performanceDbConnection.PGSQLConnection, executionPlanMetricsList *[]interface{}) {

	for _, individualQuery := range individualQueriesList {
		log.Info("individualQuery", "")
		query := "EXPLAIN (FORMAT JSON) " + *individualQuery.QueryText
		rows, err := dbConn.Queryx(query)
		if err != nil {
			continue
		}
		defer rows.Close()
		if !rows.Next() {
			log.Info("Execution plan not found for queryId", *individualQuery.QueryId)
			continue
		}
		var execPlanJSON string
		//var execPlanModal datamodels.IndividualQueryMetrics
		if err := rows.Scan(&execPlanJSON); err != nil {
			log.Error("Error scanning row: ", err.Error())
			continue
		}

		//if err := rows.StructScan(&execPlanModal); err != nil {
		//	log.Error("Error scanning row: ", err.Error())
		//	continue
		//}
		//
		//log.Info("execPlanModal", execPlanModal)

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}

		var execPlanMetrics datamodels.QueryExecutionPlanMetrics
		err = mapstructure.Decode(execPlan[0]["Plan"], &execPlanMetrics)
		if err != nil {
			log.Error("Failed to decode execPlan to execPlanMetrics: %v", err)
			continue
		}
		execPlanMetrics.QueryText = *individualQuery.QueryText
		execPlanMetrics.QueryId = *individualQuery.QueryId
		execPlanMetrics.DatabaseName = *individualQuery.DatabaseName
		execPlanMetrics.PlanId = *individualQuery.PlanId

		fmt.Printf("executionPlanMetrics: %+v\n", execPlanMetrics)
		*executionPlanMetricsList = append(*executionPlanMetricsList, execPlanMetrics)
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
