package processor

import (
	"regexp"
	"strings"

	"github.com/doug-benn/Jacques/internal/cleaner"
	"github.com/doug-benn/Jacques/internal/model"
	"github.com/doug-benn/Jacques/internal/parser"
)

var createSeqRE = regexp.MustCompile(`(?i)^CREATE\s+SEQUENCE\s+`)
var alterSeqOwnedByRE = regexp.MustCompile(`(?i)^ALTER\s+SEQUENCE\b.*\bOWNED\s+BY\b`)
var createTypeDomainSchemaRE = regexp.MustCompile(`(?m)^CREATE (TYPE|DOMAIN|SCHEMA)`)

func Process(sql string) string {
	// Pre-process: remove line comments to prevent parser from combining statements
	sql = removeLineComments(sql)

	statements := parser.SplitStatements(sql)

	tables := make(map[string]*model.TableDef)
	sequences := make(map[string]bool)
	passThroughs := []string{}
	typeStmts := []string{}
	fkPassthroughs := []string{}
	tableOrder := []string{}

	seenTable := make(map[string]bool)

	for _, stmt := range statements {
		stripped := strings.TrimSpace(stmt)

		if cleaner.IsNoise(stmt) {
			continue
		}

		if createSeqRE.MatchString(stripped) {
			seqName := extractSequenceName(stmt)
			if seqName != "" {
				sequences[seqName] = true
			}
			passThroughs = append(passThroughs, stmt)
			continue
		}

		// Track CREATE TYPE, CREATE DOMAIN, and CREATE SCHEMA statements to output before tables
		// Use a regex to match at the start of a line (not just anywhere in statement)
		if createTypeDomainSchemaRE.MatchString(stmt) {
			typeStmts = append(typeStmts, stmt)
			continue
		}

		// Skip comments (they don't need to be in output)
		if strings.HasPrefix(stripped, "--") {
			continue
		}

		if alterSeqOwnedByRE.MatchString(stripped) {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(stripped), "CREATE TABLE") {
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
		}

		if strings.HasPrefix(strings.ToUpper(stripped), "ALTER TABLE") {
			result := cleaner.RouteAlter(stmt, tables)
			if result == nil {
				continue
			}
			if strings.Contains(strings.ToUpper(stripped), "FOREIGN KEY") {
				fkPassthroughs = append(fkPassthroughs, cleaner.Transform(*result))
				continue
			}
			passThroughs = append(passThroughs, cleaner.Transform(*result))
			continue
		}

		passThroughs = append(passThroughs, stmt)
	}

	// Infer missing CREATE SCHEMA statements
	// Collect all schemas used by tables
	tableSchemas := make(map[string]bool)
	for _, td := range tables {
		if td.Schema != "" && td.Schema != "public" {
			tableSchemas[td.Schema] = true
		}
	}

	// Extract schemas already defined in typeStmts
	existingSchemas := make(map[string]bool)
	schemaRE := regexp.MustCompile(`(?i)^CREATE\s+SCHEMA\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-zA-Z_][a-zA-Z_0-9]*)`)
	for _, stmt := range typeStmts {
		if matches := schemaRE.FindStringSubmatch(stmt); matches != nil {
			existingSchemas[strings.ToLower(matches[1])] = true
		}
	}

	// Add inferred CREATE SCHEMA for missing schemas
	for schema := range tableSchemas {
		if !existingSchemas[strings.ToLower(schema)] {
			typeStmts = append(typeStmts, "CREATE SCHEMA "+schema+";")
		}
	}

	// Normalize sequence names (remove schema prefix) to properly detect shared sequences
	normalizeSequenceName := func(name string) string {
		// Handle both "global_id_seq" and "public.global_id_seq"
		parts := strings.Split(name, ".")
		return parts[len(parts)-1]
	}

	// Count how many times each sequence is used - only convert to SERIAL if used by exactly 1 column
	// Use tableOrder to avoid double-counting (tables map has duplicate entries)
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

	// Set IsSerial based on the count - use a new flag map to track processed tables
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
				// Only set SERIAL if count == 1 AND column type is bigint or integer (not smallint)
				if sequenceUsageCount[normalized] == 1 {
					rawDefLower := strings.ToLower(col.RawDef)
					if (strings.Contains(rawDefLower, "bigint") && !strings.Contains(rawDefLower, "smallint")) ||
						strings.Contains(rawDefLower, "integer") ||
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

	var output []string

	// First, collect all types from both typeStmts and passThroughs
	// This ensures types are always defined before tables that use them
	allTypes := append([]string{}, typeStmts...)
	for _, stmt := range passThroughs {
		upper := strings.ToUpper(strings.TrimSpace(stmt))
		if strings.HasPrefix(upper, "CREATE TYPE") || strings.HasPrefix(upper, "CREATE DOMAIN") {
			allTypes = append(allTypes, stmt)
		}
	}

	// Output types first (they must be created before tables that use them)
	output = append(output, allTypes...)

	// Collect sequences that should be kept (used by tables or standalone)
	// These must be created before tables that reference them
	var keptSequences []string
	for _, stmt := range passThroughs {
		stripped := strings.TrimSpace(stmt)

		if strings.HasPrefix(strings.ToUpper(stripped), "CREATE SEQUENCE") {
			seqName := extractSequenceName(stripped)
			normalized := normalizeSequenceName(seqName)
			keepSequence := false
			usageCount := sequenceUsageCount[normalized]
			if usageCount == 0 {
				keepSequence = true
			} else if usageCount >= 2 {
				keepSequence = true
			} else if usageCount == 1 {
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
			}
		}
	}

	// Output sequences before tables (they must exist before tables use DEFAULT nextval)
	output = append(output, keptSequences...)

	for _, key := range tableOrder {
		if td, ok := tables[key]; ok {
			output = append(output, cleaner.RenderTable(td))
		}
	}

	// Output other passthroughs (not types, sequences, or tables)
	for _, stmt := range passThroughs {
		stripped := strings.TrimSpace(stmt)
		upper := strings.ToUpper(stripped)

		// Skip types, sequences, and tables (already handled)
		if strings.HasPrefix(upper, "CREATE TYPE") ||
			strings.HasPrefix(upper, "CREATE DOMAIN") ||
			strings.HasPrefix(upper, "CREATE SEQUENCE") ||
			strings.HasPrefix(upper, "CREATE TABLE") {
			continue
		}

		output = append(output, stmt)
	}

	output = append(output, fkPassthroughs...)

	return strings.TrimSpace(strings.Join(output, "\n\n"))
}

func extractSequenceName(stmt string) string {
	re := regexp.MustCompile(`(?i)^CREATE\s+SEQUENCE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-zA-Z_][a-zA-Z_0-9]*(?:\.[a-zA-Z_][a-zA-Z_0-9]*)?)`)
	m := re.FindStringSubmatch(stmt)
	if m != nil {
		return m[1]
	}
	return ""
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
