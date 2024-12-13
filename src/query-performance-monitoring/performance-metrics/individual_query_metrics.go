package performance_metrics

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
	"reflect"
	"strings"
)

func PopulateIndividualQueryMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQuery) []datamodels.IndividualQuerySearch {
	isExtensionEnabled, err := validations.CheckPgStatMonitorExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_monitor' is not enabled.")
		return nil
	}
	log.Info("Extension 'pg_stat_monitor' enabled.")
	individualQueries := GetIndividualQueryMetrics(conn, slowRunningQueries)
	if len(individualQueries) == 0 {
		log.Info("No individual queries found.")
		return nil
	}
	log.Info("Populate individual queries: %+v", individualQueries)

	for _, model := range individualQueries {
		metricSet := instanceEntity.NewMetricSet("PostgresIndividualQueries")

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
				common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}
	}
	return individualQueries
}

func ConstructIndividualQuery(slowRunningQueries []datamodels.SlowRunningQuery) string {
	var queryIDs []string
	for _, query := range slowRunningQueries {
		queryIDs = append(queryIDs, fmt.Sprintf("%d", *query.QueryID))
	}
	query := fmt.Sprintf(queries.IndividualQuerySearch, strings.Join(queryIDs, ","))
	log.Info("Individual query :", query)
	return query
}

func GetIndividualQueryMetrics(conn *performanceDbConnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQuery) []datamodels.IndividualQuerySearch {
	query := ConstructIndividualQuery(slowRunningQueries)
	log.Info("Individual query :", query)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []datamodels.IndividualQuerySearch
	for rows.Next() {
		var model datamodels.IndividualQuerySearch
		if err := rows.StructScan(&model); err != nil {
			log.Error("Could not scan row: ", err)
			continue
		}
		results = append(results, model)
	}
	return results

}
