package cleaner

import (
	"regexp"
	"strings"
)

var noisePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^SET\s+\w`),
	regexp.MustCompile(`^SELECT\s+pg_catalog\.`),
	regexp.MustCompile(`^GRANT\s+`),
	regexp.MustCompile(`^REVOKE\s+`),
	regexp.MustCompile(`^COMMENT\s+ON\s+`),
	regexp.MustCompile(`^ALTER\s+\w+\s+\S+\s+OWNER\s+TO\b`),
}

func IsNoise(stmt string) bool {
	firstLine := strings.Split(stmt, "\n")[0]
	stripped := strings.TrimSpace(firstLine)
	for _, pat := range noisePatterns {
		if pat.MatchString(stripped) {
			return true
		}
	}
	return false
}
