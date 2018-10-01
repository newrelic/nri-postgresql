package metrics

import (
	"fmt"
	"strings"

	"github.com/newrelic/nri-postgresql/src/args"
)

func generateIndexDefinitions(schemaList args.SchemaList) *QueryDefinition {
	return indexDefinition.insertSchemaTableIndexes(schemaList)
}

func (qd *QueryDefinition) insertSchemaTableIndexes(schemaList args.SchemaList) *QueryDefinition {
	schemaTableIndexes := make([]string, 0)
	for schema, tableList := range schemaList {
		for table, indexList := range tableList {
			for _, index := range indexList {
				schemaTableIndexes = append(schemaTableIndexes, fmt.Sprintf("'%s.%s.%s'", schema, table, index))
			}
		}
	}

	schemaTableIndexString := strings.Join(schemaTableIndexes, ",")

	qd.query = strings.Replace(qd.query, `%SCHEMA_TABLE_INDEXES%`, schemaTableIndexString, 1)

	return qd
}

var indexDefinition = &QueryDefinition{
	query: `select -- INDEXQUERY
				current_database() as database,
				t.schemaname as schema_name,
					t.tablename as table_name,
					indexname as index_name,
					pg_relation_size(quote_ident(indexrelname)::text) AS index_size,
					idx_tup_read AS tuples_read,
					idx_tup_fetch AS tuples_fetched
			FROM pg_tables t
			LEFT OUTER JOIN pg_class c ON t.tablename=c.relname
			LEFT OUTER JOIN
					( SELECT c.relname AS ctablename, ipg.relname AS indexname, x.indnatts AS number_of_columns, idx_scan, idx_tup_read, idx_tup_fetch, indexrelname, indisunique FROM pg_index x
								 JOIN pg_class c ON c.oid = x.indrelid
								 JOIN pg_class ipg ON ipg.oid = x.indexrelid
								 JOIN pg_stat_all_indexes psai ON x.indexrelid = psai.indexrelid )
					AS foo
					ON t.tablename = foo.ctablename
			where indexname is not null and t.schemaname || '.' || t.tablename || '.' || indexname in (%SCHEMA_TABLE_INDEXES%)
			ORDER BY 1,2;`,

	dataModels: []struct {
		databaseBase
		schemaBase
		tableBase
		indexBase
		IndexSize   *int `db:"index_size"     metric_name:"index.sizeInBytes"          source_type:"gauge"`
		RowsRead    *int `db:"tuples_read"    metric_name:"index.rowsReadPerSecond"    source_type:"rate"`
		RowsFetched *int `db:"tuples_fetched" metric_name:"index.rowsFetchedPerSecond" source_type:"rate"`
	}{},
}
