package migrate

import (
	"database/sql"
	"io/fs"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestApply_runsPendingMigrations(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
	}

	db := testDB(t)
	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatal(err)
	}

	var name string
	if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name); err != nil {
		t.Fatalf("users table not found: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("schema_migrations count = %d, want 1", count)
	}
}

func TestApply_isIdempotent(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
	}

	db := testDB(t)
	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatal(err)
	}
	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatalf("second apply failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("schema_migrations count = %d, want 1 after re-apply", count)
	}
}

func TestApply_appliesOnlyNewMigrations(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
	}

	db := testDB(t)
	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatal(err)
	}

	migrations["migrations/002_posts.sql"] = &fstest.MapFile{
		Data: []byte(`CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT NOT NULL);`),
	}

	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("schema_migrations count = %d, want 2", count)
	}

	var name string
	if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&name); err != nil {
		t.Fatalf("posts table not found: %v", err)
	}
}

func TestApply_doesNotRecordFailedMigration(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_bad.sql": &fstest.MapFile{
			Data: []byte(`NOT VALID SQL;`),
		},
	}

	db := testDB(t)
	err := Apply(db, migrations, "migrations")
	if err == nil {
		t.Fatal("expected error for invalid SQL")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("schema_migrations count = %d, want 0 after failed migration", count)
	}
}

func TestStatus_listsAppliedAndPending(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY);`)},
		"migrations/002_posts.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE posts (id INTEGER PRIMARY KEY);`)},
	}

	db := testDB(t)
	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatal(err)
	}

	// Remove 002 from applied by only having run 001 - re-create db
	db2 := testDB(t)
	migrations2 := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY);`)},
		"migrations/002_posts.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE posts (id INTEGER PRIMARY KEY);`)},
	}
	if err := Apply(db2, migrations2, "migrations"); err != nil {
		t.Fatal(err)
	}
	_ = db

	statuses, err := Status(db2, migrations2, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 {
		t.Fatalf("len(statuses) = %d, want 2", len(statuses))
	}
	if !statuses[0].Applied || statuses[0].Version != "001_users" {
		t.Errorf("status[0] = %+v, want 001_users applied", statuses[0])
	}
	if !statuses[1].Applied || statuses[1].Version != "002_posts" {
		t.Errorf("status[1] = %+v, want 002_posts applied", statuses[1])
	}
}

func TestListSQL_ignoresNonSQLFiles(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/.gitkeep":      &fstest.MapFile{Data: []byte("")},
		"migrations/001_users.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY);`)},
	}

	files, err := listSQL(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "001_users.sql" {
		t.Errorf("files = %v, want [001_users.sql]", files)
	}
}

func TestRollbackLast_removesLastAppliedMigration(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
		"migrations/002_posts.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT NOT NULL);`),
		},
	}

	db := testDB(t)
	if err := Apply(db, migrations, "migrations"); err != nil {
		t.Fatal(err)
	}

	version, err := RollbackLast(db, migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if version != "002_posts" {
		t.Errorf("RollbackLast version = %q, want 002_posts", version)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("schema_migrations count = %d, want 1 after rollback", count)
	}

	applied, err := isApplied(db, "001_users")
	if err != nil {
		t.Fatal(err)
	}
	if !applied {
		t.Error("001_users should still be applied")
	}

	applied, err = isApplied(db, "002_posts")
	if err != nil {
		t.Fatal(err)
	}
	if applied {
		t.Error("002_posts should no longer be applied")
	}
}

func TestRollbackLast_errorsWhenNoAppliedMigrations(t *testing.T) {
	migrations := fstest.MapFS{
		"migrations/001_users.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY);`),
		},
	}

	db := testDB(t)
	_, err := RollbackLast(db, migrations, "migrations")
	if err == nil {
		t.Fatal("expected error when no migrations applied")
	}
}

// Ensure fstest.MapFS satisfies fs.FS at compile time.
var _ fs.FS = fstest.MapFS{}
