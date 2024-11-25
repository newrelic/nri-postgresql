package query_monitoring

import (
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

type blockingQueries struct {
	BlockedPid         *int64  `db:"blocked_pid"`
	BlockedQuery       *string `db:"blocked_query"`
	BlockedQueryId     *string `db:"blocked_query_id"`
	BlockedQueryStart  *string `db:"blocked_query_start"`
	BlockedDatabase    *string `db:"blocked_database"`
	BlockingPid        *int64  `db:"blocking_pid"`
	BlockingQuery      *string `db:"blocking_query"`
	BlockingQueryId    *string `db:"blocking_query_id"`
	BlockingQueryStart *string `db:"blocking_query_start"`
	BlockingDatabase   *string `db:"blocking_database"`
}

func AnalyzeBlockingQueries(instanceEntity *integration.Entity, connection *connection.PGSQLConnection, arguments args.ArgumentList) {
	log.Info("Querying PSQL Server for blocking queries")

	var getBlockingQuery = `SELECT blocked_activity.pid AS blocked_pid, blocked_statements.query AS blocked_query, blocked_statements.queryid AS blocked_query_id, blocked_activity.query_start AS blocked_query_start, blocked_activity.datname AS blocked_database, blocking_activity.pid AS blocking_pid, blocking_statements.query AS blocking_query, blocking_statements.queryid AS blocking_query_id, blocking_activity.query_start AS blocking_query_start, blocking_activity.datname AS blocking_database FROM pg_stat_activity AS blocked_activity JOIN pg_stat_statements AS blocked_statements ON blocked_activity.query_id = blocked_statements.queryid JOIN pg_locks blocked_locks ON blocked_activity.pid = blocked_locks.pid JOIN pg_locks blocking_locks ON blocked_locks.locktype = blocking_locks.locktype AND blocked_locks.database IS NOT DISTINCT FROM blocking_locks.database AND blocked_locks.relation IS NOT DISTINCT FROM blocking_locks.relation AND blocked_locks.page IS NOT DISTINCT FROM blocking_locks.page AND blocked_locks.tuple IS NOT DISTINCT FROM blocking_locks.tuple AND blocked_locks.transactionid IS NOT DISTINCT FROM blocking_locks.transactionid AND blocked_locks.classid IS NOT DISTINCT FROM blocking_locks.classid AND blocked_locks.objid IS NOT DISTINCT FROM blocking_locks.objid AND blocked_locks.objsubid IS NOT DISTINCT FROM blocking_locks.objsubid AND blocked_locks.pid <> blocking_locks.pid JOIN pg_stat_activity AS blocking_activity ON blocking_locks.pid = blocking_activity.pid JOIN pg_stat_statements AS blocking_statements ON blocking_activity.query_id = blocking_statements.queryid WHERE NOT blocked_locks.granted;`

	log.Info("Executing query to get blocking queries details.")

	// Ensure the connection is not nil
	if connection == nil {
		log.Error("Database connection is nil")
		return
	}

	// Slice to hold query results.
	blockingQueryModels := make([]blockingQueries, 0)

	if err := connection.Query(&blockingQueryModels, getBlockingQuery); err != nil {
		log.Error("Could not execute query: %s", err.Error())
		return
	}

	log.Info("Number of records retrieved: %d", len(blockingQueryModels))

	// Log and report each result from the query.
	for _, model := range blockingQueryModels {
		if model.BlockedDatabase == nil || model.BlockedQuery == nil {
			log.Warn("Skipping entry with nil field: BlockedQueryText or BlockingDatabaseName")
			continue // Skip this entry if any critical field is nil
		}

		blockedQuery := *model.BlockedQuery
		blockedDatabase := *model.BlockedDatabase
		blockedPid := int64(0)
		if model.BlockedPid != nil {
			blockedPid = *model.BlockedPid
		}
		blockedQueryId := ""
		if model.BlockedQueryId != nil {
			blockedQueryId = *model.BlockedQueryId
		}
		blockedQueryStart := ""
		if model.BlockedQueryStart != nil {
			blockedQueryStart = *model.BlockedQueryStart
		}
		if model.BlockedPid != nil {
			blockedPid = *model.BlockedPid
		}
		blockingPid := int64(0)
		if model.BlockingPid != nil {
			blockingPid = *model.BlockingPid
		}
		blockingQueryId := ""
		if model.BlockingQueryId != nil {
			blockingQueryId = *model.BlockingQueryId
		}
		blockingQueryStart := ""
		if model.BlockingQueryStart != nil {
			blockingQueryStart = *model.BlockingQueryStart
		}
		blockingQuery := ""
		if model.BlockingQuery != nil {
			blockingQuery = *model.BlockingQuery
		}
		blockingDatabase := ""
		if model.BlockingDatabase != nil {
			blockingDatabase = *model.BlockingDatabase
		}

		log.Info("Metrics set for blocking sessions: BlockedPid: %d, BlockedQuery: %s, BlockedQueryId: %s, BlockedDatabase: %s, blockedQueryStart: %s, BlockingPid: %d, BlockingQuery: %s, BlockingQueryId: %s, BlockingDatabase: %s, blockingQueryStart: %s",
			blockedPid,
			blockedQuery,
			blockedQueryId,
			blockedDatabase,
			blockedQueryStart,
			blockingPid,
			blockingQuery,
			blockingQueryId,
			blockingDatabase,
			blockingQueryStart)

		metricSet := instanceEntity.NewMetricSet("PostgresBlockingSesssionsGo",
			attribute.Attribute{Key: "blockedQuery", Value: blockedQuery},
			attribute.Attribute{Key: "blockedQueryId", Value: blockedQueryId},
		)

		// Add all the fields to the metric set.
		if model.BlockedPid != nil {
			metricSet.SetMetric("blockedPid", *model.BlockedPid, metric.GAUGE)
		}
		if model.BlockedDatabase != nil {
			metricSet.SetMetric("blockedDatabase", *model.BlockedDatabase, metric.ATTRIBUTE)
		}
		if model.BlockedQueryStart != nil {
			metricSet.SetMetric("blockedQueryStart", *model.BlockedQueryStart, metric.ATTRIBUTE)
		}
		if model.BlockingPid != nil {
			metricSet.SetMetric("blockingPid", *model.BlockingPid, metric.GAUGE)
		}
		if model.BlockingQuery != nil {
			metricSet.SetMetric("blockingQuery", *model.BlockingQuery, metric.ATTRIBUTE)
		}
		if model.BlockingQueryId != nil {
			metricSet.SetMetric("blockingQueryId", *model.BlockingQueryId, metric.ATTRIBUTE)
		}
		if model.BlockingDatabase != nil {
			metricSet.SetMetric("blockingDatabase", *model.BlockingDatabase, metric.ATTRIBUTE)
		}
		if model.BlockingQueryStart != nil {
			metricSet.SetMetric("blockingQueryStart", *model.BlockingQueryStart, metric.ATTRIBUTE)
		}

		log.Info("Metrics set for blocking query: %s in database: %s", blockedQuery, blockedDatabase)
	}

	log.Info("Completed processing all blocking query entries.")
}
