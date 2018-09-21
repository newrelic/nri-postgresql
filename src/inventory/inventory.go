package inventory

import (
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/connection"
)

const (
	versionQuery = `SHOW server_version`
	configQuery  = `SELECT name, setting, boot_val, reset_val FROM pg_settings`
)

type ServerVersionRow struct {
	Version string `db:"server_version"`
}

type ConfigQueryRow struct {
	Name     string      `db:"name"`
	Setting  interface{} `db:"setting"`
	BootVal  interface{} `db:"boot_val"`
	ResetVal interface{} `db:"reset_val"`
}

func PopulateInventory(entity *integration.Entity, connection *connection.PGSQLConnection) {

}

func populateConfigItems(entity *integration.Entity, connection *connection.PGSQLConnection) error {
	configRows := make([]*ConfigQueryRow, 0)
	if err := connection.Query(&configRows, configQuery); err != nil {
		return err
	}

	for _, row := range configRows {
		logInventoryFailure(entity.SetInventoryItem(row.Name, "setting", row.Setting))
		logInventoryFailure(entity.SetInventoryItem(row.Name, "boot_val", row.BootVal))
		logInventoryFailure(entity.SetInventoryItem(row.Name, "reset_val", row.ResetVal))
	}

	return nil
}

func populateVersion(entity *integration.Entity, connection *connection.PGSQLConnection) error {
	versionRows := make([]*ServerVersionRow, 0)

	if err := connection.Query(&versionRows, versionQuery); err != nil {
		return err
	}

	for _, row := range versionRows {
		logInventoryFailure(entity.SetInventoryItem("version", "value", row.Version))
	}

	return nil

}

func logInventoryFailure(err error) {
	if err != nil {
		log.Error("Failed set inventory item: %v", err)
	}
}
