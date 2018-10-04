package connection

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// CreateMockSQL creates a Test SQLConnection. Must Close con when done
func CreateMockSQL(t *testing.T) (con *PGSQLConnection, mock sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error while mocking: %s", err.Error())
		t.FailNow()
	}

	con = &PGSQLConnection{
		connection: sqlx.NewDb(mockDB, "sqlmock"),
	}

	return
}

type MockInfo struct {
	mock.Mock
}

func (mi *MockInfo) NewConnection(database string) (*PGSQLConnection, error) {
	args := mi.Called(database)
	return args.Get(0).(*PGSQLConnection), args.Error(1)
}
