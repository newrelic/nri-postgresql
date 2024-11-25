package query_monitoring

import (
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

type topSlowQueries struct {
	QueryID             *string  `db:"query_id"`
	QueryText           *string  `db:"query_text"`
	DatabaseName        *string  `db:"database_name"`
	SchemaName          *string  `db:"schema_name"`
	ExecutionCount      *int64   `db:"execution_count"`
	AvgCPUTimeMS        *float64 `db:"avg_cpu_time_ms"`
	AvgElapsedTimeMS    *float64 `db:"avg_elapsed_time_ms"`
	AvgDiskReads        *float64 `db:"avg_disk_reads"`
	AvgDiskWrites       *float64 `db:"avg_disk_writes"`
	StatementType       *string  `db:"statement_type"`
	CollectionTimestamp *string  `db:"collection_timestamp"`
}

func AnalyzeSlowQueries(instanceEntity *integration.Entity, connection *connection.PGSQLConnection, arguments args.ArgumentList) {
	log.Info("Querying SQL Server for top N slow queries")
	//fmt.Println("====", connection)
	var getTopNSlowQueryDetailsQuery = `SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema() AS schema_name, pss.calls AS execution_count, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid ORDER BY avg_elapsed_time_ms DESC LIMIT 5;`

	log.Info("Executing query to get top N slow query details.")

	// Ensure the connection is not nil
	if connection == nil {
		log.Error("Database connection is nil")
		return
	}

	// Slice to hold query results.
	slowQueryModels := make([]topSlowQueries, 0)

	// Execute the query and store the results in the slowQueryModels slice.
	if err := connection.Query(&slowQueryModels, getTopNSlowQueryDetailsQuery); err != nil {
		log.Error("Could not execute query: %s", err.Error())
		return
	}

	log.Info("Number of records retrieved: %d", len(slowQueryModels))

	// Log and report each result from the query.
	for _, model := range slowQueryModels {
		if model.DatabaseName == nil || model.QueryText == nil || model.QueryID == nil {
			log.Warn("Skipping entry with nil field: DatabaseName, QueryText, or QueryID")
			continue // Skip this entry if any critical field is nil
		}

		queryID := *model.QueryID
		databaseName := *model.DatabaseName
		queryText := *model.QueryText
		schemaName := ""
		if model.SchemaName != nil {
			schemaName = *model.SchemaName
		}
		executionCount := int64(0)
		if model.ExecutionCount != nil {
			executionCount = *model.ExecutionCount
		}
		avgCPUTimeMS := float64(0)
		if model.AvgCPUTimeMS != nil {
			avgCPUTimeMS = *model.AvgCPUTimeMS
		}
		avgElapsedTimeMS := float64(0)
		if model.AvgElapsedTimeMS != nil {
			avgElapsedTimeMS = *model.AvgElapsedTimeMS
		}
		avgDiskReads := float64(0)
		if model.AvgDiskReads != nil {
			avgDiskReads = *model.AvgDiskReads
		}
		avgDiskWrites := float64(0)
		if model.AvgDiskWrites != nil {
			avgDiskWrites = *model.AvgDiskWrites
		}
		statementType := ""
		if model.StatementType != nil {
			statementType = *model.StatementType
		}
		collectionTimestamp := ""
		if model.CollectionTimestamp != nil {
			collectionTimestamp = *model.CollectionTimestamp
		}

		log.Info("Metrics set for slow query: QueryID: %s, QueryText: %s, Database: %s, Schema: %s, ExecutionCount: %d, AvgCPUTimeMS: %f, AvgElapsedTimeMS: %f, AvgDiskReads: %f, AvgDiskWrites: %f, StatementType: %s, CollectionTimestamp: %s",
			queryID,
			queryText,
			databaseName,
			schemaName,
			executionCount,
			avgCPUTimeMS,
			avgElapsedTimeMS,
			avgDiskReads,
			avgDiskWrites,
			statementType,
			collectionTimestamp)

		metricSet := instanceEntity.NewMetricSet("PostgresSlowQueriesGo",
			attribute.Attribute{Key: "queryID", Value: queryID},
			attribute.Attribute{Key: "databaseName", Value: databaseName},
			attribute.Attribute{Key: "queryText", Value: queryText},
		)

		// Add all the fields to the metric set.
		if model.SchemaName != nil {
			metricSet.SetMetric("schemaName", *model.SchemaName, metric.ATTRIBUTE)
		}
		if model.ExecutionCount != nil {
			metricSet.SetMetric("executionCount", *model.ExecutionCount, metric.ATTRIBUTE)
		}
		if model.AvgCPUTimeMS != nil {
			metricSet.SetMetric("avgCPUTimeMS", *model.AvgCPUTimeMS, metric.GAUGE)
		}
		if model.AvgElapsedTimeMS != nil {
			metricSet.SetMetric("avgElapsedTimeMS", *model.AvgElapsedTimeMS, metric.GAUGE)
		}
		if model.AvgDiskReads != nil {
			metricSet.SetMetric("avgDiskReads", *model.AvgDiskReads, metric.GAUGE)
		}
		if model.AvgDiskWrites != nil {
			metricSet.SetMetric("avgDiskWrites", *model.AvgDiskWrites, metric.GAUGE)
		}
		if model.StatementType != nil {
			metricSet.SetMetric("statementType", *model.StatementType, metric.ATTRIBUTE)
		}
		if model.CollectionTimestamp != nil {
			metricSet.SetMetric("collectionTimestamp", *model.CollectionTimestamp, metric.GAUGE)
		}
		if model.QueryText != nil {
			metricSet.SetMetric("queryText", *model.QueryText, metric.ATTRIBUTE)
		}

		log.Info("Metrics set for slow query: %s in database: %s", queryID, databaseName)
	}

	log.Info("Completed processing all slow query entries.")
}
