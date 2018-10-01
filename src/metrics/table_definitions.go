package metrics

import (
	"fmt"
	"strings"

	"github.com/newrelic/nri-postgresql/src/args"
)

func generateTableDefinitions(schemaList args.SchemaList) []*QueryDefinition {
	queryDefinitions := make([]*QueryDefinition, 2)
	queryDefinitions[0] = tableBloatDefinition.insertSchemaTables(schemaList)
	queryDefinitions[1] = tableDefinition.insertSchemaTables(schemaList)

	return queryDefinitions
}

func (qd *QueryDefinition) insertSchemaTables(schemaList args.SchemaList) *QueryDefinition {
	schemaTables := make([]string, 0)
	for schema, tableList := range schemaList {
		for table := range tableList {
			schemaTables = append(schemaTables, fmt.Sprintf("'%s.%s'", schema, table))
		}
	}

	schemaTablesString := strings.Join(schemaTables, ",")

	qd.query = strings.Replace(qd.query, `%SCHEMA_TABLES%`, schemaTablesString, 1)

	return qd
}

var tableDefinition = &QueryDefinition{
	query: `SELECT -- TABLEQUERY
			current_database() as database,
			stat.schemaname as schema_name,
			stat.relname as table_name,
			pg_total_relation_size(c.oid), -- table.totalSizeInBytes
			pg_indexes_size(c.oid), -- table.indexSizeInBytes
			idx_blks_read, -- table.indexBlocksRead
			idx_blks_hit, -- table.indexBlocksHit
			toast_blks_read, --table.indexToastBlocksRead
			toast_blks_hit, -- table.indexToastBlocksHit
			extract(epoch from last_vacuum)::int as last_vacuum, -- table.lastVacuum
			extract(epoch from last_autovacuum)::int as last_autovacuum, -- table.lastAutoVacuum
			extract(epoch from last_analyze)::int as last_analyze, -- table.lastAnalyze
			extract(epoch from last_autoanalyze)::int as last_autoanalyze, -- table.lastAutoAnalyze
			seq_scan, -- table.sequentialScansPerSecond
			seq_tup_read, -- table.sequentialScanRowsFetchedPerSecond
			idx_scan, -- table.indexScansPerSecond
			idx_tup_fetch, -- table.indexScanRowsFetchedPerSecond
			n_tup_ins, -- table.rowsInsertedPerSecond
			n_tup_upd, -- table.rowsUpdatedPerSecond
			n_tup_del, -- table.rowsDeletedPerSecond
			n_live_tup, -- table.liveRows
			n_dead_tup -- table.deadRows
		FROM pg_statio_user_tables as statio
		join pg_stat_user_tables as stat
		on stat.relid=statio.relid
		join pg_class c 
		on c.relname=stat.relname
		where stat.schemaname::text || '.' || stat.relname::text in (%SCHEMA_TABLES%)`,

	dataModels: []struct {
		databaseBase
		schemaBase
		tableBase
		TotalSize                *int     `db:"pg_total_relation_size" metric_name:"table.totalSizeInBytes"                   source_type:"gauge"`
		IndexSize                *int     `db:"pg_indexes_size"        metric_name:"table.indexSizeInBytes"                   source_type:"gauge"`
		LiveRows                 *int     `db:"n_live_tup"             metric_name:"table.liveRows"                           source_type:"gauge"`
		DeadRows                 *int     `db:"n_dead_tup"             metric_name:"table.deadRows"                           source_type:"gauge"`
		IndexBlocksReadPerSecond *float32 `db:"idx_blks_read"          metric_name:"table.indexBlocksReadPerSecond"           source_type:"rate"`
		IndexBlocksHitPerSecond  *float32 `db:"idx_blks_hit"           metric_name:"table.indexBlocksHitPerSecond"            source_type:"rate"`
		ToastBlocksReadPerSecond *float32 `db:"toast_blks_read"        metric_name:"table.indexToastBlocksReadPerSecond"      source_type:"rate"`
		ToastBlocksHitPerSecond  *float32 `db:"toast_blks_hit"         metric_name:"table.indexToastBlocksHitPerSecond"       source_type:"rate"`
		LastVacuum               *int     `db:"last_vacuum"            metric_name:"table.lastVacuum"                         source_type:"gauge"`
		LastAutoVacuum           *int     `db:"last_autovacuum"        metric_name:"table.lastAutoVacuum"                     source_type:"gauge"`
		LastAnalyze              *int     `db:"last_analyze"           metric_name:"table.lastAnalyze"                        source_type:"gauge"`
		LastAutoAnalyze          *int     `db:"last_autoanalyze"       metric_name:"table.lastAutoAnalyze"                    source_type:"gauge"`
		SeqScans                 *float32 `db:"seq_scan"               metric_name:"table.sequentialScansPerSecond"           source_type:"rate"`
		SeqReads                 *float32 `db:"seq_tup_read"           metric_name:"table.sequentialScanRowsFetchedPerSecond" source_type:"rate"`
		IndexScans               *float32 `db:"idx_scan"               metric_name:"table.indexScansPerSecond"                source_type:"rate"`
		IndexReads               *float32 `db:"idx_tup_fetch"          metric_name:"table.indexScanRowsFetchedPerSecond"      source_type:"rate"`
		RowsInserted             *float32 `db:"n_tup_ins"              metric_name:"table.rowsInsertedPerSecond"              source_type:"rate"`
		RowsUpdated              *float32 `db:"n_tup_upd"              metric_name:"table.rowsUpdatedPerSecond"               source_type:"rate"`
		RowsDeleted              *float32 `db:"n_tup_del"              metric_name:"table.rowsDeletedPerSecond"               source_type:"rate"`
	}{},
}

