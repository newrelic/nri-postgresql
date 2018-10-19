package metrics

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func Test_generateInstanceDefinitions90(t *testing.T) {
	v := semver.MustParse("9.0.0")
	queryDefinitions := generateInstanceDefinitions(&v)

	assert.Equal(t, 1, len(queryDefinitions))
}

func Test_generateInstanceDefinitions91(t *testing.T) {
	v := semver.MustParse("9.1.0")
	queryDefinitions := generateInstanceDefinitions(&v)

	assert.Equal(t, 2, len(queryDefinitions))
}

func Test_generateInstanceDefinitions92(t *testing.T) {
	v := semver.MustParse("9.2.0")
	queryDefinitions := generateInstanceDefinitions(&v)

	assert.Equal(t, 3, len(queryDefinitions))
}

func Test_generateInstanceDefinitions10(t *testing.T) {
	v := semver.MustParse("10.2.0")
	queryDefinitions := generateInstanceDefinitions(&v)

	assert.Equal(t, 3, len(queryDefinitions))
}
