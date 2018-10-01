package main

import (
	"github.com/blang/semver"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/connection"
)

const (
	versionQuery = `SHOW server_version`
)

type serverVersionRow struct {
	Version string `db:"server_version"`
}

func collectVersion(connection *connection.PGSQLConnection) (semver.Version, error) {
	var versionRows []*serverVersionRow
	if err := connection.Query(&versionRows, versionQuery); err != nil {
		log.Error("Failed to execute version query: %v", err)
	}

	v, err := semver.ParseTolerant(versionRows[0].Version)
	if err != nil {
		return semver.Version{}, err
	}

	return v, nil
}
