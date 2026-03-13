package cleaner

import (
	"regexp"
	"strings"

	"github.com/doug-benn/Jacques/internal/model"
)

// identifierRE matches both quoted and unquoted identifiers
// Quoted: "userProfiles", unquoted: users
var identifierRE = `(?:"[^"]+"|[a-zA-Z_][a-zA-Z_0-9]*)`

var createTableRE = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:(` + identifierRE + `)\.)?(` + identifierRE + `)\s*\(`)
var inheritsRE = regexp.MustCompile(`(?i)\s+INHERITS\s*\([^)]+\)`)
var partitionByRE = regexp.MustCompile(`(?i)\s+PARTITION\s+BY\s+\w+\s*\([^)]+\)`)
var partitionOfRE = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:(` + identifierRE + `)\.)?(` + identifierRE + `)\s+PARTITION\s+OF\s+`)
var colDefRE = regexp.MustCompile(`^\s*(` + identifierRE + `)\s+(.+)$`)
var nextvalRE = regexp.MustCompile(`nextval\('([^']+)'`)
var primaryKeyNotNullRE = regexp.MustCompile(`(?i)\bPRIMARY\s+KEY\s*\([^)]+\)\s*NOT\s+NULL\b`)
var fkRefRE = regexp.MustCompile(`REFERENCES\s+(` + identifierRE + `)\s*(?:\(([^)]+)\))?(\s+ON\s+DELETE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+ON\s+UPDATE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+MATCH\s+(FULL|PARTIAL))?`)
var whitespaceRE = regexp.MustCompile(`\s+`)
var transformNotNullRE = regexp.MustCompile(`(?i)\bNOT\s+NULL\b`)

func ParseCreateTable(stmt string) (*model.TableDef, error) {
	if isPartitionOf(stmt) {
		return nil, nil
	}

	schema, name := extractTableName(stmt)
	if name == "" {
		return nil, nil
	}

	bodyStart, bodyEnd := findTableBody(stmt)
	if bodyEnd == 0 {
		return nil, nil
	}

	body := stmt[bodyStart:bodyEnd]
	body = strings.TrimSpace(body)

	td := &model.TableDef{
		Schema:    schema,
		Name:      name,
		RawHeader: strings.TrimSpace(stmt[:bodyStart]),
	}

	if inheritsMatch := inheritsRE.FindString(stmt[bodyEnd:]); inheritsMatch != "" {
		td.Inherits = strings.TrimSpace(inheritsMatch)
	}

	if partitionMatch := partitionByRE.FindString(stmt[bodyEnd:]); partitionMatch != "" {
		td.PartitionBy = strings.TrimSpace(partitionMatch)
	}

	cols, constraints, err := parseTableBody(body)
	if err != nil {
		return nil, err
	}
	td.Columns = cols
	td.TableConstraints = constraints

	return td, nil
}

func isPartitionOf(stmt string) bool {
	return partitionOfRE.MatchString(stmt)
}

func extractTableName(stmt string) (schema, name string) {
	m := createTableRE.FindStringSubmatch(stmt)
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}

func findTableBody(stmt string) (bodyStart, bodyEnd int) {
	startParen := strings.Index(stmt, "(")
	if startParen == -1 {
		return 0, 0
	}

	depth := 0
	bodyStart = startParen + 1
	for i := bodyStart; i < len(stmt); i++ {
		if stmt[i] == '(' {
			depth++
		} else if stmt[i] == ')' {
			if depth == 0 {
				bodyEnd = i
				break
			}
			depth--
		}
	}

	if bodyEnd == 0 {
		bodyEnd = len(stmt)
	}

	return bodyStart, bodyEnd
}

func parseTableBody(body string) ([]*model.ColumnDef, []string, error) {
	cols := []*model.ColumnDef{}
	constraints := []string{}

	depth := 0
	colStart := 0

	for i, ch := range body {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		} else if ch == ',' && depth == 0 {
			segment := strings.TrimSpace(body[colStart:i])
			if segment != "" {
				if isColumnDef(segment) {
					col := parseColumnDef(segment)
					if col != nil {
						cols = append(cols, col)
					}
				} else {
					constraints = append(constraints, segment)
				}
			}
			colStart = i + 1
		}
	}

	lastSegment := strings.TrimSpace(body[colStart:])
	if lastSegment != "" {
		if isColumnDef(lastSegment) {
			col := parseColumnDef(lastSegment)
			if col != nil {
				cols = append(cols, col)
			}
		} else {
			constraints = append(constraints, lastSegment)
		}
	}

	return cols, constraints, nil
}

