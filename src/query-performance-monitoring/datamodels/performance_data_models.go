package datamodels

type SlowRunningQueryMetrics struct {
	Newrelic            *string  `db:"newrelic"              metric_name:"newrelic"                   source_type:"attribute"  ingest_data:"false"`
	QueryID             *string  `db:"query_id"              metric_name:"query_id"                   source_type:"attribute"`
	QueryText           *string  `db:"query_text"            metric_name:"query_text"                 source_type:"attribute"`
	DatabaseName        *string  `db:"database_name"         metric_name:"database_name"              source_type:"attribute"`
	SchemaName          *string  `db:"schema_name"           metric_name:"schema_name"                source_type:"attribute"`
	ExecutionCount      *int64   `db:"execution_count"       metric_name:"execution_count"            source_type:"gauge"`
	AvgElapsedTimeMs    *float64 `db:"avg_elapsed_time_ms"   metric_name:"avg_elapsed_time_ms"        source_type:"gauge"`
	AvgDiskReads        *float64 `db:"avg_disk_reads"        metric_name:"avg_disk_reads"             source_type:"gauge"`
	AvgDiskWrites       *float64 `db:"avg_disk_writes"       metric_name:"avg_disk_writes"            source_type:"gauge"`
	StatementType       *string  `db:"statement_type"        metric_name:"statement_type"             source_type:"attribute"`
	CollectionTimestamp *string  `db:"collection_timestamp"  metric_name:"collection_timestamp"       source_type:"attribute"`
}
type WaitEventMetrics struct {
	WaitEventName       *string  `db:"wait_event_name"       metric_name:"wait_event_name"            source_type:"attribute"`
	WaitCategory        *string  `db:"wait_category"         metric_name:"wait_category"              source_type:"attribute"`
	TotalWaitTimeMs     *float64 `db:"total_wait_time_ms"    metric_name:"total_wait_time_ms"         source_type:"gauge"`
	CollectionTimestamp *string  `db:"collection_timestamp"  metric_name:"collection_timestamp"       source_type:"attribute"`
	QueryID             *string  `db:"query_id"              metric_name:"query_id"                   source_type:"attribute"`
	QueryText           *string  `db:"query_text"            metric_name:"query_text"                 source_type:"attribute"`
	DatabaseName        *string  `db:"database_name"         metric_name:"database_name"              source_type:"attribute"`
}
type BlockingSessionMetrics struct {
	Newrelic           *string `db:"newrelic"              metric_name:"newrelic"            source_type:"attribute"  ingest_data:"false"`
	BlockedPid         *int64  `db:"blocked_pid"          metric_name:"blocked_pid"          source_type:"gauge"`
	BlockedQuery       *string `db:"blocked_query"        metric_name:"blocked_query"        source_type:"attribute"`
	BlockedQueryID     *string `db:"blocked_query_id"     metric_name:"blocked_query_id"     source_type:"attribute"`
	BlockedQueryStart  *string `db:"blocked_query_start"  metric_name:"blocked_query_start"  source_type:"attribute"`
	BlockedDatabase    *string `db:"database_name"        metric_name:"database_name"        source_type:"attribute"`
	BlockingPid        *int64  `db:"blocking_pid"         metric_name:"blocking_pid"         source_type:"gauge"`
	BlockingQuery      *string `db:"blocking_query"       metric_name:"blocking_query"       source_type:"attribute"`
	BlockingQueryID    *string `db:"blocking_query_id"    metric_name:"blocking_query_id"    source_type:"attribute"`
	BlockingQueryStart *string `db:"blocking_query_start" metric_name:"blocking_query_start" source_type:"attribute"`
}

type IndividualQueryMetrics struct {
	QueryText       *string  `json:"query" db:"query" metric_name:"query_text" source_type:"attribute"`
	QueryID         *string  `json:"queryid" db:"queryid" metric_name:"query_id" source_type:"attribute"`
	DatabaseName    *string  `json:"datname" db:"datname" metric_name:"database_name" source_type:"attribute"`
	AvgCPUTimeInMS  *float64 `json:"cpu_time_ms" db:"cpu_time_ms" metric_name:"cpu_time_ms" source_type:"gauge"`
	PlanID          *string  `json:"planid" db:"planid" metric_name:"plan_id" source_type:"attribute"`
	RealQueryText   *string  `ingest_data:"false"`
	AvgExecTimeInMs *float64 `json:"exec_time_ms" db:"exec_time_ms" metric_name:"exec_time_ms" source_type:"gauge"`
	Newrelic        *string  `db:"newrelic"              metric_name:"newrelic"            source_type:"attribute"  ingest_data:"false"`
}

type QueryExecutionPlanMetrics struct {
	NodeType            string  `mapstructure:"Node Type"           json:"Node Type"           metric_name:"node_type"             source_type:"attribute"`
	ParallelAware       bool    `mapstructure:"Parallel Aware"      json:"Parallel Aware"      metric_name:"parallel_aware"       source_type:"gauge"`
	AsyncCapable        bool    `mapstructure:"Async Capable"       json:"Async Capable"       metric_name:"async_capable"        source_type:"gauge"`
	ScanDirection       string  `mapstructure:"Scan Direction"      json:"Scan Direction"      metric_name:"scan_direction"       source_type:"attribute"`
	IndexName           string  `mapstructure:"Index Name"          json:"Index Name"          metric_name:"index_name"           source_type:"attribute"`
	RelationName        string  `mapstructure:"Relation Name"       json:"Relation Name"       metric_name:"relation_name"        source_type:"attribute"`
	Alias               string  `mapstructure:"Alias"               json:"Alias"               metric_name:"alias"                source_type:"attribute"`
	StartupCost         float64 `mapstructure:"Startup Cost"        json:"Startup Cost"        metric_name:"startup_cost"         source_type:"gauge"`
	TotalCost           float64 `mapstructure:"Total Cost"          json:"Total Cost"          metric_name:"total_cost"           source_type:"gauge"`
	PlanRows            int64   `mapstructure:"Plan Rows"           json:"Plan Rows"           metric_name:"plan_rows"            source_type:"gauge"`
	PlanWidth           int64   `mapstructure:"Plan Width"          json:"Plan Width"          metric_name:"plan_width"           source_type:"gauge"`
	RowsRemovedByFilter int64   `mapstructure:"Rows Removed by Filter" json:"Rows Removed by Filter" metric_name:"rows_removed_by_filter" source_type:"gauge"`
	DatabaseName        string  `mapstructure:"Database"            json:"Database"            metric_name:"database_name"        source_type:"attribute"`
	QueryID             string  `mapstructure:"Query Id"            json:"Query Id"            metric_name:"query_id"             source_type:"attribute"`
	PlanID              string  `mapstructure:"Plan Id"             json:"Plan Id"             metric_name:"plan_id"              source_type:"attribute"`
	Level               int     `mapstructure:"Level"               json:"Level"               metric_name:"level_id"             source_type:"gauge"`
}
