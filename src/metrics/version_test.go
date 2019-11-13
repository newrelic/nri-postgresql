package metrics

import (
	"testing"

	"github.com/blang/semver"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_collectVersion(t *testing.T) {

	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).
		AddRow("10.3")

	mock.ExpectQuery(versionQuery).WillReturnRows(versionRows)

	expected := &semver.Version{
		Major: 10,
		Minor: 3,
	}

	version, err := collectVersion(testConnection)

	assert.Nil(t, err)
	assert.Equal(t, expected, version)
}

func Test_collectVersion_EnterpriseDB(t *testing.T) {

	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).
		AddRow("9.6.7.13")

	mock.ExpectQuery(versionQuery).WillReturnRows(versionRows)

	expected := &semver.Version{
		Major: 9,
		Minor: 6,
    Patch: 7,
	}

	version, err := collectVersion(testConnection)

	assert.Nil(t, err)
	assert.Equal(t, expected, version)
}

func Test_collectVersion_Ubuntu(t *testing.T) {

	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).
		AddRow("10.4 (Ubuntu 10.4-2.pgdg16.04+1)")

	mock.ExpectQuery(versionQuery).WillReturnRows(versionRows)

	expected := &semver.Version{
		Major: 10,
		Minor: 4,
	}

	version, err := collectVersion(testConnection)

	assert.Nil(t, err)
	assert.Equal(t, expected, version)
}

func Test_collectVersion_Debian(t *testing.T) {

	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).
		AddRow("10.4 (Debian 10.4-2.pgdg16.04+1)")

	mock.ExpectQuery(versionQuery).WillReturnRows(versionRows)

	expected := &semver.Version{
		Major: 10,
		Minor: 4,
	}

	version, err := collectVersion(testConnection)

	assert.Nil(t, err)
	assert.Equal(t, expected, version)
}

func Test_collectVersion_Err(t *testing.T) {

	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).
		AddRow("invalid.version.number")

	mock.ExpectQuery(versionQuery).WillReturnRows(versionRows)

	_, err := collectVersion(testConnection)

	assert.NotNil(t, err)
}
