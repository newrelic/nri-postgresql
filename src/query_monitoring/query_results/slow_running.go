package query_results

import (

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
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
// 	if !validations.IsExtensionEnabled(conn, "pg_stat_statements") {
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


func GetSlowRunningMetrics(conn *connection.PGSQLConnection, query string) ([]datamodels.SlowRunningQuery, error) {
	var slowQueries []datamodels.SlowRunningQuery

	err := conn.Query(&slowQueries, query) //use QueryContext
	if err != nil {
		return nil, err
	}
	return slowQueries, nil
	//log.Info("slow-running",slowQueries)
}

// PopulateSlowRunningMetrics fetches slow-running metrics and populates them into the metric set
func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
	
	// Check if the extension is enabled
	if !validations.IsExtensionEnabled(conn, "pg_stat_statements") {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}
	
	slowQueries, err := GetSlowRunningMetrics(conn, query)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return
	}

	if len(slowQueries) == 0 {
		return
	}

	// for _, query := range slowQueries {
	// 	metricSet := instanceEntity.NewMetricSet("PostgresSlowQueriesGo")
	// 	val := reflect.ValueOf(query)
	// 	typ := reflect.TypeOf(query)
	// 	for i := 0; i < val.NumField(); i++ {
	// 		field := val.Field(i)
	// 		fieldType := typ.Field(i)
	// 		metricName := fieldType.Tag.Get("metric_name")
	// 		sourceType := fieldType.Tag.Get("source_type")
	// 		if metricName != "" && !field.IsNil() {
	// 			metricSet.SetMetric(metricName, field.Elem().Interface(),)
	// 		}
	// 	}
	// }
}
