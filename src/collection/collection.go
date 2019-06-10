package collection

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

// DatabaseList is a map from database name to SchemaLists to collect
type DatabaseList map[string]SchemaList

// SchemaList is a map from schema name to TableList to collect
type SchemaList map[string]TableList

// TableList is a map from table name to an array of indexes to collect
type TableList map[string][]string

// BuildCollectionList unmarshals the collection_list from the args and builds the list of
// objects to be collected. If collection_list is a JSON array, it collects every object in
// each of the databases listed in the array. If it is a hash, it collects only the objects
// listed
func BuildCollectionList(al args.ArgumentList, ci connection.Info) (DatabaseList, error) {
	var dl DatabaseList
	if err := json.Unmarshal([]byte(al.CollectionList), &dl); err == nil {
		return dl, nil
	}

	var dbnames []string
	if err := json.Unmarshal([]byte(al.CollectionList), &dbnames); err == nil {
		return buildCollectionListFromDatabaseNames(dbnames, ci)
	}

	return nil, errors.New("failed to parse collection list")
}

func buildCollectionListFromDatabaseNames(dbnames []string, ci connection.Info) (DatabaseList, error) {
	databaseList := make(DatabaseList)
	for _, db := range dbnames {
		con, err := ci.NewConnection(db)
		if err != nil {
			return nil, err
		}

		schemaList, err := buildSchemaListForDatabase(db, con)
		if err != nil {
			return nil, err
		}

		databaseList[db] = schemaList
	}

	return databaseList, nil
}

func buildSchemaListForDatabase(dbname string, con *connection.PGSQLConnection) (SchemaList, error) {
	schemaList := make(SchemaList)

	query := `select 
      table_schema as schema_name,
      t1.table_name as table_name,
      t2.indexname as index_name
    from information_schema.tables as t1
    full outer join pg_indexes t2 
      on t2.tablename = t1.table_name
      and t2.schemaname = t1.table_schema;`

	var dataModel []struct {
		SchemaName sql.NullString `db:"schema_name"`
		TableName  sql.NullString `db:"table_name"`
		IndexName  sql.NullString `db:"index_name"`
	}
	err := con.Query(&dataModel, query)
	if err != nil {
		return nil, err
	}

	for index, row := range dataModel {
		if !row.SchemaName.Valid || !row.TableName.Valid {
			log.Error("Query responded with a null schema name or table name. Skipping row %d", index)
			continue
		}

		if _, ok := schemaList[row.SchemaName.String]; !ok {
			schemaList[row.SchemaName.String] = make(TableList)
		}

		if _, ok := schemaList[row.TableName.String]; !ok {
			schemaList[row.SchemaName.String][row.TableName.String] = make([]string, 0)
		}

		if row.IndexName.Valid {
			schemaList[row.SchemaName.String][row.TableName.String] = append(schemaList[row.SchemaName.String][row.TableName.String], row.IndexName.String)
		}
	}

	return schemaList, nil
}
