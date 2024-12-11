package datamodels

type SlowRunningQuery struct {
	QueryID             *int64   `db:"query_id"              metric_name:"slowquery.query_id"                   source_type:"gauge"`
	QueryText           *string  `db:"query_text"            metric_name:"slowquery.query_text"                 source_type:"attribute"`
	DatabaseName        *string  `db:"database_name"         metric_name:"slowquery.database_name"              source_type:"attribute"`
	SchemaName          *string  `db:"schema_name"           metric_name:"slowquery.schema_name"                source_type:"attribute"`
	ExecutionCount      *int64   `db:"execution_count"       metric_name:"slowquery.execution_count"            source_type:"gauge"`
	AvgElapsedTimeMs    *float64 `db:"avg_elapsed_time_ms"   metric_name:"slowquery.avg_elapsed_time_ms"        source_type:"gauge"`
	AvgCPUTimeMs        *float64 `db:"avg_cpu_time_ms"       metric_name:"slowquery.avg_cpu_time_ms"            source_type:"gauge"`
	AvgDiskReads        *float64 `db:"avg_disk_reads"        metric_name:"slowquery.avg_disk_reads"             source_type:"gauge"`
	AvgDiskWrites       *float64 `db:"avg_disk_writes"       metric_name:"slowquery.avg_disk_writes"            source_type:"gauge"`
	StatementType       *string  `db:"statement_type"        metric_name:"slowquery.statement_type"             source_type:"attribute"`
	CollectionTimestamp *string  `db:"collection_timestamp"  metric_name:"slowquery.collection_timestamp"       source_type:"attribute"`
}
type WaitEventQuery struct {
	WaitEventName       *string  `db:"wait_event_name"       metric_name:"waitevent.wait_event_name"            source_type:"attribute"`
	WaitCategory        *string  `db:"wait_category"         metric_name:"waitevent.wait_category"              source_type:"attribute"`
	TotalWaitTimeMs     *float64 `db:"total_wait_time_ms"    metric_name:"waitevent.total_wait_time_ms"         source_type:"gauge"`
	WaitingTasksCount   *int64   `db:"waiting_tasks_count"   metric_name:"waitevent.waiting_tasks_count"        source_type:"gauge"`
	CollectionTimestamp *string  `db:"collection_timestamp"  metric_name:"waitevent.collection_timestamp"       source_type:"attribute"`
	QueryID             *string  `db:"query_id"              metric_name:"waitevent.query_id"                   source_type:"attribute"`
	QueryText           *string  `db:"query_text"            metric_name:"waitevent.query_text"                 source_type:"attribute"`
	DatabaseName        *string  `db:"database_name"         metric_name:"waitevent.database_name"              source_type:"attribute"`
}
type BlockingQuery struct {
	BlockedPid         *int64  `db:"blocked_pid"          metric_name:"blockingquery.blocked_pid"          source_type:"gauge"`
	BlockedQuery       *string `db:"blocked_query"        metric_name:"blockingquery.blocked_query"        source_type:"attribute"`
	BlockedQueryId     *string `db:"blocked_query_id"     metric_name:"blockingquery.blocked_query_id"     source_type:"attribute"`
	BlockedQueryStart  *string `db:"blocked_query_start"  metric_name:"blockingquery.blocked_query_start"  source_type:"attribute"`
	BlockedDatabase    *string `db:"database_name"        metric_name:"blockingquery.database_name"        source_type:"attribute"`
	BlockingPid        *int64  `db:"blocking_pid"         metric_name:"blockingquery.blocking_pid"         source_type:"gauge"`
	BlockingQuery      *string `db:"blocking_query"       metric_name:"blockingquery.blocking_query"       source_type:"attribute"`
	BlockingQueryId    *string `db:"blocking_query_id"    metric_name:"blockingquery.blocking_query_id"    source_type:"attribute"`
	BlockingQueryStart *string `db:"blocking_query_start" metric_name:"blockingquery.blocking_query_start" source_type:"attribute"`
}

type IndividualQuerySearch struct {
	QueryText *string `json:"query" db:"query" metric_name:"queryplan.query" source_type:"attribute"`
}

type QueryExecutionPlanMetrics struct {
	PlanRows *float64 `metric_name:"executionplan.plan_rows"             source_type:"gauge"`
}
