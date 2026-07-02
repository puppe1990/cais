// Resource store/model/migration SQL generation for cais g resource.
package cli

import (
	"fmt"
	"strings"
)

func buildResourceModel(data scaffoldData) string {
	var b strings.Builder
	b.WriteString("package models\n\nimport \"time\"\n\n")
	fmt.Fprintf(&b, "type %s struct {\n\tID int64\n", data.Pascal)
	for _, f := range data.Fields {
		fmt.Fprintf(&b, "\t%s %s\n", f.Pascal, f.GoType)
	}
	b.WriteString("\tCreatedAt time.Time\n}\n")
	return b.String()
}

const tplSelectOptionModel = `package models

// SelectOption is a label/value pair for foreign-key select fields.
type SelectOption struct {
	ID    int64
	Label string
}
`

func buildReferenceStoreMethods(fields []FieldDef, existing string) string {
	var b strings.Builder
	for _, f := range uniqueReferenceFields(fields) {
		if strings.Contains(existing, "List"+f.RefPascal+"Options()") {
			continue
		}
		fmt.Fprintf(&b, `
func (s *SQLiteStore) List%sOptions() ([]models.SelectOption, error) {
	rows, err := s.db.Query(
		"SELECT id, COALESCE(NULLIF(name, ''), NULLIF(title, ''), CAST(id AS TEXT)) FROM %s ORDER BY 2",
	)
	if err != nil {
		return nil, fmt.Errorf("list %s options: %%w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.SelectOption
	for rows.Next() {
		var opt models.SelectOption
		if err := rows.Scan(&opt.ID, &opt.Label); err != nil {
			return nil, fmt.Errorf("scan %s option: %%w", err)
		}
		items = append(items, opt)
	}
	return items, rows.Err()
}
`, f.RefPascal, f.RefTable, f.RefTable, f.RefTable)
	}
	return b.String()
}

func buildResourceMigration(data scaffoldData) string {
	var cols []string
	for _, f := range data.Fields {
		cols = append(cols, fmt.Sprintf("    %s %s", f.Name, f.SQLType))
	}
	create := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n%s,\n    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n);",
		data.Plural, strings.Join(cols, ",\n"))
	return fmt.Sprintf("-- up\n%s\n\n-- down\nDROP TABLE IF EXISTS %s;\n", create, data.Plural)
}

func buildResourceStoreMethods(data scaffoldData) string {
	cols, ph := insertColumns(data.Fields)
	args := insertArgs(data.Fields)
	sets := updateSets(data.Fields)
	updArgs := insertArgs(data.Fields) + ", c.ID"
	sel := selectColumns(data.Fields)

	return nullableStoreHelpers(data.Fields) + fmt.Sprintf(`
func (s *SQLiteStore) Insert%s(c models.%s) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO %s (%s) VALUES (%s)",
		%s,
	)
	if err != nil {
		return 0, fmt.Errorf("insert %s: %%w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) Update%s(c models.%s) error {
	_, err := s.db.Exec(
		"UPDATE %s SET %s WHERE id = ?",
		%s,
	)
	if err != nil {
		return fmt.Errorf("update %s: %%w", err)
	}
	return nil
}

func (s *SQLiteStore) Delete%s(id int64) error {
	_, err := s.db.Exec("DELETE FROM %s WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete %s: %%w", err)
	}
	return nil
}

func (s *SQLiteStore) Find%sByID(id int64) (models.%s, error) {
	var c models.%s
%s
	err := s.db.QueryRow(
		"SELECT id, %s, created_at FROM %s WHERE id = ?",
		id,
	).Scan(%s)
	if err != nil {
		return models.%s{}, fmt.Errorf("find %s: %%w", err)
	}
%s
	return c, nil
}

func (s *SQLiteStore) ListAll%s() ([]models.%s, error) {
	rows, err := s.db.Query("SELECT id, %s, created_at FROM %s ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("list %s: %%w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.%s
	for rows.Next() {
		var c models.%s
%s
		if err := rows.Scan(%s); err != nil {
			return nil, fmt.Errorf("scan %s: %%w", err)
		}
%s
		items = append(items, c)
	}
	return items, rows.Err()
}
`,
		data.Pascal, data.Pascal, data.Plural, cols, ph, args, data.Snake,
		data.Pascal, data.Pascal, data.Plural, sets, updArgs, data.Snake,
		data.Pascal, data.Plural, data.Snake,
		data.Pascal, data.Pascal, data.Pascal, scanDeclare(data.Fields), sel, data.Plural, scanVars(data.Fields), data.Pascal, data.Snake, scanAssign(data.Fields),
		data.PluralPascal, data.Pascal, sel, data.Plural, data.Plural,
		data.Pascal, data.Pascal, scanLoopDeclare(data.Fields), scanVars(data.Fields), data.Snake, scanLoopAssign(data.Fields),
	)
}

func buildResourcePaginatedStoreMethod(data scaffoldData) string {
	sel := selectColumns(data.Fields)
	return fmt.Sprintf(`
func (s *SQLiteStore) List%s(page, perPage int) ([]models.%s, int, error) {
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM %s").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count %s: %%w", err)
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}
	offset := pagination.Offset(page, perPage)
	rows, err := s.db.Query(
		"SELECT id, %s, created_at FROM %s ORDER BY id DESC LIMIT ? OFFSET ?",
		perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list %s: %%w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.%s
	for rows.Next() {
		var c models.%s
%s
		if err := rows.Scan(%s); err != nil {
			return nil, 0, fmt.Errorf("scan %s: %%w", err)
		}
%s
		items = append(items, c)
	}
	return items, total, rows.Err()
}
`,
		data.PluralPascal, data.Pascal,
		data.Plural, data.Plural,
		sel, data.Plural, data.Plural,
		data.Pascal, data.Pascal,
		scanLoopDeclare(data.Fields), scanVars(data.Fields), data.Snake, scanLoopAssign(data.Fields),
	)
}

func buildResourceSeed(data scaffoldData) string {
	if !data.Seed {
		return ""
	}
	var inserts []string
	for _, f := range data.Fields {
		inserts = append(inserts, fmt.Sprintf("%s: %s", f.Pascal, seedValueForField(f)))
	}
	body := fmt.Sprintf("models.%s{%s}", data.Pascal, strings.Join(inserts, ", "))
	return fmt.Sprintf(`
func (s *SQLiteStore) SeedDemo%s() error {
	count, err := s.count%s()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err = s.Insert%s(%s)
	return err
}

func (s *SQLiteStore) count%s() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM %s").Scan(&count)
	return count, err
}
`, data.PluralPascal, data.PluralPascal, data.Pascal, body, data.PluralPascal, data.Plural)
}
