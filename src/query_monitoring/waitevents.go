package query_monitoring

import (
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

// WaitEventQueryDetails struct to hold query results.
type waitEventQueryDetails struct {
	QueryID             *string  `db:"query_id"`
	QueryText           *string  `db:"query_text"`
	DatabaseName        *string  `db:"database_name"`
	WaitEventName       *string  `db:"wait_event_name"`
	WaitCategory        *string  `db:"wait_category"`
	TotalWaitTimeMS     *float64 `db:"total_wait_time_ms"`
	WaitingTasksCount   *int64   `db:"waiting_tasks_count"`
	CollectionTimestamp *string  `db:"collection_timestamp"`
}

// AnalyzeWaitQueries analyzes the wait events
func WaitEventQueries(instanceEntity *integration.Entity, connection *connection.PGSQLConnection, arguments args.ArgumentList) {
	log.Info("Querying PostgreSQL Server for wait event queries")

	var getWaitEventQueryDetailsQuery = `WITH wait_history AS (
        SELECT 
            wh.pid,
            wh.event_type,
            wh.event,
            wh.ts,
            pg_database.datname AS database_name,
            LEAD(wh.ts) OVER (PARTITION BY wh.pid ORDER BY wh.ts) - wh.ts AS duration,
            sa.query AS query_text,
            sa.queryid AS query_id
        FROM
            pg_wait_sampling_history wh
        LEFT JOIN
            pg_stat_statements sa ON wh.queryid = sa.queryid
        LEFT JOIN
            pg_database ON pg_database.oid = sa.dbid
      )
      SELECT
        event_type || ':' || event AS wait_event_name,
        CASE
            WHEN event_type IN ('LWLock', 'Lock') THEN 'Locks'
            WHEN event_type = 'IO' THEN 'Disk IO'
            WHEN event_type = 'CPU' THEN 'CPU'
            ELSE 'Other'
        END AS wait_category,
        EXTRACT(EPOCH FROM SUM(duration)) * 1000 AS total_wait_time_ms,  -- Convert duration to milliseconds
        COUNT(*) AS waiting_tasks_count,
        to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp,
        query_id,
        query_text,
        database_name
      FROM wait_history
      WHERE duration IS NOT NULL and query_id IS NOT NULL and event_type IS NOT NULL
      GROUP BY event_type, event, query_id, query_text, database_name
      ORDER BY total_wait_time_ms DESC
      LIMIT 10;`

	log.Info("Executing query to get wait event query details.")

	// Slice to hold query results.
	waitQueryModels := make([]waitEventQueryDetails, 0)

	// Execute the query and store the results in the waitQueryModels slice.
	if err := connection.Query(&waitQueryModels, getWaitEventQueryDetailsQuery); err != nil {
		log.Error("Could not execute query: %s", err.Error())
		return
	}
	log.Info("Number of records retrieved: %d", len(waitQueryModels))
	//for rows.Next() {
	//	var model waitEventQueryDetails
	//	if err := rows.StructScan(&model); err != nil {
	//		log.Error("Could not scan row: %s", err.Error())
	//		continue
	//	}
	//	waitQueryModels = append(waitQueryModels, model)
	//}

	for _, model := range waitQueryModels {
		if model.DatabaseName == nil || model.QueryText == nil || model.QueryID == nil {
			log.Warn("Skipping entry with nil field: DatabaseName, QueryText, or QueryID")
			continue // Skip this entry if any critical field is nil
		}

		queryID := *model.QueryID
		databaseName := *model.DatabaseName
		queryText := *model.QueryText
		waitEventName := ""
		if model.WaitEventName != nil {
			waitEventName = *model.WaitEventName
		}
		//fmt.Println("====================================")
		//fmt.Println(waitEventName)
		//fmt.Println("====================================")
		waitCategory := ""
		if model.WaitCategory != nil {
			waitCategory = *model.WaitCategory
		}

		totalWaitTimeMS := float64(0)
		if model.TotalWaitTimeMS != nil {
			totalWaitTimeMS = *model.TotalWaitTimeMS
		}

		waitingTasksCount := int64(0)
		if model.WaitingTasksCount != nil {
			waitingTasksCount = *model.WaitingTasksCount
		}

		collectionTimestamp := ""
		if model.CollectionTimestamp != nil {
			collectionTimestamp = *model.CollectionTimestamp
		}

		log.Info("Metrics set for wait query: QueryID: %s, QueryText: %s, DatabaseName: %s, WaitEventName: %s, WaitCategory: %s, TotalWaitTimeMS: %f, WaitingTasksCount: %d, CollectionTimestamp: %s",
			queryID,
			queryText,
			databaseName,
			waitEventName,
			waitCategory,
			totalWaitTimeMS,
			waitingTasksCount,
			collectionTimestamp)

		metricSet := instanceEntity.NewMetricSet("PostgreSQLWaitEventQueriesGo",
			attribute.Attribute{Key: "queryID", Value: queryID},
			attribute.Attribute{Key: "databaseName", Value: databaseName},
			attribute.Attribute{Key: "queryText", Value: queryText},
		)

		// Add all the fields to the metric set.
		if model.WaitEventName != nil {
			metricSet.SetMetric("waitEventName", *model.WaitEventName, metric.ATTRIBUTE)
		}
		if model.WaitCategory != nil {
			metricSet.SetMetric("waitCategory", *model.WaitCategory, metric.ATTRIBUTE)
		}
		if model.TotalWaitTimeMS != nil {
			metricSet.SetMetric("totalWaitTimeMS", *model.TotalWaitTimeMS, metric.GAUGE)
		}
		if model.WaitingTasksCount != nil {
			metricSet.SetMetric("waitingTasksCount", *model.WaitingTasksCount, metric.GAUGE)
		}
		if model.CollectionTimestamp != nil {
			metricSet.SetMetric("collectionTimestamp", *model.CollectionTimestamp, metric.ATTRIBUTE)
		}

		log.Info("Metrics set for wait query: %s in database: %s", queryID, databaseName)
	}

	log.Info("Completed processing all wait query entries.")
}
