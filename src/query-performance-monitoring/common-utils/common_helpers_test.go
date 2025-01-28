package commonutils_test

import (
	"sort"
	"testing"
	"time"

	"github.com/newrelic/nri-postgresql/src/collection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/stretchr/testify/assert"
)

func TestGetDatabaseListInString(t *testing.T) {
	dbListKeys := []string{"db1"}
	sort.Strings(dbListKeys) // Sort the keys to ensure consistent order
	dbList := collection.DatabaseList{}
	for _, key := range dbListKeys {
		dbList[key] = collection.SchemaList{}
	}
	expected := "'db1'"
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
	query = "SELECT * FROM employees WHERE id = 10 OR name <> 'John Doe'   OR name != 'John Doe'   OR age < 30 OR age <= 30   OR salary > 50000OR salary >= 50000  OR department LIKE 'Sales%' OR department ILIKE 'sales%'OR join_date BETWEEN '2023-01-01' AND '2023-12-31' OR department IN ('HR', 'Engineering', 'Marketing') OR department IS NOT NULL OR department IS NULL;"
	expected = "SELECT * FROM employees WHERE id = ? OR name <> ?   OR name != ?   OR age < ? OR age <= ?   OR salary > ?OR salary >= ?  OR department LIKE ? OR department ILIKE ?OR join_date BETWEEN ? AND ? OR department IN (?, ?, ?) OR department IS NOT NULL OR department IS NULL;"
	result = commonutils.AnonymizeQueryText(query)
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
