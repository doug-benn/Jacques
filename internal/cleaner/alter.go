package cleaner

import (
	"regexp"
	"strings"

	"github.com/doug-benn/Jacques/internal/model"
)

var alterDiscardPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bOWNER\s+TO\b`),
	regexp.MustCompile(`\bCLUSTER\s+ON\b`),
	regexp.MustCompile(`\bSET\s+WITHOUT\s+CLUSTER\b`),
	regexp.MustCompile(`\bSET\s+WITHOUT\s+OIDS\b`),
}

var notNullRE = regexp.MustCompile(`(?i)\bNOT\s+NULL\b`)
var alterSeqOwnedBy = regexp.MustCompile(`^ALTER\s+SEQUENCE\b.*\bOWNED\s+BY\b`)
var setDefaultRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+SET\s+DEFAULT\s+(.*)`)
var setNotNullRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+SET\s+NOT\s+NULL\b`)
var setTypeRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+(?:SET\s+DATA\s+)?TYPE\s+(.*)`)
var addPKRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+CONSTRAINT\s+\S+\s+PRIMARY\s+KEY\s*\(([^)]+)\)`)
var addUniqueRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+CONSTRAINT\s+\S+\s+UNIQUE\s*\(([^)]+)\)`)
var addCheckRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+(ADD\s+CONSTRAINT\s+\S+\s+CHECK\s*\(.*)`)
var addFKRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+(ADD\s+CONSTRAINT\s+\S+\s+FOREIGN\s+KEY\s*\(.*)`)

