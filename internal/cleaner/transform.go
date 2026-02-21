package cleaner

import (
	"regexp"
	"strings"

	"github.com/doug-benn/Jacques/internal/model"
)

var createTableRE = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s*\(`)
var inheritsRE = regexp.MustCompile(`(?i)\s+INHERITS\s*\([^)]+\)`)
var partitionByRE = regexp.MustCompile(`(?i)\s+PARTITION\s+BY\s+\w+\s*\([^)]+\)`)
var partitionOfRE = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+PARTITION\s+OF\s+`)
var colDefRE = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z_0-9]*)\s+(.+)$`)
var nextvalRE = regexp.MustCompile(`nextval\('([^']+)'`)
var onlyKeywordRE = regexp.MustCompile(`(?i)\bONLY\b\s*`)
var primaryKeyNotNullRE = regexp.MustCompile(`(?i)\bPRIMARY\s+KEY\s*\([^)]+\)\s*NOT\s+NULL\b`)
var fkRefRE = regexp.MustCompile(`REFERENCES\s+([a-zA-Z_][a-zA-Z_0-9]*)\s*(?:\(([^)]+)\))?(\s+ON\s+DELETE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+ON\s+UPDATE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+MATCH\s+(FULL|PARTIAL))?`)
var whitespaceRE = regexp.MustCompile(`\s+`)
var transformNotNullRE = regexp.MustCompile(`(?i)\bNOT\s+NULL\b`)

func ParseCreateTable(stmt string) (*model.TableDef, error) {
	// Check if this is a PARTITION OF (child partition) - skip these
	if partitionOfRE.MatchString(stmt) {
		return nil, nil
	}

	m := createTableRE.FindStringSubmatch(stmt)
	if m == nil {
		return nil, nil
	}

	schema := m[1]
	name := m[2]

	startParen := strings.Index(stmt, "(")
	if startParen == -1 {
		return nil, nil
	}

	depth := 0
	bodyStart := startParen + 1
	var bodyEnd int
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

	body := stmt[bodyStart:bodyEnd]
	body = strings.TrimSpace(body)

	td := &model.TableDef{
		Schema:    schema,
		Name:      name,
		RawHeader: strings.TrimSpace(stmt[:startParen]),
	}

	// Capture INHERITS clause if present (comes after the closing parenthesis)
	if inheritsMatch := inheritsRE.FindString(stmt[bodyEnd:]); inheritsMatch != "" {
		td.Inherits = strings.TrimSpace(inheritsMatch)
	}

	// Capture PARTITION BY clause if present (comes after the closing parenthesis)
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

	if m := nextvalRE.FindStringSubmatch(col.RawDef); m != nil {
		col.SequenceName = m[1]
		col.Default = "nextval('" + m[1] + "'::regclass)"
		// Only convert to SERIAL for bigint or integer (not smallint)
		rawDefLower := strings.ToLower(col.RawDef)
		if strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint") {
			col.IsSerial = true
		} else if strings.Contains(rawDefLower, "integer") || strings.EqualFold(strings.TrimSpace(col.RawDef), "int") {
			col.IsSerial = true
		}
	}

	if strings.Contains(strings.ToUpper(col.RawDef), "PRIMARY KEY") {
		col.IsPrimaryKey = true
	}
	if strings.Contains(strings.ToUpper(col.RawDef), "UNIQUE") {
		col.IsUnique = true
	}

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

		// Normalize and store cascade actions
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

	return col
}

func cleanRawDef(raw string) string {
	raw = strings.TrimSuffix(raw, ",")
	raw = strings.TrimSpace(raw)

	raw = whitespaceRE.ReplaceAllString(raw, " ")

	return raw
}

func RenderTable(td *model.TableDef) string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE ")
	sb.WriteString(td.QualifiedName())
	sb.WriteString(" (\n")

	for i, col := range td.Columns {
		sb.WriteString("    ")
		sb.WriteString(renderColumn(col))
		if i < len(td.Columns)-1 || len(td.TableConstraints) > 0 || td.TableLevelPK != "" || len(td.TableLevelUniques) > 0 || len(td.TableLevelFKs) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	if td.TableLevelPK != "" {
		sb.WriteString("    PRIMARY KEY (")
		sb.WriteString(td.TableLevelPK)
		sb.WriteString(")")
		if len(td.TableConstraints) > 0 || len(td.TableLevelUniques) > 0 || len(td.TableLevelFKs) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	for i, u := range td.TableLevelUniques {
		sb.WriteString("    UNIQUE (")
		sb.WriteString(u)
		sb.WriteString(")")
		if i < len(td.TableLevelUniques)-1 || len(td.TableConstraints) > 0 || len(td.TableLevelFKs) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	for i, c := range td.TableConstraints {
		sb.WriteString("    ")
		sb.WriteString(c)
		if i < len(td.TableConstraints)-1 || len(td.TableLevelFKs) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	for i, fk := range td.TableLevelFKs {
		sb.WriteString("    ")
		sb.WriteString(fk)
		if i < len(td.TableLevelFKs)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(")")

	if td.Inherits != "" {
		sb.WriteString(" ")
		sb.WriteString(td.Inherits)
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

	// For SERIAL columns, we need to explicitly add PRIMARY KEY if it's a primary key
	// because we're replacing the entire column definition with just SERIAL/BIGSERIAL
	if col.IsSerial {
		if strings.Contains(strings.ToLower(col.RawDef), "bigint") {
			sb.WriteString("BIGSERIAL")
		} else {
			sb.WriteString("SERIAL")
		}
		// Add PRIMARY KEY if this is a primary key column
		if col.IsPrimaryKey {
			sb.WriteString(" PRIMARY KEY")
		}
	} else {
		sb.WriteString(rawDef)
	}

	// Only add PRIMARY KEY if not already in rawDef (for non-SERIAL columns)
	if !col.IsSerial && col.IsPrimaryKey && !strings.Contains(strings.ToUpper(rawDef), "PRIMARY KEY") {
		sb.WriteString(" PRIMARY KEY")
	}
	// Only add UNIQUE if not already in rawDef
	if col.IsUnique && !strings.Contains(strings.ToUpper(rawDef), "UNIQUE") {
		sb.WriteString(" UNIQUE")
	}

	if col.Default != "" && !col.IsSerial && !strings.Contains(strings.ToUpper(rawDef), "DEFAULT") {
		sb.WriteString(" DEFAULT ")
		sb.WriteString(col.Default)
	}

	// Only add REFERENCES if not already in rawDef
	// Use rawDef (modified) not col.RawDef to check
	if col.References != "" && !col.IsPrimaryKey && !strings.Contains(strings.ToUpper(rawDef), "REFERENCES") {
		sb.WriteString(" REFERENCES ")
		sb.WriteString(col.References)
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
		// Only add NOT NULL if it was in the original and we removed it due to references
		if hadReferences && strings.Contains(strings.ToUpper(col.RawDef), "NOT NULL") && !strings.Contains(strings.ToUpper(rawDef), "NOT NULL") {
			sb.WriteString(" NOT NULL")
		}
	}

	return sb.String()
}

func Transform(sql string) string {
	result := removeOnlyKeyword(sql)
	result = removeNotNullAfterPrimaryKey(result)
	return result
}

func removeOnlyKeyword(sql string) string {
	return onlyKeywordRE.ReplaceAllString(sql, "")
}

func removeNotNullAfterPrimaryKey(sql string) string {
	return primaryKeyNotNullRE.ReplaceAllStringFunc(sql, func(match string) string {
		return strings.ToUpper(strings.ReplaceAll(match, "NOT NULL", ""))
	})
}
