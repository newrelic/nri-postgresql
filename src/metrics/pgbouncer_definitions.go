package metrics

func generatePgBouncerDefinitions() []*QueryDefinition {
	queryDefinitions := make([]*QueryDefinition, 2)
	queryDefinitions[0] = pgbouncerStatsDefinition
	queryDefinitions[1] = pgbouncerPoolsDefinition

	return queryDefinitions
}

var pgbouncerStatsDefinition = &QueryDefinition{
	query: `SHOW STATS;`,

	dataModels: []struct {
		databaseBase
		TotalXactCount  *int64 `db:"total_xact_count"  metric_name:"pgbouncer.stats.transactionsPerSecond"                           source_type:"rate"`
		TotalQueryCount *int64 `db:"total_query_count" metric_name:"pgbouncer.stats.queriesPerSecond"                                source_type:"rate"`
		TotalReceived   *int64 `db:"total_received"    metric_name:"pgbouncer.stats.bytesInPerSecond"                                source_type:"rate"`
		TotalSent       *int64 `db:"total_sent"        metric_name:"pgbouncer.stats.bytesOutPerSecond"                               source_type:"rate"`
		TotalXactTime   *int64 `db:"total_xact_time"   metric_name:"pgbouncer.stats.totalTransactionDurationInMillisecondsPerSecond" source_type:"rate"`
		TotalQueryTime  *int64 `db:"total_query_time"  metric_name:"pgbouncer.stats.totalQueryDurationInMillisecondsPerSecond"       source_type:"rate"`
		TotalRequests   *int64 `db:"total_requests"    metric_name:"pgbouncer.stats.requestsPerSecond"                               source_type:"rate"`
		TotalWaitTime   *int64 `db:"total_wait_time"`
		AvgXactCount    *int64 `db:"avg_xact_count"    metric_name:"pgbouncer.stats.avgTransactionCount"                             source_type:"gauge"`
		AvgXactTime     *int64 `db:"avg_xact_time"     metric_name:"pgbouncer.stats.avgTransactionDurationInMilliseconds"            source_type:"gauge"`
		AvgQueryCount   *int64 `db:"avg_query_count"   metric_name:"pgbouncer.stats.avgQueryCount"                                   source_type:"gauge"`
		AvgRecv         *int64 `db:"avg_recv"          metric_name:"pgbouncer.stats.avgBytesIn"                                      source_type:"gauge"`
		AvgSent         *int64 `db:"avg_sent"          metric_name:"pgbouncer.stats.avgBytesOut"                                     source_type:"gauge"`
		AvgReq          *int64 `db:"avg_req"           metric_name:"pgbouncer.stats.avgRequestsPerSecond"                            source_type:"gauge"`
		AvgQueryTime    *int64 `db:"avg_query_time"    metric_name:"pgbouncer.stats.avgQueryDurationInMilliseconds"                  source_type:"gauge"`
		AvgQuery        *int64 `db:"avg_query"         metric_name:"pgbouncer.stats.avgQueryDurationInMilliseconds"                  source_type:"gauge"`
		AvgWaitTime     *int64 `db:"avg_wait_time"`
	}{},
}

var pgbouncerPoolsDefinition = &QueryDefinition{
	query: `SHOW POOLS;`,

	dataModels: []struct {
		databaseBase
		User               *string `db:"user"`
		ClCancelReq        *int64  `db:"cl_cancel_req"         metric_name:"pgbouncer.pools.clientConnectionsCancelReq"        source_type:"gauge"` //removed in v1.18
		ClActive           *int64  `db:"cl_active"             metric_name:"pgbouncer.pools.clientConnectionsActive"           source_type:"gauge"`
		ClWaiting          *int64  `db:"cl_waiting"            metric_name:"pgbouncer.pools.clientConnectionsWaiting"          source_type:"gauge"`
		ClWaitingCancelReq *int64  `db:"cl_waiting_cancel_req" metric_name:"pgbouncer.pools.clientConnectionsWaitingCancelReq" source_type:"gauge"` // added in v1.18
		ClActiveCancelReq  *int64  `db:"cl_active_cancel_req"  metric_name:"pgbouncer.pools.clientConnectionsActiveCancelReq"  source_type:"gauge"` // added in v1.18
		SvActiveCancel     *int64  `db:"sv_active_cancel"      metric_name:"pgbouncer.pools.serverConnectionsActiveCancel"     source_type:"gauge"` // added in v1.18
		SvBeingCancel      *int64  `db:"sv_being_canceled"     metric_name:"pgbouncer.pools.serverConnectionsBeingCancel"      source_type:"gauge"` // added in v1.18
		SvActive           *int64  `db:"sv_active"             metric_name:"pgbouncer.pools.serverConnectionsActive"           source_type:"gauge"`
		SvIdle             *int64  `db:"sv_idle"               metric_name:"pgbouncer.pools.serverConnectionsIdle"             source_type:"gauge"`
		SvUsed             *int64  `db:"sv_used"               metric_name:"pgbouncer.pools.serverConnectionsUsed"             source_type:"gauge"`
		SvTested           *int64  `db:"sv_tested"             metric_name:"pgbouncer.pools.serverConnectionsTested"           source_type:"gauge"`
		SvLogin            *int64  `db:"sv_login"              metric_name:"pgbouncer.pools.serverConnectionsLogin"            source_type:"gauge"`
		MaxWait            *int64  `db:"maxwait"               metric_name:"pgbouncer.pools.maxwaitInMilliseconds"             source_type:"gauge"`
		MaxWaitUs          *int64  `db:"maxwait_us"`
		PoolMode           *string `db:"pool_mode"`
	}{},
}
