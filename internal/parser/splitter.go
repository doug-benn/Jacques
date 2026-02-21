package parser

import (
	"regexp"
	"strings"
)

func SplitStatements(sql string) []string {
	statements := []string{}
	current := strings.Builder{}
	i := 0
	n := len(sql)

	for i < n {
		ch := sql[i]

		// Line comment
		if ch == '-' && i+1 < n && sql[i+1] == '-' {
			end := strings.Index(sql[i:], "\n")
			if end == -1 {
				current.WriteString(sql[i:])
				i = n
			} else {
				current.WriteString(sql[i : i+end+1])
				i = i + end + 1
			}
			continue
		}

		// Block comment
		if ch == '/' && i+1 < n && sql[i+1] == '*' {
			end := strings.Index(sql[i+2:], "*/")
			if end == -1 {
				current.WriteString(sql[i:])
				i = n
			} else {
				current.WriteString(sql[i : i+end+4])
				i = i + end + 4
			}
			continue
		}

		// Dollar-quote
		if ch == '$' {
			m := regexp.MustCompile(`^\$([A-Za-z_][A-Za-z_0-9]*)?\$`).FindStringSubmatch(sql[i:])
			if m != nil {
				tag := m[0]
				end := strings.Index(sql[i+len(tag):], tag)
				if end == -1 {
					current.WriteString(sql[i:])
					i = n
				} else {
					current.WriteString(sql[i : i+len(tag)+end+len(tag)])
					i = i + len(tag) + end + len(tag)
				}
				continue
			}
		}

		// Single-quoted string
		if ch == '\'' {
			j := i + 1
			for j < n {
				if sql[j] == '\'' {
					if j+1 < n && sql[j+1] == '\'' {
						current.WriteString(sql[i : j+2])
						i = j + 2
						j = i
						continue
					} else {
						current.WriteString(sql[i : j+1])
						i = j + 1
						break
					}
				}
				j++
			}
			if j >= n {
				current.WriteString(sql[i:j])
				i = j
			}
			continue
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
