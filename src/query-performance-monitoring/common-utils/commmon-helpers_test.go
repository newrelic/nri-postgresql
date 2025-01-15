package commonutils_test

import (
	"sort"
	"testing"
	"time"

	"github.com/newrelic/nri-postgresql/src/collection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/stretchr/testify/assert"
)

func TestGetQuotedStringFromArray(t *testing.T) {
	input := []string{"db1", "db2", "db3"}
	expected := "'db1','db2','db3'"
	result := commonutils.GetQuotedStringFromArray(input)
	assert.Equal(t, expected, result)
}

func TestGetDatabaseListInString(t *testing.T) {
	dbListKeys := []string{"db1", "db2"}
	sort.Strings(dbListKeys) // Sort the keys to ensure consistent order
	dbList := collection.DatabaseList{}
	for _, key := range dbListKeys {
		dbList[key] = collection.SchemaList{}
	}
	expected := "'db1','db2'"
	result := commonutils.GetDatabaseListInString(dbList)
	assert.Equal(t, expected, result)

	// Test with empty database list
	dbList = collection.DatabaseList{}
	expected = ""
	result = commonutils.GetDatabaseListInString(dbList)
	assert.Equal(t, expected, result)
}

func TestAnonymizeQueryText(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1 AND name = 'John'"
	expected := "SELECT * FROM users WHERE id = ? AND name = ?"
	result := commonutils.AnonymizeQueryText(query)
	assert.Equal(t, expected, result)
}

func TestGeneratePlanID(t *testing.T) {
	queryID := "query123"
	result := commonutils.GeneratePlanID(queryID)
	assert.NotNil(t, result)
	assert.Contains(t, *result, queryID)
	assert.Contains(t, *result, "-")
	assert.Contains(t, *result, time.Now().Format(commonutils.TimeFormat)[:8]) // Check date part
}
