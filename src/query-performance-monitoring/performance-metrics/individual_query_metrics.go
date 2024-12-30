package performance_metrics

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateIndividualQueryMetrics(conn *performanceDbConnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration, args args.ArgumentList) []datamodels.IndividualQueryMetrics {
	isExtensionEnabled, err := validations.CheckIndividualQueryMetricsFetchEligibility(conn)
	if err != nil {
		log.Error("Error validating eligibility for IndividualQueryMetrics: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Info("Extensions for PopulateIndividualQueryMetrics is not enabled.")
		return nil
	}
	log.Info("Extensions for PopulateIndividualQueryMetrics is enabled.")
	individualQueryMetricsInterface, individualQueriesForExecPlan := GetIndividualQueryMetrics(conn, args, slowRunningQueries)
	if len(individualQueryMetricsInterface) == 0 {
		log.Info("No individual queries found.")
		return nil
	}
	common_utils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, args)
	return individualQueriesForExecPlan
}

func ConstructIndividualQuery(slowRunningQueries []datamodels.SlowRunningQueryMetrics, args args.ArgumentList) string {
	var queryIDs []string
	for _, query := range slowRunningQueries {
		queryIDs = append(queryIDs, fmt.Sprintf("%d", *query.QueryID))
	}
	query := fmt.Sprintf(queries.IndividualQuerySearch, strings.Join(queryIDs, ","), args.QueryResponseTimeThreshold)
	log.Info("Individual Query Search Query: %s", query)
	return query
}

func GetIndividualQueryMetrics(conn *performanceDbConnection.PGSQLConnection, args args.ArgumentList, slowRunningQueries []datamodels.SlowRunningQueryMetrics) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	query := ConstructIndividualQuery(slowRunningQueries, args)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Info("Error executing query: %v", err)
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

		model.RealQueryText = model.QueryText
		model.QueryText = &anonymizedQueryText

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
	return anonymizeQueryMapByDb
}
