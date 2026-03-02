package cleaner

import (
	"regexp"
	"strings"

	"github.com/doug-benn/Jacques/internal/model"
)

// AlterResult represents the outcome of handling an ALTER TABLE statement.
// This type distinguishes between three possible outcomes that were previously
// all conflated in a single return value.
type AlterResult int

const (
	// AlterNotMatched means the handler did not match the statement.
	// The dispatcher should try the next handler.
	AlterNotMatched AlterResult = iota

	// AlterHandled means the handler matched and successfully processed the statement.
	// The statement should not be output (discarded or folded into table definition).
	AlterHandled

	// AlterPassThrough means the handler matched but could not process the statement.
	// The statement should be output as-is (passed through).
	AlterPassThrough
)

var alterDiscardPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bOWNER\s+TO\b`),
	regexp.MustCompile(`\bCLUSTER\s+ON\b`),
	regexp.MustCompile(`\bSET\s+WITHOUT\s+CLUSTER\b`),
	regexp.MustCompile(`\bSET\s+WITHOUT\s+OIDS\b`),
}

var notNullRE = regexp.MustCompile(`(?i)\bNOT\s+NULL\b`)
var alterSeqOwnedBy = regexp.MustCompile(`(?i)^ALTER\s+SEQUENCE\b.*\bOWNED\s+BY\b`)
var setDefaultRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+SET\s+DEFAULT\s+(.*)`)
var setNotNullRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+SET\s+NOT\s+NULL\b`)
var setTypeRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+(?:SET\s+DATA\s+)?TYPE\s+(.*)`)
var addPKRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+CONSTRAINT\s+(?:"[^"]+"|\S+)\s+PRIMARY\s+KEY\s*\(([^)]+)\)`)
var addUniqueRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+CONSTRAINT\s+(?:"[^"]+"|\S+)\s+UNIQUE\s*\(([^)]+)\)`)
var addCheckRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+(ADD\s+CONSTRAINT\s+(?:"[^"]+"|\S+)\s+CHECK\s*\(.*)`)
var addFKRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+(ADD\s+CONSTRAINT\s+(?:"[^"]+"|\S+)\s+FOREIGN\s+KEY\s*\(.*)`)

// FK details regex - captures column, reference, and actions
var fkDetailsRE = regexp.MustCompile(`FOREIGN\s+KEY\s*\(([^)]+)\)\s*REFERENCES\s+([a-zA-Z_][a-zA-Z_0-9]*)(?:\.([a-zA-Z_][a-zA-Z_0-9]*))?\s*\(([^)]+)\)(\s+ON\s+DELETE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+ON\s+UPDATE\s+(NO\s+ACTION|RESTRICT|CASCADE|SET\s+NULL|SET\s+DEFAULT))?(\s+MATCH\s+(FULL|PARTIAL))?`)
var addColumnRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ADD\s+(?:COLUMN\s+)?(?:IF\s+NOT\s+EXISTS\s+)?([a-zA-Z_][a-zA-Z_0-9]*)\s+(.*)`)
var dropDefaultRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+DROP\s+DEFAULT\b`)
var dropNotNullRE = regexp.MustCompile(`^ALTER\s+TABLE\s+(?:ONLY\s+)?(?:([a-zA-Z_][a-zA-Z_0-9]*)\.)?([a-zA-Z_][a-zA-Z_0-9]*)\s+ALTER\s+COLUMN\s+([a-zA-Z_][a-zA-Z_0-9]*)\s+DROP\s+NOT\s+NULL\b`)
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

// handleSetDefault processes "ALTER TABLE ... ALTER COLUMN ... SET DEFAULT ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleSetDefault(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := setDefaultRE.FindStringSubmatch(stmt); m != nil {
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
		return AlterHandled
	}
	return AlterNotMatched
}

