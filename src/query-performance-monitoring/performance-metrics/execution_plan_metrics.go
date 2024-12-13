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
	"reflect"
)

func PopulateExecutionPlanMetrics(instanceEntity *integration.Entity, results []datamodels.IndividualQuerySearch, args args.ArgumentList) {

	if len(results) == 0 {
		log.Info("No individual queries found.")
		return
	}
	log.Info("Populate individual queries: %+v", results)

	executionDetailsList := GetExecutionPlanMetrics(results, args)

	log.Info("executionDetailsList", executionDetailsList)

	for _, model := range executionDetailsList {
		metricSet := instanceEntity.NewMetricSet("PostgresExecutionPlanMetricsSample")

		modelValue := reflect.ValueOf(model)
		modelType := reflect.TypeOf(model)

		for i := 0; i < modelValue.NumField(); i++ {
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				log.Info("fielddddd", field.Interface())
				common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}
	}
}

func GetExecutionPlanMetrics(results []datamodels.IndividualQuerySearch, args args.ArgumentList) []datamodels.QueryExecutionPlanMetrics {

	var executionPlanMetricsList []datamodels.QueryExecutionPlanMetrics

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

func processExecutionPlanOfQueries(individualQueriesList []datamodels.IndividualQuerySearch, dbConn *performanceDbConnection.PGSQLConnection, executionPlanMetricsList *[]datamodels.QueryExecutionPlanMetrics) {
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

		var execPlanMetrics datamodels.QueryExecutionPlanMetrics
		err = mapstructure.Decode(execPlan[0]["Plan"], &execPlanMetrics)
		if err != nil {
			log.Error("Failed to decode execPlan to execPlanMetrics: %v", err)
			continue
		}
		execPlanMetrics.QueryText = *individualQuery.QueryText
		execPlanMetrics.QueryId = *individualQuery.QueryId
		execPlanMetrics.DatabaseName = *individualQuery.DatabaseName

		fmt.Printf("executionPlanMetrics: %+v\n", execPlanMetrics)
		*executionPlanMetricsList = append(*executionPlanMetricsList, execPlanMetrics)
	}
}

func GroupQueriesByDatabase(results []datamodels.IndividualQuerySearch) map[string][]datamodels.IndividualQuerySearch {
	databaseMap := make(map[string][]datamodels.IndividualQuerySearch)

	for _, query := range results {
		dbName := *query.DatabaseName
		databaseMap[dbName] = append(databaseMap[dbName], query)
	}

	return databaseMap
}
