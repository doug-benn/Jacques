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

		if consumed, newI := handleLineComment(sql, i, &current); consumed {
			i = newI
			continue
		}

		if consumed, newI := handleBlockComment(sql, i, &current); consumed {
			i = newI
			continue
		}

		if consumed, newI := handleDollarQuote(sql, i, &current); consumed {
			i = newI
			continue
		}

		if consumed, newI := handleSingleQuote(sql, i, n, &current); consumed {
			i = newI
			continue
		}

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

	remainder := strings.TrimSpace(current.String())
	if remainder != "" {
		statements = append(statements, remainder)
	}

	return statements
}

func handleLineComment(sql string, i int, current *strings.Builder) (bool, int) {
	if i >= len(sql) || sql[i] != '-' || i+1 >= len(sql) || sql[i+1] != '-' {
		return false, i
	}
	result, newI := SkipLineComment(sql, i)
	current.WriteString(result)
	return true, newI
}

func handleBlockComment(sql string, i int, current *strings.Builder) (bool, int) {
	if i+1 >= len(sql) || sql[i] != '/' || sql[i+1] != '*' {
		return false, i
	}
	result, newI := SkipBlockComment(sql, i)
	current.WriteString(result)
	return true, newI
}

func handleDollarQuote(sql string, i int, current *strings.Builder) (bool, int) {
	if i >= len(sql) || sql[i] != '$' {
		return false, i
	}
	endI, _ := FindDollarQuoteEnd(sql, i)
	if endI != -1 {
		current.WriteString(sql[i:endI])
		return true, endI
	}
	return false, i
}

func handleSingleQuote(sql string, i, n int, current *strings.Builder) (bool, int) {
	if i >= len(sql) || sql[i] != '\'' {
		return false, i
	}
	endI := FindSingleQuoteEnd(sql, i)
	if endI != -1 {
		current.WriteString(sql[i:endI])
		return true, endI
	} else {
		current.WriteString(sql[i:])
		return true, n
	}
}
