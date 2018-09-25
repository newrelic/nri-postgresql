package metrics

import (
	"reflect"
)

type QueryDefinition struct {
	query      string
	dataModels interface{}
}

func (qd QueryDefinition) GetQuery() string {
	return qd.query
}

func (qd QueryDefinition) GetDataModels() interface{} {
	ptr := reflect.New(reflect.ValueOf(qd.dataModels).Elem().Type())
	return ptr.Interface()
}
