package metrics

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/newrelic/nri-postgresql/src/collection"
)

// QueryDefinition holds the query and the unmarshall model
type QueryDefinition struct {
	query      string
	dataModels interface{}
}

// GetQuery returns the query of the QueryDefinition
func (qd QueryDefinition) GetQuery() string {
	return qd.query
}

// GetDataModels returns the data models of the QueryDefinition
func (qd QueryDefinition) GetDataModels() interface{} {
	ptr := reflect.New(reflect.ValueOf(qd.dataModels).Type())
	return ptr.Interface()
}

func (qd *QueryDefinition) insertDatabaseNames(databases collection.DatabaseList) *QueryDefinition {
	databaseList := ""
	for database := range databases {
		databaseList += `'` + database + `',`
	}
	databaseList = databaseList[0 : len(databaseList)-1]

	qd.query = strings.Replace(qd.query, `%DATABASES%`, databaseList, 1)

	return qd
}

func (qd *QueryDefinition) insertSchemaTables(schemaList collection.SchemaList) *QueryDefinition {
	schemaTables := make([]string, 0)
	for schema, tableList := range schemaList {
		for table := range tableList {
			schemaTables = append(schemaTables, fmt.Sprintf("'%s.%s'", schema, table))
		}
	}

	if len(schemaTables) == 0 {
		return nil
	}

	schemaTablesString := strings.Join(schemaTables, ",")

	newTableDef := &QueryDefinition{
		dataModels: qd.dataModels,
		query:      strings.Replace(qd.query, `%SCHEMA_TABLES%`, schemaTablesString, 1),
	}

	return newTableDef
}

func (qd *QueryDefinition) insertSchemaTableIndexes(schemaList collection.SchemaList) *QueryDefinition {
	schemaTableIndexes := make([]string, 0)
	for schema, tableList := range schemaList {
		for table, indexList := range tableList {
			for _, index := range indexList {
				schemaTableIndexes = append(schemaTableIndexes, fmt.Sprintf("'%s.%s.%s'", schema, table, index))
			}
		}
	}

	if len(schemaTableIndexes) == 0 {
		return nil
	}

	schemaTableIndexString := strings.Join(schemaTableIndexes, ",")

	newIndexDef := &QueryDefinition{
		dataModels: qd.dataModels,
		query:      strings.Replace(qd.query, `%SCHEMA_TABLE_INDEXES%`, schemaTableIndexString, 1),
	}

	return newIndexDef
}
