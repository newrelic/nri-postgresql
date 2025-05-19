package commonutils

import (
	"sort"
	"testing"

	"github.com/newrelic/nri-postgresql/src/collection"
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
	result := GetDatabaseListInString(dbList)
	assert.Equal(t, expected, result)

	// Test with empty database list
	dbList = collection.DatabaseList{}
	expected = ""
	result = GetDatabaseListInString(dbList)
	assert.Equal(t, expected, result)
}

func TestAnonymizeQueryText(t *testing.T) {
	query := "SELECT * FROM users WHERE id = 1 AND name = 'John'"
	expected := "SELECT * FROM users WHERE id = ? AND name = ?"
	result := AnonymizeQueryText(query)
	assert.Equal(t, expected, result)
	query = "SELECT * FROM employees WHERE id = 10 OR name <> 'John Doe'   OR name != 'John Doe'   OR age < 30 OR age <= 30   OR salary > 50000OR salary >= 50000  OR department LIKE 'Sales%' OR department ILIKE 'sales%'OR join_date BETWEEN '2023-01-01' AND '2023-12-31' OR department IN ('HR', 'Engineering', 'Marketing') OR department IS NOT NULL OR department IS NULL;"
	expected = "SELECT * FROM employees WHERE id = ? OR name <> ?   OR name != ?   OR age < ? OR age <= ?   OR salary > ?OR salary >= ?  OR department LIKE ? OR department ILIKE ?OR join_date BETWEEN ? AND ? OR department IN (?, ?, ?) OR department IS NOT NULL OR department IS NULL;"
	result = AnonymizeQueryText(query)
	assert.Equal(t, expected, result)
}

func TestAnonymizeAndNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Replace numbers with placeholders",
			input:    "SELECT * FROM users WHERE id = 123",
			expected: "select * from users where id = ?",
		},
		{
			name:     "Replace single-quoted strings with placeholders",
			input:    "SELECT * FROM users WHERE name = 'John'",
			expected: "select * from users where name = ?",
		},
		{
			name:     "Replace double-quoted strings with placeholders",
			input:    `SELECT * FROM "users" WHERE "name" = "John"`,
			expected: "select * from ? where ? = ?",
		},
		{
			name:     "Remove dollar signs",
			input:    "SELECT $1 FROM users WHERE id = $2",
			expected: "select ? from users where id = ?",
		},
		{
			name:     "Convert to lowercase",
			input:    "SELECT * FROM USERS",
			expected: "select * from users",
		},
		{
			name:     "Remove semicolons",
			input:    "SELECT * FROM users;",
			expected: "select * from users",
		},
		{
			name:     "Trim and normalize spaces",
			input:    "  SELECT   *   FROM   users   ",
			expected: "select * from users",
		},
		{
			name:     "Complex query",
			input:    "SELECT * FROM employees WHERE id = 10 OR name = 'John Doe';",
			expected: "select * from employees where id = ? or name = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnonymizeAndNormalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
