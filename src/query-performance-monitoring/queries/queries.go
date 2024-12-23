// Package queries contains the collection methods to parse and build the collection schema
package queries

const (
	SlowQueries = `SELECT
        pss.queryid AS query_id,
        pss.query AS query_text,
        pd.datname AS database_name,
        current_schema() AS schema_name,
        pss.calls AS execution_count,
        ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms,
        ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms,
        pss.shared_blks_read / pss.calls AS avg_disk_reads,
        pss.shared_blks_written / pss.calls AS avg_disk_writes,
        CASE
            WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT'
            WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT'
            WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE'
            WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE'
            ELSE 'OTHER'
        END AS statement_type,
        to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp
    FROM
        pg_stat_statements pss
    JOIN
        pg_database pd ON pss.dbid = pd.oid
    WHERE 
        pss.query NOT LIKE 'EXPLAIN (FORMAT JSON) %'   
    	AND pss.query LIKE 'SELECT e.name %'



    ORDER BY
        avg_elapsed_time_ms DESC -- Order by the average elapsed time in descending order
    LIMIT
        20;`

	WaitEvents = `WITH wait_history AS (
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
    WHERE query_text NOT LIKE 'EXPLAIN (FORMAT JSON) %' AND query_id IS NOT NULL AND event_type IS NOT NULL
    GROUP BY event_type, event, query_id, query_text, database_name
    ORDER BY total_wait_time_ms DESC
    LIMIT 10;`

	BlockingQueries = `SELECT
          blocked_activity.pid AS blocked_pid,
          blocked_statements.query AS blocked_query,
          blocked_statements.queryid AS blocked_query_id,
          blocked_activity.query_start AS blocked_query_start,
          blocked_activity.datname AS database_name,
          blocking_activity.pid AS blocking_pid,
          blocking_statements.query AS blocking_query,
          blocking_statements.queryid AS blocking_query_id,
          blocking_activity.query_start AS blocking_query_start
      FROM pg_stat_activity AS blocked_activity
      JOIN pg_stat_statements as blocked_statements on blocked_activity.query_id = blocked_statements.queryid
      JOIN pg_locks blocked_locks ON blocked_activity.pid = blocked_locks.pid
      JOIN pg_locks blocking_locks ON blocked_locks.locktype = blocking_locks.locktype
          AND blocked_locks.database IS NOT DISTINCT FROM blocking_locks.database
          AND blocked_locks.relation IS NOT DISTINCT FROM blocking_locks.relation
          AND blocked_locks.page IS NOT DISTINCT FROM blocking_locks.page
          AND blocked_locks.tuple IS NOT DISTINCT FROM blocking_locks.tuple
          AND blocked_locks.transactionid IS NOT DISTINCT FROM blocking_locks.transactionid
          AND blocked_locks.classid IS NOT DISTINCT FROM blocking_locks.classid
          AND blocked_locks.objid IS NOT DISTINCT FROM blocking_locks.objid
          AND blocked_locks.objsubid IS NOT DISTINCT FROM blocking_locks.objsubid
          AND blocked_locks.pid <> blocking_locks.pid
      JOIN pg_stat_activity AS blocking_activity ON blocking_locks.pid = blocking_activity.pid
      JOIN pg_stat_statements as blocking_statements on blocking_activity.query_id = blocking_statements.queryid
      WHERE NOT blocked_locks.granted
          AND blocked_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %'
          AND blocking_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %'
      LIMIT 10;
`
	//	IndividualQuerySearch = `SELECT query, queryid, datname,planid,
	//							ROUND((cpu_user_time + cpu_sys_time) / NULLIF(total_calls, 0), 3) AS avg_cpu_time_ms
	//							FROM pg_stat_monitor
	//							WHERE queryid IN (%s)
	//-- 							AND bucket_start_time >= NOW() - INTERVAL '15 seconds';`

	IndividualQuerySearch = `SELECT
			query,
			queryid,
			datname,
			planid,
			ROUND(((cpu_user_time + cpu_sys_time) / NULLIF(calls, 0))::numeric, 3) AS avg_cpu_time_ms
			FROM
				pg_stat_monitor
			WHERE
				queryid IN (%s)
			GROUP BY
				query, queryid, datname, planid, cpu_user_time, cpu_sys_time, calls`
)
