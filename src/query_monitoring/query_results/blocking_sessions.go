package query_results

import (
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/validations"
)

func GetBlockingMetrics(conn *connection.PGSQLConnection) ([]datamodels.BlockingQuery, error) {
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
func PopulateBlockingMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
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
func PopulateIndividualQueryMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection) {
	individualQueries := GetIndividualQueryMetrics(conn)
	if len(individualQueries) == 0 {
		log.Info("No individual queries found.")
		return
	}
	log.Info("Populate individual queries: %+v", individualQueries)

	for _, model := range individualQueries {
		metricSet := instanceEntity.NewMetricSet("PostgresIndividualQueriesSample")

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
	}
}

func GetIndividualQueryMetrics(conn *connection.PGSQLConnection) []interface{} {
	rows, err := conn.Queryx("select query from pg_stat_monitor where query like 'select * from actor%'")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var model datamodels.IndividualQuerySearch
		if err := rows.StructScan(&model); err != nil {
			log.Error("Could not scan row: ", err)
			continue
		}
		results = append(results, model)
	}
	log.Info("resultsss", results)
	return results

}
