package parser

import (
	"regexp"
	"strings"
)

func SkipLineComment(sql string, i int) (string, int) {
	if i >= len(sql) || sql[i] != '-' || i+1 >= len(sql) || sql[i+1] != '-' {
		return "", i
	}
	end := strings.Index(sql[i:], "\n")
	if end == -1 {
		return sql[i:], len(sql)
	}
	return sql[i : i+end+1], i + end + 1
}

func SkipBlockComment(sql string, i int) (string, int) {
	if i+1 >= len(sql) || sql[i] != '/' || sql[i+1] != '*' {
		return "", i
	}
	end := strings.Index(sql[i+2:], "*/")
	if end == -1 {
		return sql[i:], len(sql)
	}
	return sql[i : i+end+4], i + end + 4
}

func FindDollarQuoteEnd(sql string, i int) (int, string) {
	if i >= len(sql) || sql[i] != '$' {
		return -1, ""
	}
	m := regexp.MustCompile(`^\$([A-Za-z_][A-Za-z_0-9]*)?\$`).FindStringSubmatch(sql[i:])
	if m == nil {
		return -1, ""
	}
	tag := m[0]
	end := strings.Index(sql[i+len(tag):], tag)
	if end == -1 {
		return -1, ""
	}
	return i + len(tag) + end + len(tag), tag
}

func FindSingleQuoteEnd(sql string, i int) int {
	if i >= len(sql) || sql[i] != '\'' {
		return -1
	}
	j := i + 1
	n := len(sql)
	for j < n {
		if sql[j] == '\'' {
			if j+1 < n && sql[j+1] == '\'' {
				j += 2
				continue
			}
			return j + 1
		}
		j++
	}
	return -1
}

func SplitStatements(sql string) []string {
	statements := []string{}
	current := strings.Builder{}
	i := 0
	n := len(sql)

	for i < n {
		ch := sql[i]

		// Line comment
		if ch == '-' && i+1 < n && sql[i+1] == '-' {
			result, newI := SkipLineComment(sql, i)
			current.WriteString(result)
			i = newI
			continue
		}

		// Block comment
		if ch == '/' && i+1 < n && sql[i+1] == '*' {
			result, newI := SkipBlockComment(sql, i)
			current.WriteString(result)
			i = newI
			continue
		}

		// Dollar-quote
		if ch == '$' {
			endI, _ := FindDollarQuoteEnd(sql, i)
			if endI != -1 {
				current.WriteString(sql[i:endI])
				i = endI
				continue
			}
		}

		// Single-quoted string
		if ch == '\'' {
			endI := FindSingleQuoteEnd(sql, i)
			if endI != -1 {
				current.WriteString(sql[i:endI])
				i = endI
				continue
			} else {
				current.WriteString(sql[i:])
				i = n
				continue
			}
		}

		// Statement terminator
		if ch == ';' {
			current.WriteString(";")
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && stmt != ";" {
				statements = append(statements, stmt)
			}
			current.Reset()
			i++
			continue
		}

		current.WriteByte(ch)
		i++
	}

	// Trailing content
	remainder := strings.TrimSpace(current.String())
	if remainder != "" {
		statements = append(statements, remainder)
	}

	return statements
}
