package db

import (
	"context"
	"os"
	"time"

	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/scylladb/gocqlx/v2/migrate"
	"go.uber.org/zap"
)

// Migrates the database schema using the migration scripts specified in the
// configuration file.
func migrateDatabaseSchema() error {
	defer common.TimeIt(schedLogger, time.Now(), "migrateSchema")

	// If database schema migration is disabled, do nothing.
	if !dbConfig.SchemaMigrationEnabled {
		schedLogger.Info("Database schema migration is disabled. Skipping ...")
		return nil
	}

	ctx := context.Background()
	gSessionMutex.Lock()
	err := migrate.FromFS(ctx, gSession, os.DirFS(dbConfig.SchemaMigrationScripts))
	gSessionMutex.Unlock()
	if err != nil {
		schedLogger.Error("Failed to upgrade database schema!",
			zap.Error(err),
		)
		return err
	}

	schedLogger.Info("Successfully completed schema migration for the database!")
	return nil
}
