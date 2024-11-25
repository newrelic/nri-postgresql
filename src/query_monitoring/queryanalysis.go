package query_monitoring

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

func RunAnalysis(instanceEntity *integration.Entity, connection *connection.PGSQLConnection, arguments args.ArgumentList) {
	AnalyzeSlowQueries(instanceEntity, connection, arguments)
	WaitEventQueries(instanceEntity, connection, arguments)
	AnalyzeBlockingQueries(instanceEntity, connection, arguments)
	fmt.Println("Query analysis completed.")
}
