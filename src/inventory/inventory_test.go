package inventory

import (
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"testing"
)

func TestPopulateInventory(t *testing.T) {

	testIntegration, _ := integration.New("test", "0.1.0")
	testEntity, _ := testIntegration.Entity("test", "instance")

	testConnection, mock := connection.CreateMockSQL(t)

	configRows := sqlmock.NewRows([]string{"name", "setting", "boot_val", "reset_val"}).
		AddRow("allow_system_table_mods", "off", "on", "test").
		AddRow("authentication_timeout", 1, 2, 3)

	mock.ExpectQuery(configQuery).WillReturnRows(configRows)

	PopulateInventory(testEntity, testConnection)

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

	assert.Equal(t, expected, testEntity.Inventory.Items())
}
