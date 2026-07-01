package console

import (
	"database/sql"
	"fmt"
	"strings"
)

func formatSQLRows(rows *sql.Rows) (string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var lines []string
	lines = append(lines, strings.Join(cols, "\t"))

	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}
		parts := make([]string, len(cols))
		for i, v := range values {
			parts[i] = fmt.Sprint(v)
		}
		lines = append(lines, strings.Join(parts, "\t"))
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(lines) == 1 {
		return "(0 rows)", nil
	}
	return strings.Join(lines, "\n"), nil
}
