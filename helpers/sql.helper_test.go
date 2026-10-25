package sdk_helper

import (
	"context"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	sqlstd "database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

// nopConnector satisfies bun's *sql.DB requirement without ever dialing a
// real database; query building never calls Connect.
type nopConnector struct{}

func (nopConnector) Connect(context.Context) (driver.Conn, error) { return nil, io.EOF }
func (nopConnector) Driver() driver.Driver                        { return nopDriver{} }

type nopDriver struct{}

func (nopDriver) Open(string) (driver.Conn, error) { return nil, io.EOF }

type sqlUserRow struct {
	_ struct{} `bun:"table:users"`

	ID     int    `bun:"id,pk"`
	Name   string `bun:"name"`
	Email  string `bun:"email"`
	Secret string `bun:"secret"`
}

func newTestBunDB(t *testing.T) *bun.DB {
	t.Helper()

	sqlDB := sqlstd.OpenDB(nopConnector{})

	t.Cleanup(func() { _ = sqlDB.Close() })

	return bun.NewDB(sqlDB, mysqldialect.New())
}

func TestSqlIncColumn(t *testing.T) {
	h := NewSql()
	db := newTestBunDB(t)

	sq := db.NewSelect().Model((*sqlUserRow)(nil))

	sq = h.IncColumn("id", "name")(sq)

	got := sq.String()
	if !strings.Contains(got, "`id`") || !strings.Contains(got, "`name`") {
		t.Fatalf("expected projected columns, got %q", got)
	}

	if strings.Contains(got, "`email`") {
		t.Fatalf("unlisted column leaked: %q", got)
	}

	// No columns -> query untouched.
	plain := db.NewSelect().Model((*sqlUserRow)(nil))
	if h.IncColumn()(plain).String() != plain.String() {
		t.Fatal("empty IncColumn must be a no-op")
	}
}

func TestSqlExcColumn(t *testing.T) {
	h := NewSql()
	db := newTestBunDB(t)

	sq := h.ExcColumn("secret")(db.NewSelect().Model((*sqlUserRow)(nil)))

	got := sq.String()
	if strings.Contains(got, "`secret`") {
		t.Fatalf("excluded column still present: %q", got)
	}

	if !strings.Contains(got, "`name`") {
		t.Fatalf("kept column missing: %q", got)
	}

	plain := db.NewSelect().Model((*sqlUserRow)(nil))
	if h.ExcColumn()(plain).String() != plain.String() {
		t.Fatal("empty ExcColumn must be a no-op")
	}
}

func TestSqlColumnOptions(t *testing.T) {
	h := NewSql()
	db := newTestBunDB(t)

	sq := h.Column(&sdk_dto.ColumnOptions{
		Inc: []string{"id", "email", "secret"},
		Exc: []string{"secret"},
	})(db.NewSelect().Model((*sqlUserRow)(nil)))

	got := sq.String()
	if !strings.Contains(got, "`id`") || !strings.Contains(got, "`email`") {
		t.Fatalf("included columns missing: %q", got)
	}

	if strings.Contains(got, "`name`") || strings.Contains(got, "`secret`") {
		t.Fatalf("unexpected columns present: %q", got)
	}

	plain := db.NewSelect().Model((*sqlUserRow)(nil))
	if h.Column(nil)(plain).String() != plain.String() {
		t.Fatal("nil options must be a no-op")
	}

	empty := h.Column(&sdk_dto.ColumnOptions{})(plain).String()
	if empty != plain.String() {
		t.Fatal("empty options must be a no-op")
	}
}
