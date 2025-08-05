package postgres

import (
	"cmp"
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

const versionTable string = "schema_version_non_default"

var defaultMigratorOptions = &migrate.MigratorOptions{
	DisableTx: false,
}

//go:embed all:*.sql
var embedFS embed.FS

type FS struct {
	URL string
}

func (fsys FS) Up(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, fsys.URL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close(context.WithoutCancel(ctx))

	migrator, err := migrate.NewMigratorEx(ctx, conn, versionTable, defaultMigratorOptions)
	if err != nil {
		return fmt.Errorf("failed to return migrator: %w", err)
	}
	// needed?
	err1 := migrator.LoadMigrations(embedFS)
	err2 := migrator.Migrate(ctx)
	if err := cmp.Or(err1, err2); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return nil
}

func (fsys FS) Down(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, fsys.URL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close(context.WithoutCancel(ctx))

	migrator, err := migrate.NewMigratorEx(ctx, conn, versionTable, defaultMigratorOptions)
	if err != nil {
		return fmt.Errorf("failed to return migrator: %w", err)
	}

	err1 := migrator.LoadMigrations(embedFS)
	err2 := migrator.MigrateTo(ctx, 0)
	if err := cmp.Or(err1, err2); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return nil
}
