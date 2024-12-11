package query_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/query_results"
)

func RunAnalysis(instanceEntity *integration.Entity, connection *connection.PGSQLConnection, arguments args.ArgumentList) {
	query_results.PopulateSlowRunningMetrics(instanceEntity, connection, queries.SlowQueries)
	query_results.PopulateWaitEventMetrics(instanceEntity, connection)
	query_results.PopulateBlockingMetrics(instanceEntity, connection, queries.BlockingQueries)
	individualQueriesList = query_results.PopulateIndividualQueryMetrics(instanceEntity, connection)
	fmt.Println("Query analysis completed.")
}
