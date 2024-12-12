package query_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/query_results"
)

func RunAnalysis(instanceEntity *integration.Entity, args args.ArgumentList) {
	connectionInfo := connection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		fmt.Println("Error creating connection: ", err)
		return
	}
	slowRunningQueries := query_results.PopulateSlowRunningMetrics(instanceEntity, newConnection)
	query_results.PopulateWaitEventMetrics(instanceEntity, newConnection)
	query_results.PopulateBlockingMetrics(instanceEntity, newConnection)
	individualQueries := query_results.PopulateIndividualQueryMetrics(instanceEntity, newConnection, slowRunningQueries)
	query_results.PopulateExecutionPlanMetrics(instanceEntity, newConnection, individualQueries, args)
	fmt.Println("Query analysis completed.")
}
