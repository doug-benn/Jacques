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
	regexp.MustCompile(`^\\set\s+`),
	regexp.MustCompile(`^\\unset\s+`),
	regexp.MustCompile(`^\\restricted\s+`),
	regexp.MustCompile(`^\\unrestricted\s+`),
}

var (
	setRE             = regexp.MustCompile(`^SET\s+\w+`)
	selectPgCatalogRE = regexp.MustCompile(`^SELECT\s+pg_catalog\.`)
	grantRE           = regexp.MustCompile(`^GRANT\s+`)
	revokeRE          = regexp.MustCompile(`^REVOKE\s+`)
	commentRE         = regexp.MustCompile(`^COMMENT\s+ON\s+`)
	alterOwnerRE      = regexp.MustCompile(`^ALTER\s+\w+\s+\S+\s+OWNER\s+TO\b`)
)

func getFirstLine(stmt string) string {
	return strings.Split(stmt, "\n")[0]
}

func IsSET(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return setRE.MatchString(stripped)
}

func IsSelectPgCatalog(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return selectPgCatalogRE.MatchString(stripped)
}

func IsGRANT(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return grantRE.MatchString(stripped)
}

func IsREVOKE(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return revokeRE.MatchString(stripped)
}

func IsCOMMENT(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return commentRE.MatchString(stripped)
}

func IsALTEROwner(stmt string) bool {
	stripped := strings.TrimSpace(getFirstLine(stmt))
	return alterOwnerRE.MatchString(stripped)
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