// handleSetNotNull processes "ALTER TABLE ... ALTER COLUMN ... SET NOT NULL" statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleSetNotNull(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := setNotNullRE.FindStringSubmatch(stmt); m != nil {
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
		return AlterHandled
	}
	return AlterNotMatched
}

// handleSetType processes "ALTER TABLE ... ALTER COLUMN ... TYPE ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleSetType(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := setTypeRE.FindStringSubmatch(stmt); m != nil {
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
		return AlterHandled
	}
	return AlterNotMatched
}

// handleDropDefault processes "ALTER TABLE ... ALTER COLUMN ... DROP DEFAULT" statements.
// These are simply discarded as they don't affect the output schema.
// Returns:
//   - AlterHandled: the statement was matched and discarded
//   - AlterNotMatched: the statement did not match this handler
func handleDropDefault(stmt string, tables map[string]*model.TableDef) AlterResult {
	if dropDefaultRE.MatchString(stmt) {
		return AlterHandled
	}
	return AlterNotMatched
}

// handleDropNotNull processes "ALTER TABLE ... ALTER COLUMN ... DROP NOT NULL" statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleDropNotNull(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := dropNotNullRE.FindStringSubmatch(stmt); m != nil {
		schema, tname, col := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			for _, c := range td.Columns {
				if strings.EqualFold(c.Name, col) {
					c.RawDef = strings.TrimSpace(notNullRE.ReplaceAllString(c.RawDef, ""))
					break
				}
			}
		}
		return AlterHandled
	}
	return AlterNotMatched
}

// handleAddPK processes "ALTER TABLE ... ADD CONSTRAINT ... PRIMARY KEY ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleAddPK(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := addPKRE.FindStringSubmatch(stmt); m != nil {
		schema, tname, colsStr := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		upperStmt := strings.ToUpper(stmt)

		// Parse USING clause
		var indexMethod string
		usingRE := regexp.MustCompile(`(?i)\bUSING\s+([a-zA-Z_][a-zA-Z_0-9]*)\b`)
		usingMatch := usingRE.FindStringSubmatch(stmt)
		if usingMatch != nil {
			indexMethod = usingMatch[1]
		}

		if td != nil {
			cols := strings.Split(colsStr, ",")
			if len(cols) == 1 {
				for _, c := range td.Columns {
					if strings.EqualFold(c.Name, strings.TrimSpace(cols[0])) {
						c.IsPrimaryKey = true
						c.IsUnique = false
						c.RawDef = strings.TrimSpace(notNullRE.ReplaceAllString(c.RawDef, ""))
						// Parse DEFERRABLE for PRIMARY KEY - only add if explicitly DEFERRABLE (not NOT DEFERRABLE)
						c.IsDeferrable = strings.Contains(upperStmt, " DEFERRABLE") && !strings.Contains(upperStmt, "NOT DEFERRABLE")
						c.InitiallyDeferred = strings.Contains(upperStmt, "INITIALLY DEFERRED")
						c.IndexMethod = indexMethod
						break
					}
				}
			} else {
				td.TableLevelPK = colsStr
			}
		}
		return AlterHandled
	}
	return AlterNotMatched
}

// handleAddUnique processes "ALTER TABLE ... ADD CONSTRAINT ... UNIQUE ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleAddUnique(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := addUniqueRE.FindStringSubmatch(stmt); m != nil {
		// Don't fold USING clause - can't be folded into CREATE TABLE
		upperStmt := strings.ToUpper(stmt)
		if strings.Contains(upperStmt, "USING") {
			return AlterNotMatched
		}

		schema, tname, colsStr := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			cols := strings.Split(colsStr, ",")
			if len(cols) == 1 {
				for _, c := range td.Columns {
					if strings.EqualFold(c.Name, strings.TrimSpace(cols[0])) && !c.IsPrimaryKey {
						c.IsUnique = true
						// Parse DEFERRABLE - only add if explicitly DEFERRABLE (not NOT DEFERRABLE)
						c.IsDeferrable = strings.Contains(upperStmt, " DEFERRABLE") && !strings.Contains(upperStmt, "NOT DEFERRABLE")
						c.InitiallyDeferred = strings.Contains(upperStmt, "INITIALLY DEFERRED")
						break
					}
				}
			} else {
				td.TableLevelUniques = append(td.TableLevelUniques, colsStr)
			}
		}
		return AlterHandled
	}
	return AlterNotMatched
}

