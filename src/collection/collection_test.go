package collection

import (
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_buildSchemaListForDatabase(t *testing.T) {
	testConnection, mock := connection.CreateMockSQL(t)
	instanceRows := sqlmock.NewRows([]string{
		"schema_name",
		"table_name",
		"index_name",
	}).AddRow("schema1", "table1", "index1")

	mock.ExpectQuery(dbSchemaQuery).WillReturnRows(instanceRows)

	schemaList, err := buildSchemaListForDatabase("database1", testConnection)
	assert.Nil(t, err)

	expected := SchemaList{
		"schema1": TableList{
			"table1": []string{"index1"},
		},
	}

	assert.Equal(t, expected, schemaList)
}

func Test_buildSchemaListForDatabase_TableOnly(t *testing.T) {
	testConnection, mock := connection.CreateMockSQL(t)
	instanceRows := sqlmock.NewRows([]string{
		"schema_name",
		"table_name",
		"index_name",
	}).AddRow("schema1", "table1", "index1").AddRow("schema2", "table2", nil)

	mock.ExpectQuery(dbSchemaQuery).WillReturnRows(instanceRows)

	schemaList, err := buildSchemaListForDatabase("database1", testConnection)
	assert.Nil(t, err)

	expected := SchemaList{
		"schema1": TableList{
			"table1": []string{"index1"},
		},
		"schema2": TableList{
			"table2": []string{},
		},
	}

	assert.Equal(t, expected, schemaList)
}

func TestBuildCollectionList_DatabaseList(t *testing.T) {
	al := args.ArgumentList{
		CollectionList: `["database1", "database2"]`,
	}

	ci := connection.MockInfo{}
	testConnection1, mock1 := connection.CreateMockSQL(t)
	testConnection2, mock2 := connection.CreateMockSQL(t)

	ci.On("NewConnection", "database1").Return(testConnection1, nil)
	ci.On("NewConnection", "database2").Return(testConnection2, nil)

	instanceRows1 := sqlmock.NewRows([]string{
		"schema_name",
		"table_name",
		"index_name",
	}).AddRow("schema1", "table1", "index1")
	instanceRows2 := sqlmock.NewRows([]string{
		"schema_name",
		"table_name",
		"index_name",
	}).AddRow("schema2", "table2", nil)

	mock1.ExpectQuery(dbSchemaQuery).WillReturnRows(instanceRows1)
	mock2.ExpectQuery(dbSchemaQuery).WillReturnRows(instanceRows2)

	expected := DatabaseList{
		"database1": SchemaList{
			"schema1": TableList{
				"table1": []string{"index1"},
			},
		},
		"database2": SchemaList{
			"schema2": TableList{
				"table2": []string{},
			},
		},
	}

	dl, err := BuildCollectionList(al, &ci)
	assert.Nil(t, err)
	assert.Equal(t, expected, dl)
}

func TestBuildCollectionList_DetailedList(t *testing.T) {
	al := args.ArgumentList{
		CollectionList: `{"database1": {"schema1": { "table1": ["index1"] }}}`,
	}

	ci := connection.MockInfo{}
	testConnection, mock := connection.CreateMockSQL(t)

	ci.On("NewConnection", "database1").Return(testConnection, nil)

	instanceRows := sqlmock.NewRows([]string{
		"schema_name",
		"table_name",
		"index_name",
	}).AddRow("schema1", "table1", "index1")

	mock.ExpectQuery(".*").WillReturnRows(instanceRows)

	expected := DatabaseList{
		"database1": SchemaList{
			"schema1": TableList{
				"table1": []string{"index1"},
			},
		},
	}

	dl, err := BuildCollectionList(al, &ci)
	assert.Nil(t, err)
	assert.Equal(t, expected, dl)
}