func isColumnDef(s string) bool {
	s = strings.TrimSpace(s)
	words := strings.Fields(s)
	if len(words) < 2 {
		return false
	}
	first := words[0]
	if strings.EqualFold(first, "PRIMARY") || strings.EqualFold(first, "UNIQUE") ||
		strings.EqualFold(first, "CHECK") || strings.EqualFold(first, "FOREIGN") ||
		strings.EqualFold(first, "CONSTRAINT") {
		return false
	}
	return true
}

func parseColumnDef(s string) *model.ColumnDef {
	s = strings.TrimSpace(s)

	m := colDefRE.FindStringSubmatch(s)
	if m == nil {
		return nil
	}

	col := &model.ColumnDef{
		Name:   m[1],
		RawDef: strings.TrimSpace(m[2]),
	}

	col.RawDef = cleanRawDef(col.RawDef)

	parseSequence(col)
	parsePrimaryKey(col)
	parseUnique(col)
	parseReferences(col)

	return col
}

func parseSequence(col *model.ColumnDef) {
	if m := nextvalRE.FindStringSubmatch(col.RawDef); m != nil {
		col.SequenceName = m[1]
		col.Default = "nextval('" + m[1] + "'::regclass)"
		rawDefLower := strings.ToLower(col.RawDef)
		if strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint") {
			col.IsSerial = true
		} else if strings.Contains(rawDefLower, "integer") || strings.EqualFold(strings.TrimSpace(col.RawDef), "int") {
			col.IsSerial = true
		}
	}
}

func parsePrimaryKey(col *model.ColumnDef) {
	if strings.Contains(strings.ToUpper(col.RawDef), "PRIMARY KEY") {
		col.IsPrimaryKey = true
	}
}

func parseUnique(col *model.ColumnDef) {
	if strings.Contains(strings.ToUpper(col.RawDef), "UNIQUE") {
		col.IsUnique = true
	}
}

func parseReferences(col *model.ColumnDef) {
	refMatch := fkRefRE.FindStringSubmatch(col.RawDef)
	if refMatch != nil {
		refTable := refMatch[1]
		refCol := refMatch[2]
		onDelete := refMatch[3]
		onUpdate := refMatch[5]
		match := refMatch[7]

		if refCol != "" {
			col.References = refTable + "(" + refCol + ")"
		} else {
			col.References = refTable + "(id)"
		}

		if onDelete != "" {
			col.OnDelete = strings.TrimSpace(strings.ReplaceAll(onDelete, "ON DELETE ", ""))
		}
		if onUpdate != "" {
			col.OnUpdate = strings.TrimSpace(strings.ReplaceAll(onUpdate, "ON UPDATE ", ""))
		}
		if match != "" {
			col.Match = strings.TrimSpace(strings.ReplaceAll(match, "MATCH ", ""))
		}
	}
}

func cleanRawDef(raw string) string {
	raw = strings.TrimSuffix(raw, ",")
	raw = strings.TrimSpace(raw)

	raw = whitespaceRE.ReplaceAllString(raw, " ")

	// Remove public. prefix from type references since public is the default schema
	raw = strings.ReplaceAll(raw, "public.", "")

	return raw
}

