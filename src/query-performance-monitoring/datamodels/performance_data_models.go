package datamodels

type SlowRunningQueryMetrics struct {
	QueryID             *int64   `db:"query_id"              metric_name:"query_id"                   source_type:"gauge"`
	QueryText           *string  `db:"query_text"            metric_name:"query_text"                 source_type:"attribute"`
	DatabaseName        *string  `db:"database_name"         metric_name:"database_name"              source_type:"attribute"`
	SchemaName          *string  `db:"schema_name"           metric_name:"schema_name"                source_type:"attribute"`
	ExecutionCount      *int64   `db:"execution_count"       metric_name:"execution_count"            source_type:"gauge"`
	AvgElapsedTimeMs    *float64 `db:"avg_elapsed_time_ms"   metric_name:"avg_elapsed_time_ms"        source_type:"gauge"`
	AvgCPUTimeMs        *float64 `db:"avg_cpu_time_ms"       metric_name:"avg_cpu_time_ms"            source_type:"gauge"`
	AvgDiskReads        *float64 `db:"avg_disk_reads"        metric_name:"avg_disk_reads"             source_type:"gauge"`
	AvgDiskWrites       *float64 `db:"avg_disk_writes"       metric_name:"avg_disk_writes"            source_type:"gauge"`
	StatementType       *string  `db:"statement_type"        metric_name:"statement_type"             source_type:"attribute"`
	CollectionTimestamp *string  `db:"collection_timestamp"  metric_name:"collection_timestamp"       source_type:"attribute"`
}
type WaitEventMetrics struct {
	WaitEventName       *string  `db:"wait_event_name"       metric_name:"wait_event_name"            source_type:"attribute"`
	WaitCategory        *string  `db:"wait_category"         metric_name:"wait_category"              source_type:"attribute"`
	TotalWaitTimeMs     *float64 `db:"total_wait_time_ms"    metric_name:"total_wait_time_ms"         source_type:"gauge"`
	WaitingTasksCount   *int64   `db:"waiting_tasks_count"   metric_name:"waiting_tasks_count"        source_type:"gauge"`
	CollectionTimestamp *string  `db:"collection_timestamp"  metric_name:"collection_timestamp"       source_type:"attribute"`
	QueryID             *string  `db:"query_id"              metric_name:"query_id"                   source_type:"attribute"`
	QueryText           *string  `db:"query_text"            metric_name:"query_text"                 source_type:"attribute"`
	DatabaseName        *string  `db:"database_name"         metric_name:"database_name"              source_type:"attribute"`
}
type BlockingSessionMetrics struct {
	BlockedPid         *int64  `db:"blocked_pid"          metric_name:"blocked_pid"          source_type:"gauge"`
	BlockedQuery       *string `db:"blocked_query"        metric_name:"blocked_query"        source_type:"attribute"`
	BlockedQueryId     *string `db:"blocked_query_id"     metric_name:"blocked_query_id"     source_type:"attribute"`
	BlockedQueryStart  *string `db:"blocked_query_start"  metric_name:"blocked_query_start"  source_type:"attribute"`
	BlockedDatabase    *string `db:"database_name"        metric_name:"database_name"        source_type:"attribute"`
	BlockingPid        *int64  `db:"blocking_pid"         metric_name:"blocking_pid"         source_type:"gauge"`
	BlockingQuery      *string `db:"blocking_query"       metric_name:"blocking_query"       source_type:"attribute"`
	BlockingQueryId    *string `db:"blocking_query_id"    metric_name:"blocking_query_id"    source_type:"attribute"`
	BlockingQueryStart *string `db:"blocking_query_start" metric_name:"blocking_query_start" source_type:"attribute"`
}