// handleAddCheck processes "ALTER TABLE ... ADD CONSTRAINT ... CHECK ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleAddCheck(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := addCheckRE.FindStringSubmatch(stmt); m != nil {
		schema, tname, constraint := m[1], m[2], m[3]
		td := FindTable(tables, schema, tname)
		if td != nil {
			constraint = addConstraintRE.ReplaceAllString(constraint, "")
			td.TableConstraints = append(td.TableConstraints, strings.TrimSuffix(constraint, ";"))
		}
		return AlterHandled
	}
	return AlterNotMatched
}

// handleAddFK processes "ALTER TABLE ... ADD CONSTRAINT ... FOREIGN KEY ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and FK was inlined into column definition
//   - AlterPassThrough: the statement was matched but couldn't be inlined (e.g., self-referential FK)
//   - AlterNotMatched: the statement did not match this handler
func handleAddFK(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := addFKRE.FindStringSubmatch(stmt); m != nil {
		// Don't fold DEFERRABLE constraints - they have different semantics
		// NOT DEFERRABLE can fold (same as default)
		upperStmt := strings.ToUpper(stmt)
		if strings.Contains(upperStmt, "DEFERRABLE") && !strings.Contains(upperStmt, "NOT DEFERRABLE") {
			return AlterNotMatched
		}

		// Don't fold USING clause - can't be folded into CREATE TABLE
		if strings.Contains(upperStmt, "USING") {
			return AlterNotMatched
		}

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
							return AlterPassThrough // Self-referential - pass through
						}
						// Non-self-referential - just inline, don't pass through
						return AlterHandled
					}
				}
			}
		}
		// Fall through: pass through if we couldn't inline the FK
		return AlterPassThrough
	}
	return AlterNotMatched
}

// handleAddColumn processes "ALTER TABLE ... ADD COLUMN ..." statements.
// This adds the column to the table definition so subsequent FKs can reference it.
// Returns:
//   - AlterHandled: the statement was matched and column was added
//   - AlterPassThrough: the statement was matched but column couldn't be added
//   - AlterNotMatched: the statement did not match this handler
func handleAddColumn(stmt string, tables map[string]*model.TableDef) AlterResult {
	// Note: The ADD COLUMN regex can also match FK statements (they contain "ADD CONSTRAINT...").
	// We handle this by checking for "CONSTRAINT" as the column name, which indicates
	// it's actually a constraint statement, not a column add.
	if m := addColumnRE.FindStringSubmatch(stmt); m != nil {
		colName := m[3]
		// Don't match constraint statements - they look like ADD COLUMN but aren't
		if strings.EqualFold(colName, "CONSTRAINT") {
			return AlterNotMatched // Not actually a column add
		}
		schema, tname, _, colType := m[1], m[2], m[3], m[4]
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
				return AlterHandled
			}
		}
		// Couldn't add column (table not found or column exists) - pass through
		return AlterPassThrough
	}
	return AlterNotMatched
}

// handleAddExclude processes "ALTER TABLE ... ADD CONSTRAINT ... EXCLUDE ..." statements.
// Returns:
//   - AlterHandled: the statement was matched and processed successfully
//   - AlterNotMatched: the statement did not match this handler
func handleAddExclude(stmt string, tables map[string]*model.TableDef) AlterResult {
	if m := addExcludeRE.FindStringSubmatch(stmt); m != nil {
		schema, tname, constraintName, excludeDef := m[1], m[2], m[3], m[4]
		td := FindTable(tables, schema, tname)
		if td != nil {
			excludeDef = addConstraintRE.ReplaceAllString(excludeDef, "")
			td.TableExclusions = append(td.TableExclusions, "CONSTRAINT "+constraintName+" EXCLUDE "+strings.TrimSuffix(excludeDef, ";"))
		}
		return AlterHandled
	}
	return AlterNotMatched
}

