package inventory

import (
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/connection"
)

const (
	configQuery = `SELECT name, setting, boot_val, reset_val FROM pg_settings`
)

type ConfigQueryRow struct {
	Name     string      `db:"name"`
	Setting  interface{} `db:"setting"`
	BootVal  interface{} `db:"boot_val"`
	ResetVal interface{} `db:"reset_val"`
}

func PopulateInventory(entity *integration.Entity, version string, connection *connection.PGSQLConnection) {
	populateConfigItems(entity, connection)
	populateVersion(entity, version)
}

func populateConfigItems(entity *integration.Entity, connection *connection.PGSQLConnection) {
	configRows := make([]*ConfigQueryRow, 0)
	if err := connection.Query(&configRows, configQuery); err != nil {
		log.Error("Failed to execute config query: %v", err)
	}

	for _, row := range configRows {
		logInventoryFailure(entity.SetInventoryItem(row.Name, "setting", row.Setting))
		logInventoryFailure(entity.SetInventoryItem(row.Name, "boot_val", row.BootVal))
		logInventoryFailure(entity.SetInventoryItem(row.Name, "reset_val", row.ResetVal))
	}
}

func populateVersion(entity *integration.Entity, version string) {
	logInventoryFailure(entity.SetInventoryItem("version", "value", version))
}

func logInventoryFailure(err error) {
	if err != nil {
		log.Error("Failed set inventory item: %v", err)
	}
}
