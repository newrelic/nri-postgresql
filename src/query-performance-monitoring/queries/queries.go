// Package queries contains the collection methods to parse and build the collection schema
package queries

const (
	SlowQueriesForV13AndAbove = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		pss.queryid AS query_id, -- Unique identifier for the query
		LEFT(pss.query, 4095) AS query_text, -- Query text truncated to 4095 characters
		pd.datname AS database_name, -- Name of the database
		current_schema() AS schema_name, -- Name of the current schema
		pss.calls AS execution_count, -- Number of times the query was executed
		ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, -- Average execution time in milliseconds
		pss.shared_blks_read / pss.calls AS avg_disk_reads, -- Average number of disk reads per execution
		pss.shared_blks_written / pss.calls AS avg_disk_writes, -- Average number of disk writes per execution
		CASE
			WHEN pss.query ILIKE 'SELECT%%' THEN 'SELECT' -- Query type is SELECT
			WHEN pss.query ILIKE 'INSERT%%' THEN 'INSERT' -- Query type is INSERT
			WHEN pss.query ILIKE 'UPDATE%%' THEN 'UPDATE' -- Query type is UPDATE
			WHEN pss.query ILIKE 'DELETE%%' THEN 'DELETE' -- Query type is DELETE
			ELSE 'OTHER' -- Query type is OTHER
		END AS statement_type, -- Type of SQL statement
		to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp -- Timestamp of data collection
	FROM
		pg_stat_statements pss
	JOIN
		pg_database pd ON pss.dbid = pd.oid
	WHERE 
		pd.datname in (%s) -- List of database names
		AND pss.query NOT ILIKE 'EXPLAIN (FORMAT JSON)%%' -- Exclude EXPLAIN queries
		AND pss.query NOT ILIKE 'SELECT $1 as newrelic%%' -- Exclude specific New Relic queries
		AND pss.query NOT ILIKE 'WITH wait_history AS%%' -- Exclude specific WITH queries
		AND pss.query NOT ILIKE 'select -- BLOATQUERY%%' -- Exclude BLOATQUERY
		AND pss.query NOT ILIKE 'select -- INDEXQUERY%%' -- Exclude INDEXQUERY
		AND pss.query NOT ILIKE 'SELECT -- TABLEQUERY%%' -- Exclude TABLEQUERY
		AND pss.query NOT ILIKE 'SELECT table_schema%%' -- Exclude table_schema queries
	ORDER BY
		avg_elapsed_time_ms DESC -- Order by the average elapsed time in descending order
	LIMIT %d;`

	// SlowQueriesForV12 retrieves slow queries and their statistics for PostgreSQL version 12
	SlowQueriesForV12 = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		pss.queryid AS query_id, -- Unique identifier for the query
		LEFT(pss.query, 4095) AS query_text, -- Query text truncated to 4095 characters
		pd.datname AS database_name, -- Name of the database
		current_schema() AS schema_name, -- Name of the current schema
		pss.calls AS execution_count, -- Number of times the query was executed
		ROUND((pss.total_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, -- Average execution time in milliseconds
		pss.shared_blks_read / pss.calls AS avg_disk_reads, -- Average number of disk reads per execution
		pss.shared_blks_written / pss.calls AS avg_disk_writes, -- Average number of disk writes per execution
		CASE
		  WHEN pss.query ILIKE 'SELECT%%' THEN 'SELECT' -- Query type is SELECT
		  WHEN pss.query ILIKE 'INSERT%%' THEN 'INSERT' -- Query type is INSERT
		  WHEN pss.query ILIKE 'UPDATE%%' THEN 'UPDATE' -- Query type is UPDATE
		  WHEN pss.query ILIKE 'DELETE%%' THEN 'DELETE' -- Query type is DELETE
		  ELSE 'OTHER' -- Query type is OTHER
		END AS statement_type, -- Type of SQL statement
		to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp -- Timestamp of data collection
	FROM
		pg_stat_statements pss
	JOIN
		pg_database pd ON pss.dbid = pd.oid
		WHERE 
		pd.datname in (%s) -- List of database names
		AND pss.query NOT ILIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		AND pss.query NOT ILIKE 'SELECT $1 as newrelic%%' -- Exclude specific New Relic queries
		AND pss.query NOT ILIKE 'WITH wait_history AS%%' -- Exclude specific WITH queries
		AND pss.query NOT ILIKE 'select -- BLOATQUERY%%' -- Exclude BLOATQUERY
		AND pss.query NOT ILIKE 'select -- INDEXQUERY%%' -- Exclude INDEXQUERY
		AND pss.query NOT ILIKE 'SELECT -- TABLEQUERY%%' -- Exclude TABLEQUERY
		AND pss.query NOT ILIKE 'SELECT table_schema%%' -- Exclude table_schema queries
		AND pss.query NOT ILIKE 'SELECT D.datname%%' -- Exclude specific datname queries
	ORDER BY
		avg_elapsed_time_ms DESC -- Order by the average elapsed time in descending order
	LIMIT
		 %d; -- Limit the number of results`

	// WaitEvents retrieves wait events and their statistics from pg_wait_sampling_history
	WaitEvents = `WITH wait_history AS (
		SELECT
			wh.pid, -- Process ID
			wh.event_type, -- Type of the wait event
			wh.event, -- Wait event
			wh.ts, -- Timestamp of the wait event
			pg_database.datname AS database_name, -- Name of the database
			LEAD(wh.ts) OVER (PARTITION BY wh.pid ORDER BY wh.ts) - wh.ts AS duration, -- Duration of the wait event
			LEFT(sa.query, 4095) AS query_text, -- Query text truncated to 4095 characters
			sa.queryid AS query_id -- Unique identifier for the query
		FROM
			pg_wait_sampling_history wh
		LEFT JOIN
			pg_stat_statements sa ON wh.queryid = sa.queryid
		LEFT JOIN
			pg_database ON pg_database.oid = sa.dbid
		WHERE pg_database.datname in (%s) -- List of database names
	)
	SELECT
		event_type || ':' || event AS wait_event_name, -- Concatenated wait event name
		CASE
			WHEN event_type IN ('LWLock', 'Lock') THEN 'Locks' -- Wait category is Locks
			WHEN event_type = 'IO' THEN 'Disk IO' -- Wait category is Disk IO
			WHEN event_type = 'CPU' THEN 'CPU' -- Wait category is CPU
			ELSE 'Other' -- Wait category is Other
		END AS wait_category, -- Category of the wait event
		EXTRACT(EPOCH FROM SUM(duration)) * 1000 AS total_wait_time_ms, -- Convert duration to milliseconds
		to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp, -- Timestamp of data collection
		query_id, -- Unique identifier for the query
		query_text, -- Query text
		database_name -- Name of the database
	FROM wait_history
	WHERE query_text NOT LIKE 'EXPLAIN (FORMAT JSON) %%' AND query_id IS NOT NULL AND event_type IS NOT NULL
	GROUP BY event_type, event, query_id, query_text, database_name
	ORDER BY total_wait_time_ms DESC -- Order by the total wait time in descending order
	LIMIT %d; -- Limit the number of results`

	// WaitEvents retrieves wait events and their statistics from pg_stat_activity
	WaitEventsFromPgStatActivity = `WITH wait_history AS (
        SELECT
            sa.pid, -- Process ID
            sa.wait_event_type AS event_type, -- Type of the wait event
            sa.wait_event AS event, -- Wait event           
            sa.backend_start AS duration, -- Timestamp of the wait event
            pg_database.datname AS database_name, -- Name of the database
			sa.query as query_text
        FROM
            pg_stat_activity sa
        LEFT JOIN
            pg_database ON pg_database.oid = sa.datid
        WHERE pg_database.datname in (%s) -- List of database names 
      )
    SELECT
        event_type || ':' || event AS wait_event_name, -- Concatenated wait event name
        CASE
            WHEN event_type IN ('LWLock', 'Lock') THEN 'Locks' -- Wait category is Locks
            WHEN event_type = 'IO' THEN 'Disk IO' -- Wait category is Disk IO
            WHEN event_type = 'CPU' THEN 'CPU' -- Wait category is CPU
            ELSE 'Other' -- Wait category is Other
        END AS wait_category, -- Category of the wait event
		query_text,
        to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp, -- Timestamp of data collection
        database_name -- Name of the database
    FROM wait_history
    WHERE query_text NOT LIKE 'EXPLAIN (FORMAT JSON) %%' AND event_type IS NOT NULL
    GROUP BY event_type, event, database_name,duration,query_text
    ORDER BY duration DESC -- Order by the total wait time in descending order
    LIMIT %d;  -- Limit the number of results`

	// BlockingQueriesForV14AndAbove retrieves information about blocking and blocked queries for PostgreSQL version 14 and above
	BlockingQueriesForV14AndAbove = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		  blocked_activity.pid AS blocked_pid, -- Process ID of the blocked query
		  LEFT(blocked_statements.query, 4095) AS blocked_query, -- Blocked query text truncated to 4095 characters
		  blocked_statements.queryid AS blocked_query_id, -- Unique identifier for the blocked query
		  blocked_activity.query_start AS blocked_query_start, -- Start time of the blocked query
		  blocked_activity.datname AS database_name, -- Name of the database
		  blocking_activity.pid AS blocking_pid, -- Process ID of the blocking query
		  LEFT(blocking_statements.query, 4095) AS blocking_query, -- Blocking query text truncated to 4095 characters
		  blocking_statements.queryid AS blocking_query_id, -- Unique identifier for the blocking query
		  blocking_activity.query_start AS blocking_query_start -- Start time of the blocking query
		FROM pg_stat_activity AS blocked_activity
		JOIN pg_stat_statements AS blocked_statements ON blocked_activity.query_id = blocked_statements.queryid
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
		JOIN pg_stat_statements AS blocking_statements ON blocking_activity.query_id = blocking_statements.queryid
		WHERE NOT blocked_locks.granted
		  AND blocked_activity.datname IN (%s) -- List of database names
		  AND blocked_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		  AND blocking_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		ORDER BY blocked_activity.query_start ASC -- Order by the start time of the blocked query in ascending order
		LIMIT %d; -- Limit the number of results`

	RDSPostgresBlockingQuery = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		  blocked_activity.pid AS blocked_pid, -- Process ID of the blocked query
		  blocked_activity.query AS blocked_query, -- Blocked query text truncated to 4095 characters
		  blocked_activity.query_start AS blocked_query_start, -- Start time of the blocked query
		  blocked_activity.datname AS database_name, -- Name of the database
		  blocking_activity.pid AS blocking_pid, -- Process ID of the blocking query
		  blocking_activity.query AS blocking_query, -- Blocking query text truncated to 4095 characters
		  blocking_activity.query_start AS blocking_query_start -- Start time of the blocking query
		FROM pg_stat_activity AS blocked_activity
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
		WHERE NOT blocked_locks.granted
          AND blocked_activity.datname IN (%s) -- List of database names
		  AND blocked_activity.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		  AND blocking_activity.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		ORDER BY blocked_activity.query_start ASC -- Order by the start time of the blocked query in ascending order
		LIMIT %d; -- Limit the number of results`

	// BlockingQueriesForV12AndV13 retrieves information about blocking and blocked queries for PostgreSQL versions 12 and 13
	BlockingQueriesForV12AndV13 = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		blocked_activity.pid AS blocked_pid, -- Process ID of the blocked query
		LEFT(blocked_activity.query, 4095) AS blocked_query, -- Blocked query text truncated to 4095 characters
		blocked_activity.query_start AS blocked_query_start, -- Start time of the blocked query
		blocked_activity.datname AS database_name, -- Name of the database
		blocking_activity.pid AS blocking_pid, -- Process ID of the blocking query
		LEFT(blocking_activity.query, 4095) AS blocking_query, -- Blocking query text truncated to 4095 characters
		blocking_activity.query_start AS blocking_query_start -- Start time of the blocking query
	FROM pg_stat_activity AS blocked_activity
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
	WHERE NOT blocked_locks.granted
		AND blocked_activity.datname IN (%s) -- List of database names
		AND blocked_activity.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		AND blocking_activity.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%' -- Exclude EXPLAIN queries
		ORDER BY blocked_activity.query_start ASC -- Order by the start time of the blocked query in ascending order
		LIMIT %d; -- Limit the number of results`

	// IndividualQuerySearchV13AndAbove retrieves individual query statistics for PostgreSQL version 13 and above
	IndividualQuerySearchV13AndAbove = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		 LEFT(query, 4095) as query, -- Query text truncated to 4095 characters
		 queryid, -- Unique identifier for the query
		 datname, -- Name of the database
		 planid, -- Plan identifier
		 ROUND(((cpu_user_time + cpu_sys_time) / NULLIF(calls, 0))::numeric, 3) AS cpu_time_ms, -- Average CPU time in milliseconds
		 total_exec_time / NULLIF(calls, 0) AS exec_time_ms -- Average execution time in milliseconds
		FROM
		 pg_stat_monitor
		WHERE
		 queryid = %s -- Query identifier
		 AND datname IN (%s) -- List of database names
		 AND (total_exec_time / NULLIF(calls, 0)) > %d -- Minimum average execution time
		 AND bucket_start_time >= NOW() - INTERVAL '60 seconds' -- Time interval
		GROUP BY
		 query, queryid, datname, planid, cpu_user_time, cpu_sys_time, calls, total_exec_time
		ORDER BY
		 exec_time_ms DESC -- Order by average execution time in descending order
		LIMIT %d; -- Limit the number of results`

	IndividualQueryFromPgStat = "select query  from pg_stat_activity where query is not null and query !='';"

	// IndividualQuerySearchV12 retrieves individual query statistics for PostgreSQL version 12
	IndividualQuerySearchV12 = `SELECT 'newrelic' as newrelic, -- Common value to filter with like operator in slow query metrics
		 LEFT(query, 4095) as query, -- Query text truncated to 4095 characters
		 queryid, -- Unique identifier for the query
		 datname, -- Name of the database
		 planid, -- Plan identifier
		 ROUND(((cpu_user_time + cpu_sys_time) / NULLIF(calls, 0))::numeric, 3) AS cpu_time_ms, -- Average CPU time in milliseconds
		 total_time / NULLIF(calls, 0) AS exec_time_ms -- Average execution time in milliseconds
		FROM
		 pg_stat_monitor
		WHERE
		 queryid = %s -- Query identifier
		 AND datname IN (%s) -- List of database names
		 AND (total_time / NULLIF(calls, 0)) > %d -- Minimum average execution time
		 AND bucket_start_time >= NOW() - INTERVAL '60 seconds' -- Time interval
		GROUP BY
		 query, queryid, datname, planid, cpu_user_time, cpu_sys_time, calls, total_time
		ORDER BY
		 exec_time_ms DESC -- Order by average execution time in descending order
		LIMIT %d; -- Limit the number of results`
)
