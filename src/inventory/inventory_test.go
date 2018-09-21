package inventory

import (
	"errors"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"testing"
)

func Test_populateConfigItems(t *testing.T) {

	testIntegration, _ := integration.New("test", "0.1.0")
	testEntity, _ := testIntegration.Entity("test", "instance")

	testConnection, mock := connection.CreateMockSQL(t)

	configRows := sqlmock.NewRows([]string{"name", "setting", "boot_val", "reset_val"}).
		AddRow("allow_system_table_mods", "off", "on", "test").
		AddRow("authentication_timeout", 1, 2, 3)

	mock.ExpectQuery(configQuery).WillReturnRows(configRows)

	err := populateConfigItems(testEntity, testConnection)

	expected := inventory.Items{
		"allow_system_table_mods": {
			"setting":   "off",
			"boot_val":  "on",
			"reset_val": "test",
		},
		"authentication_timeout": {
			"setting":   1,
			"boot_val":  2,
			"reset_val": 3,
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, testEntity.Inventory.Items())
}

func Test_populateConfigItems_Err(t *testing.T) {
	testIntegration, _ := integration.New("test", "0.1.0")
	testEntity, _ := testIntegration.Entity("test", "instance")

	testConnection, mock := connection.CreateMockSQL(t)

	mock.ExpectQuery(configQuery).WillReturnError(errors.New("test failure"))

	err := populateConfigItems(testEntity, testConnection)

	assert.Equal(t, "test failure", err.Error())

}

func Test_populateVersion(t *testing.T) {
	testIntegration, _ := integration.New("test", "0.1.0")
	testEntity, _ := testIntegration.Entity("test", "instance")

	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).
		AddRow("10.3")

	mock.ExpectQuery(versionQuery).WillReturnRows(versionRows)

	err := populateVersion(testEntity, testConnection)

	expected := inventory.Items{
		"version": {
			"value": "10.3",
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, testEntity.Inventory.Items())
}

func Test_populateVersion_Err(t *testing.T) {
	testIntegration, _ := integration.New("test", "0.1.0")
	testEntity, _ := testIntegration.Entity("test", "instance")

	testConnection, mock := connection.CreateMockSQL(t)

	mock.ExpectQuery(versionQuery).WillReturnError(errors.New("test failure"))

	err := populateVersion(testEntity, testConnection)

	assert.Equal(t, "test failure", err.Error())

}
