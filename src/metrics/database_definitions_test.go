package metrics

import (
	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_generateDatabaseDefinitions_LengthV8(t *testing.T) {
	v8 := semver.MustParse("8.0.0")

	queryDefinitions := generateDatabaseDefinitions([]string{"test1"}, v8)

	assert.Equal(t, 1, len(queryDefinitions))
}

func Test_generateDatabaseDefinitions_LengthV912(t *testing.T) {
	v912 := semver.MustParse("9.1.2")

	queryDefinitions := generateDatabaseDefinitions([]string{"test1"}, v912)

	assert.Equal(t, 1, len(queryDefinitions))
}

func Test_generateDatabaseDefinitions_LengthV925(t *testing.T) {
	v925 := semver.MustParse("9.2.5")

	queryDefinitions := generateDatabaseDefinitions([]string{"test1"}, v925)

	assert.Equal(t, 2, len(queryDefinitions))
}