var tableBloatDefinition = &QueryDefinition{
	query: `SELECT -- BLOATQUERY
			current_database() as database, 
			schemaname as schema_name, tblname as table_name, bs*tblpages AS real_size,
			(tblpages-est_tblpages_ff)*bs AS bloat_size,
			CASE WHEN tblpages - est_tblpages_ff > 0
				THEN 100 * (tblpages - est_tblpages_ff)/tblpages::float
				ELSE 0
			END AS bloat_ratio
			-- , (pst).free_percent + (pst).dead_tuple_percent AS real_frag
		FROM (
			SELECT ceil( reltuples / ( (bs-page_hdr)/tpl_size ) ) + ceil( toasttuples / 4 ) AS est_tblpages,
				ceil( reltuples / ( (bs-page_hdr)*fillfactor/(tpl_size*100) ) ) + ceil( toasttuples / 4 ) AS est_tblpages_ff,
				tblpages, fillfactor, bs, tblid, schemaname, tblname, heappages, toastpages, is_na
				-- , stattuple.pgstattuple(tblid) AS pst
			FROM (
				SELECT
					( 4 + tpl_hdr_size + tpl_data_size + (2*ma)
						- CASE WHEN tpl_hdr_size%ma = 0 THEN ma ELSE tpl_hdr_size%ma END
						- CASE WHEN ceil(tpl_data_size)::int%ma = 0 THEN ma ELSE ceil(tpl_data_size)::int%ma END
					) AS tpl_size, bs - page_hdr AS size_per_block, (heappages + toastpages) AS tblpages, heappages,
					toastpages, reltuples, toasttuples, bs, page_hdr, tblid, schemaname, tblname, fillfactor, is_na
				FROM (
					SELECT
						tbl.oid AS tblid, ns.nspname AS schemaname, tbl.relname AS tblname, tbl.reltuples,
						tbl.relpages AS heappages, coalesce(toast.relpages, 0) AS toastpages,
						coalesce(toast.reltuples, 0) AS toasttuples,
						coalesce(substring(
							array_to_string(tbl.reloptions, ' ')
							FROM 'fillfactor=([0-9]+)')::smallint, 100) AS fillfactor,
						current_setting('block_size')::numeric AS bs,
						CASE WHEN version()~'mingw32' OR version()~'64-bit|x86_64|ppc64|ia64|amd64' THEN 8 ELSE 4 END AS ma,
						24 AS page_hdr,
						CASE WHEN current_setting('server_version_num')::integer < 80300 THEN 27 ELSE 23 END
							+ CASE WHEN MAX(coalesce(null_frac,0)) > 0 THEN ( 7 + count(*) ) / 8 ELSE 0::int END
							+ CASE WHEN tbl.relhasoids THEN 4 ELSE 0 END AS tpl_hdr_size,
						sum( (1-coalesce(s.null_frac, 0)) * coalesce(s.avg_width, 1024) ) AS tpl_data_size,
						bool_or(att.atttypid = 'pg_catalog.name'::regtype)
							OR count(att.attname) <> count(s.attname) AS is_na
					FROM pg_attribute AS att
						JOIN pg_class AS tbl ON att.attrelid = tbl.oid
						JOIN pg_namespace AS ns ON ns.oid = tbl.relnamespace
						LEFT JOIN pg_stats AS s ON s.schemaname=ns.nspname
							AND s.tablename = tbl.relname AND s.attname=att.attname
						LEFT JOIN pg_class AS toast ON tbl.reltoastrelid = toast.oid
					WHERE att.attnum > 0 AND NOT att.attisdropped
						AND tbl.relkind = 'r'
					GROUP BY 1,2,3,4,5,6,7,8,9,10, tbl.relhasoids
					ORDER BY 2,3
				) AS s
			) AS s2
		) AS s3
		where not is_na
		and schemaname || '.' || tblname in (%SCHEMA_TABLES%)`,

	dataModels: []struct {
		databaseBase
		schemaBase
		tableBase
		BloatSize  *float64 `db:"bloat_size" metric_name:"table.bloatSizeInBytes" source_type:"gauge"`
		RealSize   *float64 `db:"real_size" metric_name:"table.dataSizeInBytes" source_type:"gauge"`
		BloatRatio *float64 `db:"bloat_ratio" metric_name:"table.bloatRatio" source_type:"gauge"`
	}{},
}
