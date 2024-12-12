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
	QueryText    *string `json:"query" db:"query" metric_name:"queryplan.query" source_type:"attribute"`
	QueryId      *int64  `json:"queryid" db:"queryid" metric_name:"queryplan.query_id" source_type:"gauge"`
	DatabaseName *string `json:"datname" db:"datname" metric_name:"queryplan.database" source_type:"attribute"`
}

//type QueryExecutionPlanMetrics struct {
//	NodeType            string  `json:"Node Type"           metric_name:"executionplan.node_type"             source_type:"attribute"`
//	StartupCost         float64 `json:"Startup Cost"        metric_name:"executionplan.startup_cost"          source_type:"gauge"`
//	TotalCost           float64 `json:"Total Cost"          metric_name:"executionplan.total_cost"            source_type:"gauge"`
//	PlanRows            int64   `json:"Plan Rows"           metric_name:"executionplan.plan_rows"             source_type:"gauge"`
//	ActualStartupTime   float64 `json:"Actual Startup Time" metric_name:"executionplan.actual_startup_time"   source_type:"gauge"`
//	ActualTotalTime     float64 `json:"Actual Total Time"   metric_name:"executionplan.actual_total_time"     source_type:"gauge"`
//	ActualRows          int64   `json:"Actual Rows"         metric_name:"executionplan.actual_rows"           source_type:"gauge"`
//	ActualLoops         int64   `json:"Actual Loops"        metric_name:"executionplan.actual_loops"          source_type:"gauge"`
//	SharedHitBlocks     int64   `json:"Shared Hit Blocks"   metric_name:"executionplan.shared_hit_blocks"     source_type:"gauge"`
//	SharedReadBlocks    int64   `json:"Shared Read Blocks"  metric_name:"executionplan.shared_read_blocks"    source_type:"gauge"`
//	SharedDirtiedBlocks int64   `json:"Shared Dirtied Blocks" metric_name:"executionplan.shared_dirtied_blocks" source_type:"gauge"`
//	SharedWrittenBlocks int64   `json:"Shared Written Blocks" metric_name:"executionplan.shared_written_blocks" source_type:"gauge"`
//	LocalHitBlocks      int64   `json:"Local Hit Blocks"    metric_name:"executionplan.local_hit_blocks"      source_type:"gauge"`
//	LocalReadBlocks     int64   `json:"Local Read Blocks"   metric_name:"executionplan.local_read_blocks"     source_type:"gauge"`
//	LocalDirtiedBlocks  int64   `json:"Local Dirtied Blocks" metric_name:"executionplan.local_dirtied_blocks"  source_type:"gauge"`
//	LocalWrittenBlocks  int64   `json:"Local Written Blocks" metric_name:"executionplan.local_written_blocks"  source_type:"gauge"`
//	TempReadBlocks      int64   `json:"Temp Read Blocks"    metric_name:"executionplan.temp_read_blocks"      source_type:"gauge"`
//	TempWrittenBlocks   int64   `json:"Temp Written Blocks" metric_name:"executionplan.temp_written_blocks"   source_type:"gauge"`
//	DatabaseName        string  `json:"Database"            metric_name:"executionplan.database"              source_type:"attribute"`
//	QueryText           string  `json:"Query"               metric_name:"executionplan.query"                 source_type:"attribute"`
//	QueryId             int64   `json:"Query Id"              metric_name:"executionplan.query_id"                   source_type:"gauge"`
//}

type QueryExecutionPlanMetrics struct {
	NodeType            string  `mapstructure:"Node Type"           json:"Node Type"           metric_name:"executionplan.node_type"             source_type:"attribute"`
	StartupCost         float64 `mapstructure:"Startup Cost"        json:"Startup Cost"        metric_name:"executionplan.startup_cost"          source_type:"gauge"`
	TotalCost           float64 `mapstructure:"Total Cost"          json:"Total Cost"          metric_name:"executionplan.total_cost"            source_type:"gauge"`
	PlanRows            int64   `mapstructure:"Plan Rows"           json:"Plan Rows"           metric_name:"executionplan.plan_rows"             source_type:"gauge"`
	ActualStartupTime   float64 `mapstructure:"Actual Startup Time" json:"Actual Startup Time" metric_name:"executionplan.actual_startup_time"   source_type:"gauge"`
	ActualTotalTime     float64 `mapstructure:"Actual Total Time"   json:"Actual Total Time"   metric_name:"executionplan.actual_total_time"     source_type:"gauge"`
	ActualRows          int64   `mapstructure:"Actual Rows"         json:"Actual Rows"         metric_name:"executionplan.actual_rows"           source_type:"gauge"`
	ActualLoops         int64   `mapstructure:"Actual Loops"        json:"Actual Loops"        metric_name:"executionplan.actual_loops"          source_type:"gauge"`
	SharedHitBlocks     int64   `mapstructure:"Shared Hit Blocks"   json:"Shared Hit Blocks"   metric_name:"executionplan.shared_hit_blocks"     source_type:"gauge"`
	SharedReadBlocks    int64   `mapstructure:"Shared Read Blocks"  json:"Shared Read Blocks"  metric_name:"executionplan.shared_read_blocks"    source_type:"gauge"`
	SharedDirtiedBlocks int64   `mapstructure:"Shared Dirtied Blocks" json:"Shared Dirtied Blocks" metric_name:"executionplan.shared_dirtied_blocks" source_type:"gauge"`
	SharedWrittenBlocks int64   `mapstructure:"Shared Written Blocks" json:"Shared Written Blocks" metric_name:"executionplan.shared_written_blocks" source_type:"gauge"`
	LocalHitBlocks      int64   `mapstructure:"Local Hit Blocks"    json:"Local Hit Blocks"    metric_name:"executionplan.local_hit_blocks"      source_type:"gauge"`
	LocalReadBlocks     int64   `mapstructure:"Local Read Blocks"   json:"Local Read Blocks"   metric_name:"executionplan.local_read_blocks"     source_type:"gauge"`
	LocalDirtiedBlocks  int64   `mapstructure:"Local Dirtied Blocks" json:"Local Dirtied Blocks" metric_name:"executionplan.local_dirtied_blocks"  source_type:"gauge"`
	LocalWrittenBlocks  int64   `mapstructure:"Local Written Blocks" json:"Local Written Blocks" metric_name:"executionplan.local_written_blocks"  source_type:"gauge"`
	TempReadBlocks      int64   `mapstructure:"Temp Read Blocks"    json:"Temp Read Blocks"    metric_name:"executionplan.temp_read_blocks"      source_type:"gauge"`
	TempWrittenBlocks   int64   `mapstructure:"Temp Written Blocks" json:"Temp Written Blocks" metric_name:"executionplan.temp_written_blocks"   source_type:"gauge"`
	DatabaseName        string  `mapstructure:"Database"            json:"Database"            metric_name:"executionplan.database"              source_type:"attribute"`
	QueryText           string  `mapstructure:"Query"               json:"Query"               metric_name:"executionplan.query"                 source_type:"attribute"`
	QueryId             int64   `mapstructure:"Query Id"            json:"Query Id"            metric_name:"executionplan.query_id"              source_type:"gauge"`
}
