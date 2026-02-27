package cleaner

import (
	"regexp"
	"strings"
)

var noisePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^SET\s+\w+`),
	regexp.MustCompile(`^SELECT\s+pg_catalog\.`),
	regexp.MustCompile(`^GRANT\s+`),
	regexp.MustCompile(`^REVOKE\s+`),
	regexp.MustCompile(`^COMMENT\s+ON\s+`),
	regexp.MustCompile(`^ALTER\s+\w+\s+\S+\s+OWNER\s+TO\b`),
}

func getFirstLine(stmt string) string {
	return strings.Split(stmt, "\n")[0]
}

func IsSET(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return regexp.MustCompile(`^SET\s+\w+`).MatchString(stripped)
}

func IsSelectPgCatalog(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return regexp.MustCompile(`^SELECT\s+pg_catalog\.`).MatchString(stripped)
}

func IsGRANT(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return regexp.MustCompile(`^GRANT\s+`).MatchString(stripped)
}

func IsREVOKE(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return regexp.MustCompile(`^REVOKE\s+`).MatchString(stripped)
}

func IsCOMMENT(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return regexp.MustCompile(`^COMMENT\s+ON\s+`).MatchString(stripped)
}

func IsALTEROwner(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return regexp.MustCompile(`^ALTER\s+\w+\s+\S+\s+OWNER\s+TO\b`).MatchString(stripped)
}

func IsNoise(stmt string) bool {
	firstLine := getFirstLine(stmt)
	stripped := strings.TrimSpace(firstLine)
	for _, pat := range noisePatterns {
		if pat.MatchString(stripped) {
			return true
		}
	}
	return false
}
