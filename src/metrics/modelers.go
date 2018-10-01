package metrics

import (
	"errors"
	"reflect"
)

// DatabaseModeler is an interface to represent somethign which has a database field
type DatabaseModeler interface {
	GetDatabaseName() (string, error)
}

type databaseBase struct {
	Database *string `db:"database"`
}

// GetDatabaseName returns the database name for the object
func (d databaseBase) GetDatabaseName() (string, error) {
	if d.Database == nil {
		return "", errors.New("database name not returned")
	}
	return *d.Database, nil
}

// GetDatabaseName returns the database name for the object
func GetDatabaseName(dataModel interface{}) (string, error) {
	v := reflect.ValueOf(dataModel)
	modeler, ok := v.Interface().(DatabaseModeler)
	if !ok {
		return "", errors.New("data model does not implement DatabaseModeler interface")
	}

	name, err := modeler.GetDatabaseName()
	if err != nil {
		return "", err
	}

	return name, nil
}

// SchemaModeler is an interface to represent something which has a schema field
type SchemaModeler interface {
	GetSchemaName() (string, error)
}

type schemaBase struct {
	Schema *string `db:"schema_name"`
}

// GetSchemaName returns a schema name
func (d schemaBase) GetSchemaName() (string, error) {
	if d.Schema == nil {
		return "", errors.New("schema name not returned")
	}
	return *d.Schema, nil
}

// GetSchemaName returns a schema name
func GetSchemaName(dataModel interface{}) (string, error) {
	v := reflect.ValueOf(dataModel)
	modeler, ok := v.Interface().(SchemaModeler)
	if !ok {
		return "", errors.New("data model does not implement SchemaModeler interface")
	}

	name, err := modeler.GetSchemaName()
	if err != nil {
		return "", err
	}

	return name, nil
}

// TableModeler is an interface to represent something which has a table field
type TableModeler interface {
	GetTableName() (string, error)
}

type tableBase struct {
	Table *string `db:"table_name"`
}

// GetTableName returns the table name
func (d tableBase) GetTableName() (string, error) {
	if d.Table == nil {
		return "", errors.New("table name not returned")
	}
	return *d.Table, nil
}

// GetTableName returns the table name
func GetTableName(dataModel interface{}) (string, error) {
	v := reflect.ValueOf(dataModel)
	modeler, ok := v.Interface().(TableModeler)
	if !ok {
		return "", errors.New("data model does not implement TableModeler interface")
	}

	name, err := modeler.GetTableName()
	if err != nil {
		return "", err
	}

	return name, nil
}

// IndexModeler represents something with a table field
type IndexModeler interface {
	GetIndexName() (string, error)
}

type indexBase struct {
	Index *string `db:"index_name"`
}

// GetIndexName returns the index name
func (d indexBase) GetIndexName() (string, error) {
	if d.Index == nil {
		return "", errors.New("index name not returned")
	}
	return *d.Index, nil
}

// GetIndexName returns the index name
func GetIndexName(dataModel interface{}) (string, error) {
	v := reflect.ValueOf(dataModel)
	modeler, ok := v.Interface().(IndexModeler)
	if !ok {
		return "", errors.New("data model does not implement IndexModeler interface")
	}

	name, err := modeler.GetIndexName()
	if err != nil {
		return "", err
	}

	return name, nil
}
