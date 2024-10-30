package metrics

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func Test_generateInstanceDefinitions(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		expectedLength int
	}{
		{
			name:           "PostgreSQL 9.0",
			version:        "9.0.0",
			expectedLength: 1,
		},
		{
			name:           "PostgreSQL 9.1",
			version:        "9.1.0",
			expectedLength: 2,
		},
		{
			name:           "PostgreSQL 9.2",
			version:        "9.2.0",
			expectedLength: 3,
		},
		{
			name:           "PostgreSQL 10.2",
			version:        "10.2.0",
			expectedLength: 3,
		},
		{
			name:           "PostgreSQL 16.4",
			version:        "16.4.2",
			expectedLength: 3,
		},
		{
			name:           "PostgreSQL 17.0",
			version:        "17.0.0",
			expectedLength: 3,
		},
		{
			name:           "PostgreSQL 17.5",
			version:        "17.5.0",
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := semver.MustParse(tt.version)
			queryDefinitions := generateInstanceDefinitions(&version)
			assert.Equal(t, tt.expectedLength, len(queryDefinitions))
		})
	}
}