type IndividualQueryMetrics struct {
	QueryText    *string `json:"query" db:"query" metric_name:"query" source_type:"attribute"`
	QueryId      *int64  `json:"queryid" db:"queryid" metric_name:"query_id" source_type:"gauge"`
	DatabaseName *string `json:"datname" db:"datname" metric_name:"database" source_type:"attribute"`
	AvgTimeInMS  *string `json:"avg_cpu_time_ms" db:"avg_cpu_time_ms" metric_name:"avg_cpu_time_ms" source_type:"gauge"`
	PlanId       *int64  `json:"planid" db:"planid" metric_name:"plan_id" source_type:"gauge"`
}

type QueryExecutionPlanMetrics struct {
	NodeType            string  `mapstructure:"Node Type"           json:"Node Type"           metric_name:"node_type"             source_type:"attribute"`
	StartupCost         float64 `mapstructure:"Startup Cost"        json:"Startup Cost"        metric_name:"startup_cost"          source_type:"gauge"`
	TotalCost           float64 `mapstructure:"Total Cost"          json:"Total Cost"          metric_name:"total_cost"            source_type:"gauge"`
	PlanRows            int64   `mapstructure:"Plan Rows"           json:"Plan Rows"           metric_name:"plan_rows"             source_type:"gauge"`
	ActualStartupTime   float64 `mapstructure:"Actual Startup Time" json:"Actual Startup Time" metric_name:"actual_startup_time"   source_type:"gauge"`
	ActualTotalTime     float64 `mapstructure:"Actual Total Time"   json:"Actual Total Time"   metric_name:"actual_total_time"     source_type:"gauge"`
	ActualRows          int64   `mapstructure:"Actual Rows"         json:"Actual Rows"         metric_name:"actual_rows"           source_type:"gauge"`
	ActualLoops         int64   `mapstructure:"Actual Loops"        json:"Actual Loops"        metric_name:"actual_loops"          source_type:"gauge"`
	SharedHitBlocks     int64   `mapstructure:"Shared Hit Blocks"   json:"Shared Hit Blocks"   metric_name:"shared_hit_blocks"     source_type:"gauge"`
	SharedReadBlocks    int64   `mapstructure:"Shared Read Blocks"  json:"Shared Read Blocks"  metric_name:"shared_read_blocks"    source_type:"gauge"`
	SharedDirtiedBlocks int64   `mapstructure:"Shared Dirtied Blocks" json:"Shared Dirtied Blocks" metric_name:"shared_dirtied_blocks" source_type:"gauge"`
	SharedWrittenBlocks int64   `mapstructure:"Shared Written Blocks" json:"Shared Written Blocks" metric_name:"shared_written_blocks" source_type:"gauge"`
	LocalHitBlocks      int64   `mapstructure:"Local Hit Blocks"    json:"Local Hit Blocks"    metric_name:"local_hit_blocks"      source_type:"gauge"`
	LocalReadBlocks     int64   `mapstructure:"Local Read Blocks"   json:"Local Read Blocks"   metric_name:"local_read_blocks"     source_type:"gauge"`
	LocalDirtiedBlocks  int64   `mapstructure:"Local Dirtied Blocks" json:"Local Dirtied Blocks" metric_name:"local_dirtied_blocks"  source_type:"gauge"`
	LocalWrittenBlocks  int64   `mapstructure:"Local Written Blocks" json:"Local Written Blocks" metric_name:"local_written_blocks"  source_type:"gauge"`
	TempReadBlocks      int64   `mapstructure:"Temp Read Blocks"    json:"Temp Read Blocks"    metric_name:"temp_read_blocks"      source_type:"gauge"`
	TempWrittenBlocks   int64   `mapstructure:"Temp Written Blocks" json:"Temp Written Blocks" metric_name:"temp_written_blocks"   source_type:"gauge"`
	DatabaseName        string  `mapstructure:"Database"            json:"Database"            metric_name:"database"              source_type:"attribute"`
	QueryText           string  `mapstructure:"Query"               json:"Query"               metric_name:"query"                 source_type:"attribute"`
	QueryId             int64   `mapstructure:"Query Id"            json:"Query Id"            metric_name:"query_id"              source_type:"gauge"`
}
