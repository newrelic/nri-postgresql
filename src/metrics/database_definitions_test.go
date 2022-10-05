package metrics

import (
	"regexp"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/stretchr/testify/assert"
)

func Test_generateDatabaseDefinitions_LengthV8(t *testing.T) {
	v8 := semver.MustParse("8.0.0")

	databaseList := collection.DatabaseList{"test1": {}}

	queryDefinitions := generateDatabaseDefinitions(databaseList, &v8)

	assert.Equal(t, 1, len(queryDefinitions))
}

func Test_generateDatabaseDefinitions_LengthV912(t *testing.T) {
	v912 := semver.MustParse("9.1.2")
	databaseList := collection.DatabaseList{"test1": {}}

	queryDefinitions := generateDatabaseDefinitions(databaseList, &v912)

	assert.Equal(t, 1, len(queryDefinitions))
}

func Test_generateDatabaseDefinitions_LengthV925(t *testing.T) {
	v925 := semver.MustParse("9.2.5")
	databaseList := collection.DatabaseList{"test1": {}}

	queryDefinitions := generateDatabaseDefinitions(databaseList, &v925)

	assert.Equal(t, 2, len(queryDefinitions))
}

func Test_insertDatabaseNames(t *testing.T) {
	testDefinition := &QueryDefinition{
		query:      `SELECT * FROM test WHERE database IN (%DATABASES%);`,
		dataModels: &[]struct{}{},
	}

	databaseList := collection.DatabaseList{"test1": {}, "test2": {}}
	td := testDefinition.insertDatabaseNames(databaseList)

	expectedRegexp := `SELECT \* FROM test WHERE database IN \('test[12]','test[22]'\);`

	assert.Regexp(t, regexp.MustCompile(expectedRegexp), td.query)
}
