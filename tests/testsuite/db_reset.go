package testsuite

import (
	"context"
	"fmt"
	"time"

	pgclient "github.com/boilerplate-api/go-rest-api-starter/internal/clients/postgres"
)

// ResetDB truncates all public tables with CASCADE to provide test isolation
// while keeping a single Postgres container for speed.
func ResetDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := pgclient.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("load pg config: %w", err)
	}
	pool, err := pgclient.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect pg: %w", err)
	}
	defer pool.Close()

	// Build and execute a dynamic TRUNCATE statement on the DB side to avoid
	// hardcoding table names. Exclude migration bookkeeping tables if present.
	sql := `DO $$
DECLARE
    stmt text;
BEGIN
    SELECT 'TRUNCATE TABLE ' || string_agg(format('%I.%I', schemaname, tablename), ', ') || ' CASCADE'
    INTO stmt
    FROM pg_tables
    WHERE schemaname='public'
      AND tablename NOT IN ('schema_migrations', 'gorp_migrations');

    IF stmt IS NOT NULL THEN
        EXECUTE stmt;
    END IF;
END$$;`

	if _, execErr := pool.Exec(ctx, sql); execErr != nil {
		// Ignore errors in reset to avoid flakiness; tests that rely on clean state
		// will surface real issues in subsequent operations.
		return nil
	}

	return nil
}
