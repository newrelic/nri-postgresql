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
	"strings"
)

func PopulateIndividualQueryMetrics(conn *performanceDbConnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration) []datamodels.IndividualQueryMetrics {
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
	individualQueryMetricsInterface, individualQueriesForExecPlan := GetIndividualQueryMetrics(conn, slowRunningQueries)
	if len(individualQueryMetricsInterface) == 0 {
		log.Info("No individual queries found.")
		return nil
	}
	log.Info("Populate individual queries: %+v forExecPlan : %+v", individualQueryMetricsInterface, individualQueriesForExecPlan)
	common_utils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueriesV5", pgIntegration)
	return individualQueriesForExecPlan
}

func ConstructIndividualQuery(slowRunningQueries []datamodels.SlowRunningQueryMetrics) string {
	var queryIDs []string
	for _, query := range slowRunningQueries {
		queryIDs = append(queryIDs, fmt.Sprintf("%d", *query.QueryID))
	}
	query := fmt.Sprintf(queries.IndividualQuerySearch, strings.Join(queryIDs, ","))
	log.Info("Individual query :", query)
	return query
}

func GetIndividualQueryMetrics(conn *performanceDbConnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	query := ConstructIndividualQuery(slowRunningQueries)
	log.Info("Individual query :", query)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	anonymizedQueriesByDb := processForAnonymizeQueryMap(slowRunningQueries)
	var individualQueryMetricsForExecPlanList []datamodels.IndividualQueryMetrics
	var individualQueryMetricsListInterface []interface{}
	for rows.Next() {

		var model datamodels.IndividualQueryMetrics
		if err := rows.StructScan(&model); err != nil {
			log.Error("Could not scan row: ", err)
			continue
		}
		individualQueryMetric := model
		anonymizedQueryText := anonymizedQueriesByDb[*model.DatabaseName][*model.QueryId]
		individualQueryMetric.QueryText = &anonymizedQueryText

		individualQueryMetricsForExecPlanList = append(individualQueryMetricsForExecPlanList, model)
		individualQueryMetricsListInterface = append(individualQueryMetricsListInterface, individualQueryMetric)
	}
	return individualQueryMetricsListInterface, individualQueryMetricsForExecPlanList

}

func processForAnonymizeQueryMap(queryCpuMetricsList []datamodels.SlowRunningQueryMetrics) map[string]map[int64]string {
	anonymizeQueryMapByDb := make(map[string]map[int64]string)

	for _, metric := range queryCpuMetricsList {
		dbName := *metric.DatabaseName
		queryID := *metric.QueryID
		anonymizedQuery := *metric.QueryText

		if _, exists := anonymizeQueryMapByDb[dbName]; !exists {
			anonymizeQueryMapByDb[dbName] = make(map[int64]string)
		}
		anonymizeQueryMapByDb[dbName][queryID] = anonymizedQuery
	}
	log.Info("Anonymize Query Map By Db: %+v", anonymizeQueryMapByDb)
	return anonymizeQueryMapByDb
}
