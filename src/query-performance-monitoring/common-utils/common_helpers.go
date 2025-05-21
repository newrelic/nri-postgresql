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

// re is a regular expression that matches single-quoted strings, numbers, or double-quoted strings
var re = regexp.MustCompile(`'[^']*'|\d+|".*?"`)

func GetDatabaseListInString(dbMap collection.DatabaseList) string {
	if len(dbMap) == 0 {
		return ""
	}
	var quotedNames = make([]string, 0)
	for dbName := range dbMap {
		quotedNames = append(quotedNames, fmt.Sprintf("'%s'", dbName))
	}
	return strings.Join(quotedNames, ",")
}

func AnonymizeQueryText(query string) string {
	anonymizedQuery := re.ReplaceAllString(query, "?")
	return anonymizedQuery
}

// This function is used to generate a unique plan ID for a query
func GeneratePlanID() (string, error) {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(RandomIntRange))
	if err != nil {
		return "", ErrUnExpectedError
	}
	currentTime := time.Now().Format(TimeFormat)
	result := fmt.Sprintf("%d-%s", randomInt.Int64(), currentTime)
	return result, nil
}

func AnonymizeAndNormalize(query string) string {
	reNumbers := regexp.MustCompile(`\d+`)
	cleanedQuery := reNumbers.ReplaceAllString(query, "?")

	reSingleQuotes := regexp.MustCompile(`'[^']*'`)
	cleanedQuery = reSingleQuotes.ReplaceAllString(cleanedQuery, "?")

	reDoubleQuotes := regexp.MustCompile(`"[^"]*"`)
	cleanedQuery = reDoubleQuotes.ReplaceAllString(cleanedQuery, "?")

	cleanedQuery = strings.ReplaceAll(cleanedQuery, "$", "")

	cleanedQuery = strings.ToLower(cleanedQuery)

	cleanedQuery = strings.ReplaceAll(cleanedQuery, ";", "")

	cleanedQuery = strings.TrimSpace(cleanedQuery)

	cleanedQuery = strings.Join(strings.Fields(cleanedQuery), " ")

	return cleanedQuery
}
