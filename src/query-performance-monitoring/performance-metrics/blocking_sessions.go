package performance_metrics

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
	"reflect"
)

func GetBlockingMetrics(conn *performanceDbConnection.PGSQLConnection) ([]datamodels.BlockingQuery, error) {
	var blockingQueries []datamodels.BlockingQuery
	var query = queries.BlockingQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var blockingQuery datamodels.BlockingQuery
		if err := rows.StructScan(&blockingQuery); err != nil {
			return nil, err
		}
		blockingQueries = append(blockingQueries, blockingQuery)
	}

	for _, query := range blockingQueries {
		log.Info("Blocking Query: %+v", query)
	}
	return blockingQueries, nil
}

// PopulateSlowRunningMetrics fetches slow-running metrics and populates them into the metric set
func PopulateBlockingMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' enabled.")
		blockingQueries, err := GetBlockingMetrics(conn)
		if err != nil {
			log.Error("Error fetching Blocking queries: %v", err)
			return
		}

		if len(blockingQueries) == 0 {
			log.Info("No Blocking queries found.")
			return
		}
		log.Info("Populate Blocking running: %+v", blockingQueries)

		for _, model := range blockingQueries {
			metricSet := instanceEntity.NewMetricSet("PostgresBlockingQueries")

			modelValue := reflect.ValueOf(model)
			modelType := reflect.TypeOf(model)

			for i := 0; i < modelValue.NumField(); i++ {
				field := modelValue.Field(i)
				fieldType := modelType.Field(i)
				metricName := fieldType.Tag.Get("metric_name")
				sourceType := fieldType.Tag.Get("source_type")

				if field.Kind() == reflect.Ptr && !field.IsNil() {
					setMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
				} else if field.Kind() != reflect.Ptr {
					setMetric(metricSet, metricName, field.Interface(), sourceType)
				}
			}

			//	log.Info("Metrics set for slow query: %s in database: %s", *model.QueryID, *model.DatabaseName)
		}
	} else {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}

}

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
				setMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				log.Info("fielddddd", field.Interface())
				setMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}
	}
}

func GetExecutionPlanMetrics(results []datamodels.IndividualQuerySearch, args args.ArgumentList) []datamodels.QueryExecutionPlanMetrics {

	var executionPlanMetricsList []datamodels.QueryExecutionPlanMetrics

	var groupIndividualQueriesByDatabase = GroupQueriesByDatabase(results)

	for dbName, indiQueries := range groupIndividualQueriesByDatabase {
		fmt.Printf("Database: %s\n", dbName)
		fmt.Printf("Queries: %+v\n", indiQueries)
		newConn, err := performanceDbConnection.OpenDB(args, dbName)
		if err != nil {
			log.Error("Error opening database connection: %v", err)
			continue
		}
		defer newConn.Close()
		for _, individualQuery := range indiQueries {
			if err != nil {
				log.Error("Error opening database connection: %v", err)
				continue
			}
			log.Info("individualQuery", "")
			query := "EXPLAIN (FORMAT JSON) " + *individualQuery.QueryText
			rows, err := newConn.Queryx(query)
			if err != nil {
				continue
			}
			defer rows.Close()
			if !rows.Next() {
				return nil
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
				return nil
			}
			log.Info("execPlanMetrics", execPlanMetrics, execPlan[0]["Plan"])

			execPlanMetrics.QueryText = *individualQuery.QueryText
			execPlanMetrics.QueryId = *individualQuery.QueryId
			execPlanMetrics.DatabaseName = *individualQuery.DatabaseName

			fmt.Printf("QueryExecutionPlanMetricsssssss: %+v\n", execPlanMetrics)
			executionPlanMetricsList = append(executionPlanMetricsList, execPlanMetrics)
		}
	}

	//for _, individualQuery := range results {
	//	newConn, err := connection.OpenDB(args, *individualQuery.DatabaseName)
	//	if err != nil {
	//		log.Error("Error opening database connection: %v", err)
	//		continue
	//	}
	//	defer newConn.Close()
	//	log.Info("individualQuery", "")
	//	query := "EXPLAIN (FORMAT JSON) " + *individualQuery.QueryText
	//	rows, err := newConn.Queryx(query)
	//	if err != nil {
	//		continue
	//	}
	//	defer rows.Close()
	//	if !rows.Next() {
	//		return nil
	//	}
	//	var execPlanJSON string
	//	if err := rows.Scan(&execPlanJSON); err != nil {
	//		log.Error("Error scanning row: ", err.Error())
	//		continue
	//	}
	//
	//	var execPlan []map[string]interface{}
	//	err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
	//	if err != nil {
	//		log.Error("Failed to unmarshal execution plan: %v", err)
	//		continue
	//	}
	//
	//	var execPlanMetrics datamodels.QueryExecutionPlanMetrics
	//	err = mapstructure.Decode(execPlan[0]["Plan"], &execPlanMetrics)
	//	if err != nil {
	//		log.Error("Failed to decode execPlan to execPlanMetrics: %v", err)
	//		return nil
	//	}
	//	log.Info("execPlanMetrics", execPlanMetrics, execPlan[0]["Plan"])
	//
	//	execPlanMetrics.QueryText = *individualQuery.QueryText
	//	execPlanMetrics.QueryId = *individualQuery.QueryId
	//	execPlanMetrics.DatabaseName = *individualQuery.DatabaseName
	//
	//	fmt.Printf("QueryExecutionPlanMetricsssssss: %+v\n", execPlanMetrics)
	//	executionPlanMetricsList = append(executionPlanMetricsList, execPlanMetrics)
	//if err != nil {
	//	fmt.Println("Error unmarshalling JSON:", err)
	//	return nil
	//}
	//executionPlanMetricsList = append(executionPlanMetricsList, execPlanMetrics)
	//}
	return executionPlanMetricsList

}

func GroupQueriesByDatabase(results []datamodels.IndividualQuerySearch) map[string][]datamodels.IndividualQuerySearch {
	databaseMap := make(map[string][]datamodels.IndividualQuerySearch)

	for _, query := range results {
		dbName := *query.DatabaseName
		databaseMap[dbName] = append(databaseMap[dbName], query)
	}

	return databaseMap
}

func setMetric(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		metricSet.SetMetric(name, value, metric.GAUGE)
	case `attribute`:
		metricSet.SetMetric(name, value, metric.ATTRIBUTE)
	default:
		metricSet.SetMetric(name, value, metric.GAUGE)
	}
}