// RouteAlter routes an ALTER TABLE statement to the appropriate handler.
// It returns:
//   - nil: the statement was handled (discarded or folded into table definition)
//   - &stmt: the statement should be passed through (output as-is)
//
// The routing is done by trying each handler in order until one matches.
// Each handler returns an AlterResult that indicates what happened.
func RouteAlter(stmt string, tables map[string]*model.TableDef) *string {
	stripped := strings.TrimSpace(stmt)

	// First check for discard patterns (OWNER TO, CLUSTER ON, etc.)
	// These are simply discarded and not processed
	for _, pat := range alterDiscardPatterns {
		if pat.MatchString(stripped) {
			return nil
		}
	}

	// Check for ALTER SEQUENCE OWNED BY - also discarded
	if alterSeqOwnedBy.MatchString(stripped) {
		return nil
	}

	// Try each handler in sequence
	// Use switch to clearly handle each return type
	switch handleSetDefault(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleSetNotNull(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleSetType(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleDropDefault(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleDropNotNull(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleAddPK(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleAddUnique(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleAddCheck(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	// FK handler must come before ADD COLUMN because FK statements
	// can look like ADD COLUMN (they contain "ADD CONSTRAINT...")
	switch handleAddFK(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleAddColumn(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	switch handleAddExclude(stripped, tables) {
	case AlterHandled:
		return nil
	case AlterPassThrough:
		return &stmt
	case AlterNotMatched:
		// Continue to next handler
	}

	// No handler matched - pass through
	return &stmt
}

func isAlterDiscardPattern(stmt string) bool {
	for _, pat := range alterDiscardPatterns {
		if pat.MatchString(stmt) {
			return true
		}
	}
	return false
}

func matchAlterSequenceOwnedBy(stmt string) bool {
	return alterSeqOwnedBy.MatchString(stmt)
}

func matchSetDefault(stmt string) []string {
	return setDefaultRE.FindStringSubmatch(stmt)
}

func matchSetNotNull(stmt string) []string {
	return setNotNullRE.FindStringSubmatch(stmt)
}

func matchSetType(stmt string) []string {
	return setTypeRE.FindStringSubmatch(stmt)
}

func matchDropNotNull(stmt string) []string {
	return dropNotNullRE.FindStringSubmatch(stmt)
}

func matchAddPK(stmt string) []string {
	return addPKRE.FindStringSubmatch(stmt)
}

func matchAddUnique(stmt string) []string {
	return addUniqueRE.FindStringSubmatch(stmt)
}

func matchAddCheck(stmt string) []string {
	return addCheckRE.FindStringSubmatch(stmt)
}

func matchAddFK(stmt string) []string {
	return addFKRE.FindStringSubmatch(stmt)
}

func matchAddColumn(stmt string) []string {
	return addColumnRE.FindStringSubmatch(stmt)
}

func matchDropDefault(stmt string) []string {
	return dropDefaultRE.FindStringSubmatch(stmt)
}

func matchAddExclude(stmt string) []string {
	return addExcludeRE.FindStringSubmatch(stmt)
}

func newTable(name string, cols []*model.ColumnDef) *model.TableDef {
	return &model.TableDef{
		Schema:    "",
		Name:      name,
		RawHeader: "CREATE TABLE " + name + " (",
		Columns:   cols,
	}
}

func newTableWithConstraints(name string, cols []*model.ColumnDef, constraints []string) *model.TableDef {
	return &model.TableDef{
		Schema:           "",
		Name:             name,
		RawHeader:        "CREATE TABLE " + name + " (",
		Columns:          cols,
		TableConstraints: constraints,
	}
}
