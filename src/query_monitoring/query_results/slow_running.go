package query_results

import (
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/validations"
)

// FetchAndLogSlowRunningQueries fetches slow-running queries and logs the results
//func FetchAndLogSlowRunningQueries(instanceEntity *integration.Entity, conn *connection.PGSQLConnection) {
//	var slowQueries []datamodels.SlowRunningQuery
//
//	// Execute the slow queries SQL
//	err := conn.Query(&slowQueries, queries.SlowQueries)
//	if err != nil {
//		log.Error("Error fetching slow-running queries: %v", err)
//		return
//	}
//
//	// Log the results
//	for _, query := range slowQueries {
//		log.Info("Slow Query: %+v", query)
//		//	//	//log.Info("Slow Query: ID=%d, Text=%s, Database=%s, Schema=%s, ExecutionCount=%d, AvgElapsedTimeMs=%.3f, AvgCPUTimeMs=%.3f, AvgDiskReads=%.3f, AvgDiskWrites=%.3f, StatementType=%s, CollectionTimestamp=%s",
//		//	//	//	*query.QueryID, *query.QueryText, *query.DatabaseName, *query.SchemaName, *query.ExecutionCount, *query.AvgElapsedTimeMs, *query.AvgCPUTimeMs, *query.AvgDiskReads, *query.AvgDiskWrites, *query.StatementType, *query.CollectionTimestamp)
//	}
//	// Log the results
//
//}

// GetSlowRunningMetrics executes the given query and returns the result
// func GetSlowRunningMetrics(conn *connection.PGSQLConnection, query string) ([]datamodels.SlowRunningQuery, error) {
// 	if !validations.CheckPgStatStatementsExtensionEnabled(conn, "pg_stat_statements") {
// 		log.Info("Extension 'pg_stat_statements' is not enabled.")
// 		return nil, nil
// 	}
// 	var slowQueries []datamodels.SlowRunningQuery

// 	err := conn.Query(&slowQueries, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return slowQueries, nil
// 	//log.Info("slow-running",slowQueries)
// }

func GetSlowRunningMetrics(conn *connection.PGSQLConnection) ([]datamodels.SlowRunningQuery, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
	}

	for _, query := range slowQueries {
		log.Info("Slow Query: %+v", query)
	}
	return slowQueries, nil
}

// PopulateSlowRunningMetrics fetches slow-running metrics and populates them into the metric set
func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' enabled.")
		slowQueries, err := GetSlowRunningMetrics(conn)
		if err != nil {
			log.Error("Error fetching slow-running queries: %v", err)
			return
		}

		if len(slowQueries) == 0 {
			log.Info("No slow-running queries found.")
			return
		}
		log.Info("Populate-slow running: %+v", slowQueries)

		for _, model := range slowQueries {
			metricSet := instanceEntity.NewMetricSet("PostgresSlowQueriestharun")

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
