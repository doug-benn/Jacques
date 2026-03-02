package processor

import (
	"regexp"
	"strings"

	"github.com/doug-benn/Jacques/internal/cleaner"
	"github.com/doug-benn/Jacques/internal/model"
	"github.com/doug-benn/Jacques/internal/parser"
)

// StatementType represents the type of SQL statement being processed.
// This is used to categorize statements during the processing pipeline.
type StatementType int

const (
	// StatementNoise represents statements that should be ignored (e.g., noise comments)
	StatementNoise StatementType = iota

	// StatementSequence represents CREATE SEQUENCE statements
	StatementSequence

	// StatementTypeDomainSchema represents CREATE TYPE, CREATE DOMAIN, or CREATE SCHEMA statements
	StatementTypeDomainSchema

	// StatementTable represents CREATE TABLE statements
	StatementTable

	// StatementAlter represents ALTER TABLE statements
	StatementAlter

	// StatementDrop represents DROP statements
	StatementDrop

	// StatementUnknown represents any other statement type
	StatementUnknown
)

// identifierRE matches both quoted and unquoted identifiers
var identifierRE = `(?:"[^"]+"|[a-zA-Z_][a-zA-Z_0-9]*)`

var createSeqRE = regexp.MustCompile(`(?i)^CREATE\s+SEQUENCE\s+`)
var alterSeqOwnedByRE = regexp.MustCompile(`(?i)^ALTER\s+SEQUENCE\b.*\bOWNED\s+BY\b`)
var createTypeDomainSchemaRE = regexp.MustCompile(`(?m)^CREATE (TYPE|DOMAIN|SCHEMA)`)
var createDomainRE = regexp.MustCompile(`(?m)^CREATE\s+DOMAIN`)
var createDomainWithCheckRE = regexp.MustCompile(`(?i)^CREATE\s+DOMAIN[\s\S]*?CHECK`)
var createCompositeTypeRE = regexp.MustCompile(`(?m)^CREATE\s+TYPE\s+.*\s+AS\s+\(`)
var partitionOfRE = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+.*\s+PARTITION\s+OF\s+`)
var blockCommentRE = regexp.MustCompile(`(?s)/\*.*?\*/`)
var dropRE = regexp.MustCompile(`(?i)^DROP\s+(TABLE|INDEX|SEQUENCE|VIEW|MATERIALIZED\s+VIEW)\s+(IF\s+EXISTS\s+)?(\S+)`)
var schemaRE = regexp.MustCompile(`(?i)^CREATE\s+SCHEMA\s+(?:IF\s+NOT\s+EXISTS\s+)?(` + identifierRE + `)`)

// detectStatementType determines the type of SQL statement.
// It returns the StatementType based on the statement content and options.
func detectStatementType(stmt string, opts *Options) StatementType {
	stripped := strings.TrimSpace(stmt)
	upper := strings.ToUpper(stripped)

	// Check for noise first (comments, etc.)
	if cleaner.IsNoise(stmt) {
		return StatementNoise
	}

	// Check for CREATE SEQUENCE
	if createSeqRE.MatchString(stripped) {
		return StatementSequence
	}

	// Check for CREATE TYPE, DOMAIN, or SCHEMA
	if createTypeDomainSchemaRE.MatchString(stmt) {
		// Skip DOMAIN with CHECK constraints unless ExperimentalFolding is enabled
		if createDomainWithCheckRE.MatchString(stmt) && !opts.ExperimentalFolding {
			return StatementNoise // Skip DOMAIN with CHECK when not enabled
		}
		// Skip COMPOSITE types unless ExperimentalFolding is enabled
		if createCompositeTypeRE.MatchString(stmt) && !opts.ExperimentalFolding {
			return StatementNoise // Skip COMPOSITE when not enabled
		}
		return StatementTypeDomainSchema
	}

	// Skip comments
	if strings.HasPrefix(stripped, "--") {
		return StatementNoise
	}

	// Skip ALTER SEQUENCE OWNED BY
	if alterSeqOwnedByRE.MatchString(stripped) {
		return StatementNoise
	}

	// Check for CREATE TABLE
	if strings.HasPrefix(upper, "CREATE TABLE") {
		return StatementTable
	}

	// Check for ALTER TABLE
	if strings.HasPrefix(upper, "ALTER TABLE") {
		return StatementAlter
	}

	// Check for DROP statements - add IF EXISTS for idempotency
	if strings.HasPrefix(upper, "DROP ") {
		dropMatch := dropRE.FindStringSubmatch(stmt)
		if dropMatch != nil && dropMatch[2] == "" {
			return StatementDrop
		}
	}

	// Everything else is unknown (will be passed through)
	return StatementUnknown
}

func Process(sql string, opts *Options) string {
	if opts == nil {
		opts = &Options{}
	}

	// Pre-process: remove block comments and line comments
	sql = preprocessSQL(sql)

	// Split SQL into statements
	statements := parser.SplitStatements(sql)

	// Categorize statements into tables, sequences, types, and pass-throughs
	tables, _, typeStmts, passThroughs, fkPassthroughs, tableOrder := categorizeStatements(statements, opts)

	// Infer missing CREATE SCHEMA statements and append to typeStmts
	typeStmts = append(typeStmts, inferMissingSchemas(tables, typeStmts)...)

	// Count sequence usage across tables
	usageCount := countSequenceUsage(tables, tableOrder)

	// Apply SERIAL conversion based on usage count
	applySerialConversion(tables, tableOrder, usageCount)

	// Extract sequences to keep from pass-throughs
	keptSequences, convertedToSerial := extractSequencesFromPassthroughs(passThroughs, usageCount, tables)

	// Build final output
	return buildOutput(tables, keptSequences, typeStmts, passThroughs, fkPassthroughs, tableOrder, convertedToSerial, opts)
}

// categorizeStatements parses SQL statements and categorizes them into tables, sequences, types, and pass-throughs.
// It returns maps and slices for tables, sequences, type statements, pass-through statements, FK pass-throughs,
// and the order in which tables were encountered.
//
// Parameters:
//   - statements: slice of SQL statements to categorize
//   - opts: processing options (can be nil)
//
// Returns:
//   - tables: map of table name to TableDef
//   - sequences: map of sequence name to existence flag
//   - typeStmts: slice of CREATE TYPE/DOMAIN/SCHEMA statements
//   - passThroughs: slice of statements to pass through unchanged
//   - fkPassthroughs: slice of FK-related statements that need special handling
//   - tableOrder: slice of table keys in the order they were encountered
func categorizeStatements(statements []string, opts *Options) (
	tables map[string]*model.TableDef,
	sequences map[string]bool,
	typeStmts []string,
	passThroughs []string,
	fkPassthroughs []string,
	tableOrder []string,
) {
	tables = make(map[string]*model.TableDef)
	sequences = make(map[string]bool)
	typeStmts = []string{}
	passThroughs = []string{}
	fkPassthroughs = []string{}
	tableOrder = []string{}

	// Track seen tables to avoid duplicates in tableOrder
	seenTable := make(map[string]bool)

	// Track gated type names (DOMAIN and COMPOSITE types when not in ExperimentalFolding mode)
	gatedTypeNames := make(map[string]bool)

	// First pass: collect gated type names
	if opts != nil && !opts.ExperimentalFolding {
		for _, stmt := range statements {
			stripped := strings.TrimSpace(stmt)
			upper := strings.ToUpper(stripped)
			// Only gate DOMAIN with CHECK constraints (not basic domains)
			if strings.HasPrefix(upper, "CREATE DOMAIN") && createDomainWithCheckRE.MatchString(stmt) {
				// Extract domain name
				re := regexp.MustCompile(`(?i)^CREATE\s+DOMAIN\s+(?:IF\s+NOT\s+EXISTS\s+)?(\S+)`)
				if match := re.FindStringSubmatch(stripped); len(match) > 1 {
					typeName := strings.Trim(match[1], "\"")
					gatedTypeNames[typeName] = true
				}
			} else if createCompositeTypeRE.MatchString(stmt) {
				// Extract composite type name
				re := regexp.MustCompile(`(?i)^CREATE\s+TYPE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\S+)\s+AS`)
				if match := re.FindStringSubmatch(stripped); len(match) > 1 {
					typeName := strings.Trim(match[1], "\"")
					gatedTypeNames[typeName] = true
				}
			}
		}
	}

	for _, stmt := range statements {
		stmtType := detectStatementType(stmt, opts)

		switch stmtType {
		case StatementNoise:
			// Skip noise statements
			continue

		case StatementSequence:
			seqName := extractSequenceName(stmt)
			if seqName != "" {
				sequences[seqName] = true
			}
			passThroughs = append(passThroughs, stmt)
			continue

		case StatementTypeDomainSchema:
			// Skip COMPOSITE types unless ExperimentalFolding is enabled
			if createCompositeTypeRE.MatchString(stmt) && opts != nil && !opts.ExperimentalFolding {
				passThroughs = append(passThroughs, stmt)
				continue
			}
			// Skip DOMAIN with CHECK unless ExperimentalFolding is enabled
			if createDomainWithCheckRE.MatchString(stmt) && opts != nil && !opts.ExperimentalFolding {
				passThroughs = append(passThroughs, stmt)
				continue
			}
			typeStmts = append(typeStmts, stmt)
			continue

		case StatementTable:
			// Check if table uses a gated type - if so, pass through unchanged
			usesGatedType := false
			if len(gatedTypeNames) > 0 {
				for _, col := range strings.Split(stmt, ",") {
					for typeName := range gatedTypeNames {
						if strings.Contains(strings.ToUpper(col), strings.ToUpper(typeName)) {
							usesGatedType = true
							break
						}
					}
					if usesGatedType {
						break
					}
				}
			}
			if usesGatedType {
				passThroughs = append(passThroughs, stmt)
				continue
			}

			td, err := cleaner.ParseCreateTable(stmt)
			if err != nil {
				passThroughs = append(passThroughs, stmt)
				continue
			}
			if td == nil {
				passThroughs = append(passThroughs, stmt)
				continue
			}

			key := td.QualifiedName()
			tables[key] = td
			tableOrder = append(tableOrder, key)

			if td.Schema != "" && td.Schema != "public" {
				tables[td.Schema+"."+td.Name] = td
			}
			if !seenTable[td.Name] {
				tables[td.Name] = td
				seenTable[td.Name] = true
			}
			continue

		case StatementAlter:
			result := cleaner.RouteAlter(stmt, tables)
			if result == nil {
				continue
			}
			if strings.Contains(strings.ToUpper(stmt), "FOREIGN KEY") {
				fkPassthroughs = append(fkPassthroughs, cleaner.Transform(*result))
				continue
			}
			passThroughs = append(passThroughs, cleaner.Transform(*result))
			continue

		case StatementDrop:
			dropMatch := dropRE.FindStringSubmatch(stmt)
			if dropMatch != nil && dropMatch[2] == "" {
				// No IF EXISTS, add it
				dropType := dropMatch[1]
				objName := dropMatch[3]
				transformed := "DROP " + dropType + " IF EXISTS " + objName
				// Preserve semicolon if present
				if strings.HasSuffix(strings.TrimSuffix(stmt, ";"), ";") {
					transformed += ";"
				}
				passThroughs = append(passThroughs, transformed)
				continue
			}
			// Fall through to default

		default:
			// Unknown statement - pass through
			passThroughs = append(passThroughs, stmt)
		}
	}

	return tables, sequences, typeStmts, passThroughs, fkPassthroughs, tableOrder
}

// inferMissingSchemas analyzes tables and existing type statements to find schemas that are used but not explicitly created.
// It returns a slice of CREATE SCHEMA statements for any missing schemas.
//
// Parameters:
//   - tables: map of table definitions
//   - typeStmts: existing CREATE TYPE/DOMAIN/SCHEMA statements
//
// Returns:
//   - slice of CREATE SCHEMA statements for missing schemas
func inferMissingSchemas(tables map[string]*model.TableDef, typeStmts []string) []string {
	// Collect all schemas used by tables
	tableSchemas := make(map[string]bool)
	for _, td := range tables {
		if td.Schema != "" && td.Schema != "public" {
			tableSchemas[td.Schema] = true
		}
	}

	// Extract schemas already defined in typeStmts
	existingSchemas := make(map[string]bool)
	for _, stmt := range typeStmts {
		if matches := schemaRE.FindStringSubmatch(stmt); matches != nil {
			existingSchemas[strings.ToLower(matches[1])] = true
		}
	}

	// Add inferred CREATE SCHEMA for missing schemas
	var inferredSchemas []string
	for schema := range tableSchemas {
		if !existingSchemas[strings.ToLower(schema)] {
			inferredSchemas = append(inferredSchemas, "CREATE SCHEMA "+schema+";")
		}
	}

	return inferredSchemas
}

// normalizeSequenceName extracts just the sequence name without schema prefix.
// For example, "public.my_seq" becomes "my_seq".
func normalizeSequenceName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// countSequenceUsage counts how many times each sequence is used across all tables.
// It uses tableOrder to avoid double-counting duplicate table entries.
//
// Parameters:
//   - tables: map of table definitions
//   - tableOrder: ordered slice of table keys
//
// Returns:
//   - map of normalized sequence name to usage count
func countSequenceUsage(tables map[string]*model.TableDef, tableOrder []string) map[string]int {
	sequenceUsageCount := make(map[string]int)
	countedTables := make(map[*model.TableDef]bool)

	for _, key := range tableOrder {
		td := tables[key]
		if countedTables[td] {
			continue
		}
		countedTables[td] = true
		for _, col := range td.Columns {
			if col.SequenceName != "" {
				normalized := normalizeSequenceName(col.SequenceName)
				sequenceUsageCount[normalized]++
			}
		}
	}

	return sequenceUsageCount
}

// applySerialConversion sets the IsSerial flag on columns based on sequence usage count.
// A sequence is converted to SERIAL only if it's used by exactly one column
// and the column type is compatible (bigint, integer, or smallint).
//
// Parameters:
//   - tables: map of table definitions
//   - tableOrder: ordered slice of table keys
//   - usageCount: map of normalized sequence name to usage count
func applySerialConversion(tables map[string]*model.TableDef, tableOrder []string, usageCount map[string]int) {
	processedForSerial := make(map[*model.TableDef]bool)

	for _, key := range tableOrder {
		td := tables[key]
		if processedForSerial[td] {
			continue
		}
		processedForSerial[td] = true

		for _, col := range td.Columns {
			if col.SequenceName != "" {
				normalized := normalizeSequenceName(col.SequenceName)
				// Only set SERIAL if count == 1 AND column type is bigint, integer, or smallint
				if usageCount[normalized] == 1 {
					rawDefLower := strings.ToLower(col.RawDef)
					if strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint") {
						col.IsSerial = true
					} else if strings.Contains(rawDefLower, "smallint") {
						col.IsSerial = true
					} else if strings.Contains(rawDefLower, "integer") ||
						strings.EqualFold(strings.TrimSpace(col.RawDef), "int") {
						col.IsSerial = true
					} else {
						col.IsSerial = false
					}
				} else {
					col.IsSerial = false
				}
			}
		}
	}
}

// extractSequencesFromPassthroughs processes pass-through statements to determine which sequences should be kept.
// A sequence is kept if:
// - It's not used by any column (usageCount == 0)
// - It's used by multiple columns (usageCount >= 2)
// - It's used by exactly one column but can't be converted to SERIAL (e.g., wrong type)
//
// Parameters:
//   - passThroughs: slice of pass-through statements
//   - usageCount: map of normalized sequence name to usage count
//   - tables: map of table definitions
//
// Returns:
//   - keptSequences: slice of CREATE SEQUENCE statements to keep
//   - convertedToSerial: map of normalized sequence names that were converted to SERIAL
func extractSequencesFromPassthroughs(passThroughs []string, usageCount map[string]int, tables map[string]*model.TableDef) ([]string, map[string]bool) {
	var keptSequences []string
	convertedToSerial := make(map[string]bool)

	for _, stmt := range passThroughs {
		stripped := strings.TrimSpace(stmt)

		if strings.HasPrefix(strings.ToUpper(stripped), "CREATE SEQUENCE") {
			seqName := extractSequenceName(stmt)
			normalized := normalizeSequenceName(seqName)
			keepSequence := false
			usageCnt := usageCount[normalized]

			if usageCnt == 0 {
				keepSequence = true
			} else if usageCnt >= 2 {
				keepSequence = true
			} else if usageCnt == 1 {
				// Check if we can convert to SERIAL
				for _, td := range tables {
					for _, col := range td.Columns {
						if normalizeSequenceName(col.SequenceName) == normalized {
							rawDefLower := strings.ToLower(col.RawDef)
							canConvertToSerial := (strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint")) ||
								strings.Contains(rawDefLower, "integer") ||
								strings.EqualFold(strings.TrimSpace(col.RawDef), "int")
							if !canConvertToSerial {
								keepSequence = true
							}
						}
					}
				}
			}

			if keepSequence {
				keptSequences = append(keptSequences, stmt)
			} else {
				// Sequence was converted to SERIAL, track it to filter ALTER SEQUENCE
				convertedToSerial[normalized] = true
			}
		}
	}

	return keptSequences, convertedToSerial
}

// buildOutput assembles the final output string from categorized components.
// The output order is: types, sequences, tables, other pass-throughs, FK pass-throughs.
//
// Parameters:
//   - tables: map of table definitions
//   - sequences: slice of CREATE SEQUENCE statements to keep
//   - typeStmts: slice of CREATE TYPE/DOMAIN/SCHEMA statements (including inferred schemas)
//   - passThroughs: slice of other pass-through statements
//   - fkPassthroughs: slice of FK-related pass-through statements
//   - tableOrder: ordered slice of table keys
//   - convertedToSerial: map of sequence names converted to SERIAL
//   - opts: processing options
//
// Returns:
//   - final output string
func buildOutput(
	tables map[string]*model.TableDef,
	sequences []string,
	typeStmts []string,
	passThroughs []string,
	fkPassthroughs []string,
	tableOrder []string,
	convertedToSerial map[string]bool,
	opts *Options,
) string {
	var output []string

	// First, collect all types from both typeStmts and passThroughs
	// This ensures types are always defined before tables that use them
	// Filter COMPOSITE and DOMAIN with CHECK unless ExperimentalFolding is enabled
	var allTypes []string
	for _, stmt := range typeStmts {
		if createCompositeTypeRE.MatchString(stmt) && !opts.ExperimentalFolding {
			continue // Skip COMPOSITE types when not enabled
		}
		if createDomainWithCheckRE.MatchString(stmt) && !opts.ExperimentalFolding {
			continue // Skip DOMAIN with CHECK when not enabled
		}
		allTypes = append(allTypes, stmt)
	}
	for _, stmt := range passThroughs {
		upper := strings.ToUpper(strings.TrimSpace(stmt))
		// Handle TYPE and DOMAIN
		if strings.HasPrefix(upper, "CREATE TYPE") {
			// Skip COMPOSITE types unless ExperimentalFolding is enabled
			if createCompositeTypeRE.MatchString(stmt) && !opts.ExperimentalFolding {
				continue // Skip COMPOSITE types when not enabled
			}
			allTypes = append(allTypes, stmt)
		} else if strings.HasPrefix(upper, "CREATE DOMAIN") {
			// Skip DOMAIN with CHECK unless ExperimentalFolding is enabled
			if createDomainWithCheckRE.MatchString(stmt) && !opts.ExperimentalFolding {
				continue // Skip DOMAIN with CHECK when not enabled
			}
			allTypes = append(allTypes, stmt)
		}
	}

	// Output types first (they must be created before tables that use them)
	output = append(output, allTypes...)

	// Output sequences before tables (they must exist before tables use DEFAULT nextval)
	output = append(output, sequences...)

	// Output tables
	for _, key := range tableOrder {
		if td, ok := tables[key]; ok {
			// Skip complex INHERITS clause unless ExperimentalFolding is enabled
			// Simple inheritance (single parent) is now default
			if td.Inherits != "" && !opts.ExperimentalFolding {
				// Complex inheritance: multiple parents (comma-separated)
				if strings.Contains(td.Inherits, ",") {
					td.Inherits = ""
				}
			}
			output = append(output, cleaner.RenderTable(td))
		}
	}

	// Build implicit index map for redundant index detection
	implicitIndexes := buildImplicitIndexMap(tables)

	// Track seen index definitions for duplicate detection
	seenIndexes := make(map[string]bool)

	// Output other pass-throughs (not types, sequences, or tables)
	for _, stmt := range passThroughs {
		stripped := strings.TrimSpace(stmt)
		upper := strings.ToUpper(stripped)

		// Skip CREATE INDEX statements that are redundant or duplicates
		if strings.HasPrefix(upper, "CREATE INDEX") || strings.HasPrefix(upper, "CREATE UNIQUE INDEX") {
			if isRedundantOrDuplicateIndex(stmt, implicitIndexes, seenIndexes) {
				continue
			}
			// Track this index definition
			seenIndexes[normalizeIndexDef(stmt)] = true
		}

		// Skip types, sequences, and tables (already handled)
		// Exception: partition children when ExperimentalFolding is enabled
		// Exception: tables using gated types (DOMAIN, COMPOSITE) when not ExperimentalFolding
		if strings.HasPrefix(upper, "CREATE TABLE") {
			if opts != nil && opts.ExperimentalFolding && partitionOfRE.MatchString(stripped) {
				output = append(output, stmt)
			}
			// In default mode, tables using gated types need to be output
			if opts == nil || !opts.ExperimentalFolding {
				output = append(output, stmt)
			}
			continue
		}
		if strings.HasPrefix(upper, "CREATE TYPE") ||
			strings.HasPrefix(upper, "CREATE DOMAIN") ||
			strings.HasPrefix(upper, "CREATE SEQUENCE") {
			continue
		}

		// Skip ALTER SEQUENCE statements for sequences converted to SERIAL
		if strings.HasPrefix(upper, "ALTER SEQUENCE") {
			seqName := extractAlterSequenceName(stripped)
			if seqName != "" && convertedToSerial[normalizeSequenceName(seqName)] {
				continue
			}
		}

		output = append(output, stmt)
	}

	output = append(output, fkPassthroughs...)

	return strings.TrimSpace(strings.Join(output, "\n\n"))
}

func extractSequenceName(stmt string) string {
	re := regexp.MustCompile(`(?i)^CREATE\s+SEQUENCE\s+(?:IF\s+NOT\s+EXISTS\s+)?(` + identifierRE + `(?:\.` + identifierRE + `)?)`)
	m := re.FindStringSubmatch(stmt)
	if m != nil {
		return m[1]
	}
	return ""
}

func extractAlterSequenceName(stmt string) string {
	re := regexp.MustCompile(`(?i)^ALTER\s+SEQUENCE\s+(?:IF\s+EXISTS\s+)?(?:(` + identifierRE + `)\.)?(` + identifierRE + `)`)
	m := re.FindStringSubmatch(stmt)
	if m != nil {
		// m[1] is schema (optional), m[2] is sequence name
		if m[1] != "" {
			return m[1] + "." + m[2]
		}
		return m[2]
	}
	return ""
}

// preprocessSQL removes block comments and line comments from SQL
// This prevents the parser from combining multiple statements that have comments between them
func preprocessSQL(sql string) string {
	// 1. Remove block comments /* ... */
	sql = removeBlockComments(sql)

	// 2. Remove line comments -- ...
	sql = removeLineComments(sql)

	return sql
}

// removeBlockComments removes block comments (/* ... */) from SQL
func removeBlockComments(sql string) string {
	return blockCommentRE.ReplaceAllString(sql, "")
}

// removeLineComments removes line comments (-- comment) from SQL
// This prevents the parser from combining multiple statements that have comments between them
func removeLineComments(sql string) string {
	var result strings.Builder
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		// Check if this line starts with a line comment
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			// Skip the comment line entirely
			continue
		}
		// Check if there's a comment in the middle of the line
		if idx := strings.Index(line, "--"); idx >= 0 {
			// Keep only the part before the comment
			line = line[:idx]
		}
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}

// buildImplicitIndexMap builds a map of table.columns that have implicit indexes
// from PRIMARY KEY and UNIQUE constraints
func buildImplicitIndexMap(tables map[string]*model.TableDef) map[string]bool {
	implicit := make(map[string]bool)

	for _, td := range tables {
		tableKey := td.Schema + "." + td.Name
		if td.Schema == "" {
			tableKey = td.Name
		}

		// Check table-level PRIMARY KEY
		if td.TableLevelPK != "" {
			// Table-level PK with multiple columns
			cols := strings.Split(td.TableLevelPK, ", ")
			for _, col := range cols {
				col = strings.TrimSpace(cols[0])
				implicit[tableKey+"."+col] = true
			}
		}

		// Check table-level UNIQUE constraints
		for _, uniqueCols := range td.TableLevelUniques {
			cols := strings.Split(uniqueCols, ", ")
			for _, col := range cols {
				col = strings.TrimSpace(col)
				implicit[tableKey+"."+col] = true
			}
		}

		// Check column-level constraints
		for _, col := range td.Columns {
			if col.IsPrimaryKey {
				implicit[tableKey+"."+col.Name] = true
			}
			if col.IsUnique {
				implicit[tableKey+"."+col.Name] = true
			}
		}
	}

	return implicit
}

// normalizeIndexDef creates a normalized key for an index definition
// to detect duplicates
func normalizeIndexDef(stmt string) string {
	// Extract table name, columns, and options to create a unique key
	// Format: table(col1,col2)[UNIQUE][WHERE...][INCLUDE...]
	re := regexp.MustCompile(`(?i)^CREATE\s+(UNIQUE\s+)?INDEX\s+(?:\S+\s+)?ON\s+(` + identifierRE + `)\.(` + identifierRE + `)\s*\(([^)]+)\)`)
	m := re.FindStringSubmatch(stmt)
	if m == nil {
		return ""
	}

	isUnique := m[1] != ""
	tableSchema := m[2]
	tableName := m[3]
	columnsStr := m[4]

	// Normalize: lowercase, trim spaces
	columnsStr = strings.ToLower(strings.ReplaceAll(columnsStr, " ", ""))

	// Build key: table(col)[UNIQUE][WHERE...][INCLUDE...]
	key := tableSchema + "." + tableName + "(" + columnsStr + ")"
	if isUnique {
		key = "UNIQUE " + key
	}

	// Add WHERE clause if present
	whereMatch := regexp.MustCompile(`(?i)\bWHERE\s+(.+)$`).FindStringSubmatch(stmt)
	if whereMatch != nil {
		key += "[WHERE " + strings.ToLower(strings.TrimSpace(whereMatch[1])) + "]"
	}

	// Add INCLUDE columns if present
	includeMatch := regexp.MustCompile(`(?i)\bINCLUDE\s+\(([^)]+)\)`).FindStringSubmatch(stmt)
	if includeMatch != nil {
		includeCols := strings.ToLower(strings.ReplaceAll(includeMatch[1], " ", ""))
		key += "[INCLUDE(" + includeCols + ")]"
	}

	return key
}

// isRedundantOrDuplicateIndex checks if a CREATE INDEX statement is redundant or duplicate
func isRedundantOrDuplicateIndex(stmt string, implicitIndexes map[string]bool, seenIndexes map[string]bool) bool {
	// First check if it's a redundant index (implicit from PK/UNIQUE)
	if isRedundantIndex(stmt, implicitIndexes) {
		return true
	}

	// Then check if it's a duplicate
	normalized := normalizeIndexDef(stmt)
	if normalized != "" && seenIndexes[normalized] {
		return true
	}

	return false
}

// isRedundantIndex checks if a CREATE INDEX statement is redundant
// (i.e., the index is already implicitly created by PRIMARY KEY or UNIQUE)
func isRedundantIndex(stmt string, implicitIndexes map[string]bool) bool {
	// Parse CREATE INDEX statement
	// Format: CREATE [UNIQUE] INDEX idx_name ON table(col [, col...]) [WHERE...] [INCLUDE...]
	re := regexp.MustCompile(`(?i)^CREATE\s+(UNIQUE\s+)?INDEX\s+(?:\S+\s+)?ON\s+(` + identifierRE + `)\.(` + identifierRE + `)\s*\(([^)]+)\)`)
	m := re.FindStringSubmatch(stmt)
	if m == nil {
		return false
	}

	isUnique := m[1] != ""
	tableSchema := m[2]
	tableName := m[3]
	columnsStr := m[4]

	// Build table key
	tableKey := tableSchema + "." + tableName

	// Check if this is an expression index (contains parentheses in column list)
	// Expression indexes are NOT redundant
	if strings.Contains(columnsStr, "(") {
		return false
	}

	// Check if this is a partial index (has WHERE clause)
	// Partial indexes are NOT redundant
	if regexp.MustCompile(`(?i)\bWHERE\b`).MatchString(stmt) {
		return false
	}

	// Check if this is a covering index (has INCLUDE clause)
	// Covering indexes are NOT redundant
	if regexp.MustCompile(`(?i)\bINCLUDE\b`).MatchString(stmt) {
		return false
	}

	// Get individual columns
	columns := strings.Split(columnsStr, ", ")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}

	// Single-column index on implicit index column is redundant
	// Only remove non-unique indexes (unique indexes may have different properties)
	if len(columns) == 1 && !isUnique {
		col := columns[0]
		// Remove quotes if present
		col = strings.Trim(col, `"`)
		if implicitIndexes[tableKey+"."+col] {
			return true
		}
	}

	return false
}
