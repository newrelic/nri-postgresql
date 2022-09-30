package collection

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

const (
	allDBQuery    = `SELECT datname FROM pg_database WHERE datistemplate = false;`
	dbSchemaQuery = `SELECT table_schema AS schema_name, t1.table_name AS table_name, t2.indexname AS index_name 
                     FROM information_schema.tables AS t1
                     FULL OUTER JOIN pg_indexes t2
                     ON t2.tablename = t1.table_name
                     AND t2.schemaname = t1.table_schema;`
)

// DatabaseList is a map from database name to SchemaLists to collect
type DatabaseList map[string]SchemaList

// SchemaList is a map from schema name to TableList to collect
type SchemaList map[string]TableList

// TableList is a map from table name to an array of indexes to collect
type TableList map[string][]string

type databaseIgnoreList map[string]struct{}

// BuildCollectionList unmarshals the collection_list from the args and builds the list of
// objects to be collected. If collection_list is a JSON array, it collects every object in
// each of the databases listed in the array. If it is a hash, it collects only the objects
// listed
func BuildCollectionList(al args.ArgumentList, ci connection.Info) (DatabaseList, error) {
	var dbList DatabaseList
	var dbNames []string
	var err error

	ignoreDBList, err := parseIgnoreList(al.CollectionIgnoreDatabaseList)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ignore db list: %w", err)
	}

	switch {
	case strings.ToLower(al.CollectionList) == "all":
		if dbNames, err = getAllDatabaseNames(ci); err != nil {
			return nil, fmt.Errorf("failed to get all databases names: %w", err)
		}

	case nil == json.Unmarshal([]byte(al.CollectionList), &dbList):
		for ignoredDB := range ignoreDBList {
			delete(dbList, ignoredDB)
		}

	case nil == json.Unmarshal([]byte(al.CollectionList), &dbNames):
	default:
		return nil, errors.New("failed to parse collection list")
	}

	if len(dbNames) != 0 {
		if dbList, err = buildCollectionListFromDatabaseNames(dbNames, ignoreDBList, ci); err != nil {
			return nil, err
		}
	}

	return dbList, nil
}

func parseIgnoreList(list string) (databaseIgnoreList, error) {
	ignoreDBList := []string{}
	ignoreDBMap := databaseIgnoreList{}

	if list == "" {
		return ignoreDBMap, nil
	}

	if err := json.Unmarshal([]byte(list), &ignoreDBList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list arg '%s': %w", list, err)
	}

	for _, db := range ignoreDBList {
		ignoreDBMap[db] = struct{}{}
	}

	return ignoreDBMap, nil
}

func getAllDatabaseNames(ci connection.Info) ([]string, error) {
	con, err := ci.NewConnection(ci.DatabaseName())
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var dataModel []struct {
		DatabaseName sql.NullString `db:"datname"`
	}
	err = con.Query(&dataModel, allDBQuery)
	if err != nil {
		return nil, err
	}

	databaseNames := make([]string, 0, len(dataModel))
	for _, database := range dataModel {
		if database.DatabaseName.Valid {
			databaseNames = append(databaseNames, database.DatabaseName.String)
		}
	}

	return databaseNames, nil
}

func buildCollectionListFromDatabaseNames(dbnames []string, ignoreDBList databaseIgnoreList, ci connection.Info) (DatabaseList, error) {
	databaseList := DatabaseList{}
	for _, db := range dbnames {
		if _, ok := ignoreDBList[db]; ok {
			continue
		}

		con, err := ci.NewConnection(db)
		if err != nil {
			log.Error("Failed to open connection to database '%s' to build collection list: %s", db, err)
			continue
		}
		defer con.Close()

		schemaList, err := buildSchemaListForDatabase(con)
		if err != nil {
			log.Error("Failed to build schema list for database '%s': %s", db, err)
			continue
		}

		databaseList[db] = schemaList
	}
	if len(databaseList) == 0 {
		return nil, fmt.Errorf("no database to collect data")
	}

	return databaseList, nil
}

func buildSchemaListForDatabase(con *connection.PGSQLConnection) (SchemaList, error) {
	schemaList := make(SchemaList)

	var dataModel []struct {
		SchemaName sql.NullString `db:"schema_name"`
		TableName  sql.NullString `db:"table_name"`
		IndexName  sql.NullString `db:"index_name"`
	}
	err := con.Query(&dataModel, dbSchemaQuery)
	if err != nil {
		return nil, err
	}

	for index, row := range dataModel {
		if !row.SchemaName.Valid || !row.TableName.Valid {
			if row.IndexName.Valid {
				log.Debug("Skipping Index %s. Schema name or Table name null", row.IndexName.String)
			} else {
				log.Debug("Query responded with a null schema name or table name. Skipping row %d", index)
			}
			continue
		}

		if _, ok := schemaList[row.SchemaName.String]; !ok {
			schemaList[row.SchemaName.String] = make(TableList)
		}

		if _, ok := schemaList[row.SchemaName.String][row.TableName.String]; !ok {
			schemaList[row.SchemaName.String][row.TableName.String] = make([]string, 0)
		}

		if row.IndexName.Valid {
			schemaList[row.SchemaName.String][row.TableName.String] = append(schemaList[row.SchemaName.String][row.TableName.String], row.IndexName.String)
		}
	}

	return schemaList, nil
}
