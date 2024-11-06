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
	}{
		{
			name:            "PostgreSQL 9.0",
			version:         "9.0.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase},
		},
		{
			name:            "PostgreSQL 9.1",
			version:         "9.1.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91},
		},
		{
			name:            "PostgreSQL 9.2",
			version:         "9.2.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91, instanceDefinition92},
		},
		{
			name:            "PostgreSQL 10.2",
			version:         "10.2.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91, instanceDefinition92},
		},
		{
			name:            "PostgreSQL 16.4",
			version:         "16.4.2",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase, instanceDefinition91, instanceDefinition92},
		},
		{
			name:            "PostgreSQL 17.0",
			version:         "17.0.0",
			expectedQueries: []*QueryDefinition{instanceDefinitionBase170, instanceDefinition170, instanceDefinitionInputOutput170},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := semver.MustParse(tt.version)
			queryDefinitions := generateInstanceDefinitions(&version)
			assert.Equal(t, len(tt.expectedQueries), len(queryDefinitions))
			assert.Equal(t, tt.expectedQueries, queryDefinitions)
		})
	}

}

func Test_generateInstanceDefinitionsOutOfOrder(t *testing.T) {
	t.Run("PostgreSQL 17.5 order check", func(t *testing.T) {
		version := semver.MustParse("17.5.0")
		queryDefinitions := generateInstanceDefinitions(&version)
		expectedQueries := []*QueryDefinition{instanceDefinitionInputOutput170, instanceDefinition170, instanceDefinitionBase170}

		// This fails beacuse order is different
		assert.False(t, assert.ObjectsAreEqual(expectedQueries, queryDefinitions), "Query definitions should be in the correct order")
	})
}
