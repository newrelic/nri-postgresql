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
		TotalXactCount  *int `db:"total_xact_count"  metric_name:"pgbouncer.stats.transactionsPerSecond"                           source_type:"rate"`
		TotalQueryCount *int `db:"total_query_count" metric_name:"pgbouncer.stats.queriesPerSecond"                                source_type:"rate"`
		TotalReceived   *int `db:"total_received"    metric_name:"pgbouncer.stats.bytesInPerSecond"                                source_type:"rate"`
		TotalSent       *int `db:"total_sent"        metric_name:"pgbouncer.stats.bytesOutPerSecond"                               source_type:"rate"`
		TotalXactTime   *int `db:"total_xact_time"   metric_name:"pgbouncer.stats.totalTransactionDurationInMillisecondsPerSecond" source_type:"rate"`
		TotalQueryTime  *int `db:"total_query_time"  metric_name:"pgbouncer.stats.totalQueryDurationInMillisecondsPerSecond"       source_type:"rate"`
		TotalRequests   *int `db:"total_requests"    metric_name:"pgbouncer.stats.requestsPerSecond"                               source_type:"rate"`
		TotalWaitTime   *int `db:"total_wait_time"`
		AvgXactCount    *int `db:"avg_xact_count"    metric_name:"pgbouncer.stats.avgTransactionCount"                             source_type:"gauge"`
		AvgXactTime     *int `db:"avg_xact_time"     metric_name:"pgbouncer.stats.avgTransactionDurationInMilliseconds"            source_type:"gauge"`
		AvgQueryCount   *int `db:"avg_query_count"   metric_name:"pgbouncer.stats.avgQueryCount"                                   source_type:"gauge"`
		AvgRecv         *int `db:"avg_recv"          metric_name:"pgbouncer.stats.avgBytesIn"                                      source_type:"gauge"`
		AvgSent         *int `db:"avg_sent"          metric_name:"pgbouncer.stats.avgBytesOut"                                     source_type:"gauge"`
		AvgReq          *int `db:"avg_req"           metric_name:"pgbouncer.stats.avgRequestsPerSecond"                            source_type:"gauge"`
		AvgQueryTime    *int `db:"avg_query_time"    metric_name:"pgbouncer.stats.avgQueryDurationInMilliseconds"                  source_type:"gauge"`
		AvgQuery        *int `db:"avg_query"         metric_name:"pgbouncer.stats.avgQueryDurationInMilliseconds"                  source_type:"gauge"`
		AvgWaitTime     *int `db:"avg_wait_time"`
	}{},
}

var pgbouncerPoolsDefinition = &QueryDefinition{
	query: `SHOW POOLS;`,

	dataModels: []struct {
		databaseBase
		User      *string `db:"user"`
		ClActive  *int    `db:"cl_active"  metric_name:"pgbouncer.pools.clientConnectionsActive"  source_type:"gauge"`
		ClWaiting *int    `db:"cl_waiting" metric_name:"pgbouncer.pools.clientConnectionsWaiting" source_type:"gauge"`
		SvActive  *int    `db:"sv_active"  metric_name:"pgbouncer.pools.serverConnectionsActive"  source_type:"gauge"`
		SvIdle    *int    `db:"sv_idle"    metric_name:"pgbouncer.pools.serverConnectionsIdle"    source_type:"gauge"`
		SvUsed    *int    `db:"sv_used"    metric_name:"pgbouncer.pools.serverConnectionsUsed"    source_type:"gauge"`
		SvTested  *int    `db:"sv_tested"  metric_name:"pgbouncer.pools.serverConnectionsTested"  source_type:"gauge"`
		SvLogin   *int    `db:"sv_login"   metric_name:"pgbouncer.pools.serverConnectionsLogin"   source_type:"gauge"`
		MaxWait   *int    `db:"maxwait"    metric_name:"pgbouncer.pools.maxwaitInMilliseconds"    source_type:"gauge"`
		MaxWaitUs *int    `db:"maxwait_us"`
		PoolMode  *string `db:"pool_mode"`
	}{},
}
