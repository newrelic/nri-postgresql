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

func GetWaitEventMetrics(conn *performanceDbConnection.PGSQLConnection) ([]interface{}, error) {
	var waitEventMetrics []interface{}
	var query = queries.WaitEvents
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var waitQuery datamodels.WaitEventQuery
		if err := rows.StructScan(&waitQuery); err != nil {
			return nil, err
		}
		waitEventMetrics = append(waitEventMetrics, waitQuery)
	}
	return waitEventMetrics, nil
}

func PopulateWaitEventMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection) {
	isExtensionEnabled, err := validations.CheckPgWaitSamplingExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return
	}
	log.Info("Extension 'pg_wait_sampling' enabled.")
	waitQueries, err := GetWaitEventMetrics(conn)
	if err != nil {
		log.Error("Error fetching wait event queries: %v", err)
		return
	}

	if len(waitQueries) == 0 {
		log.Info("No wait event queries found.")
		return
	}
	log.Info("Populate wait event : %+v", waitQueries)

	common_utils.IngestMetric(waitQueries, instanceEntity, "PostgresWaitEvents")
	//for _, model := range waitQueries {
	//	metricSet := instanceEntity.NewMetricSet("PostgresWaitEvents")
	//	modelValue := reflect.ValueOf(model)
	//	modelType := reflect.TypeOf(model)
	//	for i := 0; i < modelValue.NumField(); i++ {
	//		field := modelValue.Field(i)
	//		fieldType := modelType.Field(i)
	//		metricName := fieldType.Tag.Get("metric_name")
	//		sourceType := fieldType.Tag.Get("source_type")
	//		if field.Kind() == reflect.Ptr && !field.IsNil() {
	//			common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
	//		} else if field.Kind() != reflect.Ptr {
	//			common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
	//		}
	//	}
	//	log.Info("Metrics set for wait event queryId: %s in database: %s", *model.QueryID, *model.DatabaseName)
	//}

}
