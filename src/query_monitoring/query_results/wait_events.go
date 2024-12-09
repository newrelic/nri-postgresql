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

func GetWaitEventMetrics(conn *connection.PGSQLConnection) ([]datamodels.WaitEventQuery, error) {
	var waitQueries []datamodels.WaitEventQuery
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
		waitQueries = append(waitQueries, waitQuery)
	}

	for _, query := range waitQueries {
		log.Info("Wait Query: %+v", query)
	}
	return waitQueries, nil
}

// PopulateSlowRunningMetrics fetches slow-running metrics and populates them into the metric set
func PopulateWaitEventMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection) {
	isExtensionEnabled, err := validations.CheckPgWaitSamplingExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
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

		for _, model := range waitQueries {
			//metricSet := instanceEntity.NewMetricSet("PostgresWaitQueries")

			//modelValue := reflect.ValueOf(model)
			//modelType := reflect.TypeOf(model)

			for _, model := range waitQueries {
				metricSet := instanceEntity.NewMetricSet("PostgresWaitQueries")
				modelValue := reflect.ValueOf(model)
				modelType := reflect.TypeOf(model)
				for i := 0; i < modelValue.NumField(); i++ {
					field := modelValue.Field(i)
					fieldName := modelType.Field(i).Tag.Get("metric_name")
					sourceType := modelType.Field(i).Tag.Get("source_type")
					setMetric(metricSet, fieldName, field.Interface(), sourceType)
				}
			}

			log.Info("Metrics set for slow query: %s in database: %s", *model.QueryID, *model.DatabaseName)
		}
	} else {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return
	}

}