// FK details regex - captures column, reference, and actions
var fkDetailsRE = regexp.MustCompile(`FOREIGN\s+KEY\s*\(([^)]+)\)\s*REFERENCES\s+([a-zA-Z_][a-zA-Z_0-9]*)(?:\.([a-zA-Z_][a-zA-Z_0-9]*))?\s*\(([^)]+)\)(\s+ON\s+DELETE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+ON\s+UPDATE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+MATCH\s+(FULL|PARTIAL))?`)
var addColumnRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+(?:COLUMN\s+)?(?:IF\s+NOT\s+EXISTS\s+)?([a-zA-Z_][a-zA-Z_0-9]*)\s+(.*)`)
var dropDefaultRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+DROP\s+DEFAULT\b`)
var alterNextvalRE = regexp.MustCompile(`nextval\('([^']+)'`)
var addConstraintRE = regexp.MustCompile(`^ADD\s+`)
var addExcludeRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+CONSTRAINT\s+(\S+)\s+EXCLUDE\s+(.+)`)

func FindTable(tables map[string]*model.TableDef, schema, name string) *model.TableDef {
	key := ""
	if schema != "" {
		key = schema + "." + name
	}
	if key != "" {
		if t, ok := tables[key]; ok {
			return t
		}
	}
	if t, ok := tables[name]; ok {
		return t
	}
	for _, t := range tables {
		if strings.EqualFold(t.Name, name) {
			if schema == "" || strings.EqualFold(t.Schema, schema) {
				return t
			}
		}
	}
	return nil
}

func RouteAlter(stmt string, tables map[string]*model.TableDef) *string {
	stripped := strings.TrimSpace(stmt)

	for _, pat := range alterDiscardPatterns {
		if pat.MatchString(stripped) {
			return nil
		}
	}

	if alterSeqOwnedBy.MatchString(stripped) {
		return nil
	}

	if m := setDefaultRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, col, def := m[1], m[2], m[3], m[4]
		td := FindTable(tables, schema, tname)
		if td != nil {
			for _, c := range td.Columns {
				if strings.EqualFold(c.Name, col) {
					c.Default = strings.TrimSuffix(def, ";")
					if strings.Contains(def, "nextval") {
						seqMatch := alterNextvalRE.FindStringSubmatch(def)
						if seqMatch != nil {
							c.SequenceName = seqMatch[1]
							// Only convert to SERIAL for bigint or integer (not smallint)
							rawDefLower := strings.ToLower(c.RawDef)
							if strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint") {
								c.IsSerial = true
							} else if strings.Contains(rawDefLower, "integer") || strings.EqualFold(strings.TrimSpace(c.RawDef), "int") {
								c.IsSerial = true
							}
						}
					}
					break
				}
			}
		}
		return nil
	}

	if m := setNotNullRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, col := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			for _, c := range td.Columns {
				if strings.EqualFold(c.Name, col) {
					if !strings.Contains(strings.ToUpper(c.RawDef), "NOT NULL") {
						c.RawDef = c.RawDef + " NOT NULL"
					}
					break
				}
			}
		}
		return nil
	}

	if m := setTypeRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, col, newType := m[1], m[2], m[3], m[4]
		td := FindTable(tables, schema, tname)
		if td != nil {
			for _, c := range td.Columns {
				if strings.EqualFold(c.Name, col) {
					c.RawDef = strings.TrimSuffix(newType, ";")
					break
				}
			}
		}
		return nil
	}

	if dropDefaultRE.MatchString(stripped) {
		return nil
	}

	if m := addPKRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, colsStr := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			cols := strings.Split(colsStr, ",")
			if len(cols) == 1 {
				for _, c := range td.Columns {
					if strings.EqualFold(c.Name, strings.TrimSpace(cols[0])) {
						c.IsPrimaryKey = true
						c.IsUnique = false
						c.RawDef = strings.TrimSpace(notNullRE.ReplaceAllString(c.RawDef, ""))
						break
					}
				}
			} else {
				td.TableLevelPK = colsStr
			}
		}
		return nil
	}

	if m := addUniqueRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, colsStr := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			cols := strings.Split(colsStr, ",")
			if len(cols) == 1 {
				for _, c := range td.Columns {
					if strings.EqualFold(c.Name, strings.TrimSpace(cols[0])) && !c.IsPrimaryKey {
						c.IsUnique = true
						break
					}
				}
			} else {
				td.TableLevelUniques = append(td.TableLevelUniques, colsStr)
			}
		}
		return nil
	}

	if m := addCheckRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, constraint := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			constraint = addConstraintRE.ReplaceAllString(constraint, "")
			td.TableConstraints = append(td.TableConstraints, strings.TrimSuffix(constraint, ";"))
		}
		return nil
	}

	if m := addFKRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, fkPart := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			if fkMatch := fkDetailsRE.FindStringSubmatch(fkPart); fkMatch != nil {
				fkCol := fkMatch[1]
				refSchema := fkMatch[2]
				refTable := fkMatch[3]
				refCol := fkMatch[4]
				onDelete := strings.TrimSpace(fkMatch[5])
				onUpdate := strings.TrimSpace(fkMatch[7])
				match := strings.TrimSpace(fkMatch[9])

				if refTable == "" {
					refTable = refSchema
					refSchema = ""
				}

				// Normalize ON DELETE/UPDATE/MATCH (remove prefixes)
				onDelete = strings.TrimPrefix(onDelete, "ON DELETE ")
				onUpdate = strings.TrimPrefix(onUpdate, "ON UPDATE ")
				if match != "" && strings.HasPrefix(match, "MATCH ") {
					match = strings.TrimPrefix(match, "MATCH ")
				}

				for _, c := range td.Columns {
					if strings.EqualFold(c.Name, fkCol) {
						// Only set References if not already set (from CREATE TABLE)
						// This avoids duplicate REFERENCES in output
						if c.References == "" {
							if refSchema != "" {
								c.References = refSchema + "." + refTable + "(" + refCol + ")"
							} else {
								c.References = refTable + "(" + refCol + ")"
							}
						}
						// Store cascade actions
						if onDelete != "" {
							c.OnDelete = onDelete
						}
						if onUpdate != "" {
							c.OnUpdate = onUpdate
						}
						if match != "" {
							c.Match = match
						}
						// For self-referential FKs (same table), also pass through
						// For normal FKs, just inline (no need for duplicate constraint)
						if strings.EqualFold(tname, refTable) {
							return &stmt // Self-referential - pass through
						}
						// Non-self-referential - just inline, don't pass through
						return nil
					}
				}
			}
		}
		// Fall through: pass through if we couldn't inline the FK
		return &stmt
	}

	if m := addColumnRE.FindStringSubmatch(stripped); m != nil {
		// Add the column to the table so subsequent FKs can reference it
		schema, tname, colName, colType := m[1], m[2], m[3], m[4]
		td := FindTable(tables, schema, tname)
		if td != nil {
			// Check if column already exists
			exists := false
			for _, c := range td.Columns {
				if strings.EqualFold(c.Name, colName) {
					exists = true
					break
				}
			}
			if !exists {
				// Trim trailing semicolon if present
				colType = strings.TrimSuffix(strings.TrimSpace(colType), ";")
				td.Columns = append(td.Columns, &model.ColumnDef{
					Name:   colName,
					RawDef: colType,
				})
				// Column was added to table, don't output the ALTER statement
				return nil
			}
		}
		return &stmt
	}

	if m := addExcludeRE.FindStringSubmatch(stripped); m != nil {
		schema, tname, constraintName, excludeDef := m[1], m[2], m[3], m[4]
		td := FindTable(tables, schema, tname)
		if td != nil {
			excludeDef = addConstraintRE.ReplaceAllString(excludeDef, "")
			td.TableExclusions = append(td.TableExclusions, "CONSTRAINT "+constraintName+" EXCLUDE "+strings.TrimSuffix(excludeDef, ";"))
		}
		return nil
	}

	return &stmt
}