func RenderTable(td *model.TableDef) string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE ")
	sb.WriteString(td.QualifiedName())
	sb.WriteString(" (\n")

	hasSuffix := hasTableSuffixConstraints(td)

	// Render columns
	for i, col := range td.Columns {
		sb.WriteString("    ")
		sb.WriteString(renderColumn(col))
		if needsTrailingComma(i, len(td.Columns), hasSuffix) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// Render PRIMARY KEY
	if pk, ok := renderPrimaryKey(td); ok {
		sb.WriteString(pk)
		if needsTrailingComma(0, 1, len(td.TableLevelUniques) > 0 || len(td.TableConstraints) > 0 || len(td.TableLevelFKs) > 0 || len(td.TableExclusions) > 0) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// Render UNIQUE constraints
	if uniques, ok := renderUniqueConstraints(td); ok {
		for i, u := range uniques {
			sb.WriteString(u)
			if needsTrailingComma(i, len(uniques), len(td.TableConstraints) > 0 || len(td.TableLevelFKs) > 0 || len(td.TableExclusions) > 0) {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
	}

	// Render CHECK constraints
	if constraints, ok := renderTableConstraints(td); ok {
		for i, c := range constraints {
			sb.WriteString(c)
			if needsTrailingComma(i, len(constraints), len(td.TableLevelFKs) > 0 || len(td.TableExclusions) > 0) {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
	}

	// Render table-level FKs
	if fks, ok := renderTableLevelFKs(td); ok {
		for i, fk := range fks {
			sb.WriteString(fk)
			if needsTrailingComma(i, len(fks), len(td.TableExclusions) > 0) {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
	}

	// Render EXCLUDE constraints
	if exclusions, ok := renderExclusions(td); ok {
		for i, ex := range exclusions {
			sb.WriteString(ex)
			if needsTrailingComma(i, len(exclusions), false) {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(")")

	if td.Inherits != "" {
		sb.WriteString(" ")
		// Remove public. prefix from inherits since public is the default schema
		inherits := strings.ReplaceAll(td.Inherits, "public.", "")
		sb.WriteString(inherits)
	}

	if td.PartitionBy != "" {
		sb.WriteString(" ")
		sb.WriteString(td.PartitionBy)
	}

	sb.WriteString(";")

	return sb.String()
}

func renderColumn(col *model.ColumnDef) string {
	var sb strings.Builder

	sb.WriteString(col.Name)
	sb.WriteString(" ")

	// Work with a copy so we can modify it without affecting the original
	rawDef := col.RawDef
	hadReferences := false
	if col.References != "" && !col.IsPrimaryKey {
		// Remove NOT NULL from rawDef when we have references (we'll add it back if needed)
		rawDef = strings.TrimSpace(transformNotNullRE.ReplaceAllString(rawDef, ""))
		hadReferences = true
	}

	// Render column type (SERIAL or regular)
	if col.IsSerial {
		sb.WriteString(renderSerialColumn(col))
	} else {
		sb.WriteString(rawDef)
	}

	// Only add PRIMARY KEY if not already in rawDef (for non-SERIAL columns)
	if !col.IsSerial && col.IsPrimaryKey && !strings.Contains(strings.ToUpper(rawDef), "PRIMARY KEY") {
		sb.WriteString(" PRIMARY KEY")
		if col.IndexMethod != "" {
			sb.WriteString(" USING ")
			sb.WriteString(col.IndexMethod)
		}
		if col.IsDeferrable {
			sb.WriteString(" DEFERRABLE")
			if col.InitiallyDeferred {
				sb.WriteString(" INITIALLY DEFERRED")
			}
		}
	}
	// Only add UNIQUE if not already in rawDef
	if col.IsUnique && !strings.Contains(strings.ToUpper(rawDef), "UNIQUE") {
		sb.WriteString(" UNIQUE")
		if col.IsDeferrable {
			sb.WriteString(" DEFERRABLE")
			if col.InitiallyDeferred {
				sb.WriteString(" INITIALLY DEFERRED")
			}
		}
	}

	if col.Default != "" && !col.IsSerial && !strings.Contains(strings.ToUpper(rawDef), "DEFAULT") {
		sb.WriteString(" DEFAULT ")
		sb.WriteString(col.Default)
	}

	// Only add REFERENCES if not already in rawDef
	// Use rawDef (modified) not col.RawDef to check
	if col.References != "" && !col.IsPrimaryKey && !strings.Contains(strings.ToUpper(rawDef), "REFERENCES") {
		sb.WriteString(renderReferences(col))
		// Only add NOT NULL if it was in the original and we removed it due to references
		if hadReferences && strings.Contains(strings.ToUpper(col.RawDef), "NOT NULL") && !strings.Contains(strings.ToUpper(rawDef), "NOT NULL") {
			sb.WriteString(" NOT NULL")
		}
	}

	return sb.String()
}

func Transform(sql string) string {
	result := removeNotNullAfterPrimaryKey(sql)
	return result
}

// renderPrimaryKey renders the PRIMARY KEY constraint if present.
// Returns the PRIMARY KEY clause with trailing comma indicator.
func renderPrimaryKey(td *model.TableDef) (string, bool) {
	if td.TableLevelPK == "" {
		return "", false
	}
	return "    PRIMARY KEY (" + td.TableLevelPK + ")", true
}

// renderUniqueConstraints renders all UNIQUE constraints.
// Returns the constraints with trailing comma indicator.
func renderUniqueConstraints(td *model.TableDef) ([]string, bool) {
	if len(td.TableLevelUniques) == 0 {
		return nil, false
	}
	var result []string
	for _, u := range td.TableLevelUniques {
		result = append(result, "    UNIQUE ("+u+")")
	}
	return result, true
}

// renderTableConstraints renders all CHECK constraints.
// Returns the constraints with trailing comma indicator.
func renderTableConstraints(td *model.TableDef) ([]string, bool) {
	if len(td.TableConstraints) == 0 {
		return nil, false
	}
	var result []string
	for _, c := range td.TableConstraints {
		result = append(result, "    "+c)
	}
	return result, true
}

// renderTableLevelFKs renders all table-level FOREIGN KEY constraints.
// Returns the constraints with trailing comma indicator.
func renderTableLevelFKs(td *model.TableDef) ([]string, bool) {
	if len(td.TableLevelFKs) == 0 {
		return nil, false
	}
	var result []string
	for _, fk := range td.TableLevelFKs {
		result = append(result, "    "+fk)
	}
	return result, true
}

// renderExclusions renders all EXCLUDE constraints.
// Returns the constraints with trailing comma indicator.
func renderExclusions(td *model.TableDef) ([]string, bool) {
	if len(td.TableExclusions) == 0 {
		return nil, false
	}
	var result []string
	for _, ex := range td.TableExclusions {
		result = append(result, "    "+ex)
	}
	return result, true
}

// hasTableSuffixConstraints returns true if there are any table-level constraints
// after columns (PK, unique, check, FK, exclusions).
func hasTableSuffixConstraints(td *model.TableDef) bool {
	return td.TableLevelPK != "" || len(td.TableLevelUniques) > 0 ||
		len(td.TableConstraints) > 0 || len(td.TableLevelFKs) > 0 ||
		len(td.TableExclusions) > 0
}

// renderSerialColumn renders a SERIAL/BIGSERIAL/SMALLSERIAL column.
// Handles primary key detection for SERIAL columns.
func renderSerialColumn(col *model.ColumnDef) string {
	var sb strings.Builder
	rawDefLower := strings.ToLower(col.RawDef)
	if strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint") {
		sb.WriteString("BIGSERIAL")
	} else if strings.Contains(rawDefLower, "smallint") {
		sb.WriteString("SMALLSERIAL")
	} else {
		sb.WriteString("SERIAL")
	}
	// Add PRIMARY KEY if this is a primary key column
	if col.IsPrimaryKey {
		sb.WriteString(" PRIMARY KEY")
	}
	return sb.String()
}

// renderReferences builds the REFERENCES clause with ON DELETE/UPDATE actions and MATCH.
func renderReferences(col *model.ColumnDef) string {
	var sb strings.Builder
	sb.WriteString(" REFERENCES ")
	// Strip public. prefix from references since public is the default schema
	ref := strings.ReplaceAll(col.References, "public.", "")
	sb.WriteString(ref)
	// Add ON DELETE if specified
	if col.OnDelete != "" {
		sb.WriteString(" ON DELETE ")
		sb.WriteString(col.OnDelete)
	}
	// Add ON UPDATE if specified
	if col.OnUpdate != "" {
		sb.WriteString(" ON UPDATE ")
		sb.WriteString(col.OnUpdate)
	}
	// Add MATCH if specified
	if col.Match != "" {
		sb.WriteString(" MATCH ")
		sb.WriteString(col.Match)
	}
	return sb.String()
}

// needsTrailingComma checks if a comma is needed after a line in the table definition.
// It returns true if there are more items to come.
func needsTrailingComma(currentIndex, totalCount int, hasMoreAfter bool) bool {
	return currentIndex < totalCount-1 || hasMoreAfter
}

func removeNotNullAfterPrimaryKey(sql string) string {
	return primaryKeyNotNullRE.ReplaceAllStringFunc(sql, func(match string) string {
		return strings.ToUpper(strings.ReplaceAll(match, "NOT NULL", ""))
	})
}
