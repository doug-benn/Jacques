package cleaner

import (
	"regexp"
	"strings"
)

var noisePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^SET\s+\w+`),
	regexp.MustCompile(`(?i)^SELECT\s+pg_catalog\.`),
	regexp.MustCompile(`(?i)^GRANT\s+`),
	regexp.MustCompile(`(?i)^REVOKE\s+`),
	regexp.MustCompile(`(?i)^COMMENT\s+ON\s+`),
	regexp.MustCompile(`(?i)^ALTER\s+\w+\s+\S+\s+OWNER\s+TO\b`),
	regexp.MustCompile(`(?i)^\\`),
}

var (
	setRE             = regexp.MustCompile(`(?i)^SET\s+\w+`)
	selectPgCatalogRE = regexp.MustCompile(`(?i)^SELECT\s+pg_catalog\.`)
	grantRE           = regexp.MustCompile(`(?i)^GRANT\s+`)
	revokeRE          = regexp.MustCompile(`(?i)^REVOKE\s+`)
	commentRE         = regexp.MustCompile(`(?i)^COMMENT\s+ON\s+`)
	alterOwnerRE      = regexp.MustCompile(`(?i)^ALTER\s+\w+\s+\S+\s+OWNER\s+TO\b`)
)

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

func getFirstLine(stmt string) string {
	idx := strings.Index(stmt, "\n")
	if idx == -1 {
		return stmt
	}
	return stmt[:idx]
}

func IsSET(stmt string) bool {
	return setRE.MatchString(strings.TrimSpace(getFirstLine(stmt)))
}

func IsSelectPgCatalog(stmt string) bool {
	return selectPgCatalogRE.MatchString(strings.TrimSpace(getFirstLine(stmt)))
}

func IsGRANT(stmt string) bool {
	return grantRE.MatchString(strings.TrimSpace(getFirstLine(stmt)))
}

func IsREVOKE(stmt string) bool {
	return revokeRE.MatchString(strings.TrimSpace(getFirstLine(stmt)))
}

func IsCOMMENT(stmt string) bool {
	return commentRE.MatchString(strings.TrimSpace(getFirstLine(stmt)))
}

func IsALTEROwner(stmt string) bool {
	return alterOwnerRE.MatchString(strings.TrimSpace(getFirstLine(stmt)))
}
