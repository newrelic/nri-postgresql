package metrics

import (
	"reflect"
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
