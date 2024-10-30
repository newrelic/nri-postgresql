package metrics

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func Test_generateInstanceDefinitions(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		expectedQueries []*QueryDefinition
		errorExpected   bool
	}{
		{
			name:            "PostgreSQL 9.0",
			version:         "9.0.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase},
			errorExpected:   false,
		},
		{
			name:            "PostgreSQL 9.1",
			version:         "9.1.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91},
			errorExpected:   false,
		},
		{
			name:            "PostgreSQL 9.2",
			version:         "9.2.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91, instanceDefinition92},
			errorExpected:   false,
		},
		{
			name:            "PostgreSQL 10.2",
			version:         "10.2.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91, instanceDefinition92},
			errorExpected:   false,
		},
		{
			name:            "PostgreSQL 16.4",
			version:         "16.4.2",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91, instanceDefinition92},
			errorExpected:   false,
		},
		{
			name:            "PostgreSQL 17.0",
			version:         "17.0.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase170, instanceDefinition170, instanceDefinitionInputOutput170},
			errorExpected:   false,
		},
		{
			name:            "PostgreSQL 17.5 out of order",
			version:         "17.5.0",
			expectedQueries: []*QueryDefinition{instanceDefinition170, instanceDefinitionInputOutput170, instanceDefinitionBase170},
			errorExpected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := semver.MustParse(tt.version)
			queryDefinitions := generateInstanceDefinitions(&version)
			assert.Equal(t, len(tt.expectedQueries), len(queryDefinitions))

			if tt.errorExpected {
				assert.False(t, assert.ObjectsAreEqual(tt.expectedQueries, queryDefinitions))
			} else {
				assert.ElementsMatch(t, tt.expectedQueries, queryDefinitions)
			}
		})
	}
}
