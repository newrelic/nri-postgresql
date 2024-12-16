package performance_metrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetBlockingMetrics(conn *performanceDbConnection.PGSQLConnection) ([]interface{}, error) {
	var blockingQueries []interface{}
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

	return blockingQueries, nil
}

func PopulateBlockingMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}
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
	common_utils.IngestMetric(blockingQueries, instanceEntity, "PostgresBlockingQueries")
	//for _, model := range blockingQueries {
	//	metricSet := instanceEntity.NewMetricSet("PostgresBlockingSessions")
	//
	//	modelValue := reflect.ValueOf(model)
	//	modelType := reflect.TypeOf(model)
	//
	//	for i := 0; i < modelValue.NumField(); i++ {
	//		field := modelValue.Field(i)
	//		fieldType := modelType.Field(i)
	//		metricName := fieldType.Tag.Get("metric_name")
	//		sourceType := fieldType.Tag.Get("source_type")
	//
	//		if field.Kind() == reflect.Ptr && !field.IsNil() {
	//			common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
	//		} else if field.Kind() != reflect.Ptr {
	//			common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
	//		}
	//	}
	//}

}
