package commonutils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/newrelic/nri-postgresql/src/collection"
)

func getQuotedStringFromArray(array []string) string {
	var quotedNames = make([]string, 0)
	for _, name := range array {
		quotedNames = append(quotedNames, fmt.Sprintf("'%s'", name))
	}
	return strings.Join(quotedNames, ",")
}

func GetDatabaseListInString(dbList collection.DatabaseList) string {
	var databaseNames = make([]string, 0)
	for dbName := range dbList {
		databaseNames = append(databaseNames, dbName)
	}
	if len(databaseNames) == 0 {
		return ""
	}
	return getQuotedStringFromArray(databaseNames)
}

func AnonymizeQueryText(query string) string {
	re := regexp.MustCompile(`'[^']*'|\d+|".*?"`)
	anonymizedQuery := re.ReplaceAllString(query, "?")
	return anonymizedQuery
}

// This function is used to generate a unique plan ID for a query
func GeneratePlanID(queryID string) *string {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(RANDOM_INT_RANGE))
	if err != nil {
		return nil
	}
	currentTime := time.Now().Format(TIME_FORMAT)
	result := fmt.Sprintf("%s-%d-%s", queryID, randomInt.Int64(), currentTime)
	return &result
}
