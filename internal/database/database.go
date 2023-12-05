package database

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect"
	"github.com/Masterminds/semver/v3"
	_ "github.com/lib/pq"          // PostgreSQL driver
	_ "github.com/xiaoqidun/entps" // Modernc wrapper for ent

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/appconfig"
	"github.com/nikoksr/dbench/ent/migrate"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/models"
)

var _ Store = (*DB)(nil)

// Store is an interface that defines the methods that a dbench store should implement.
type Store interface {
	Save(ctx context.Context, res *models.Benchmark) (*models.Benchmark, error)
	SaveMany(ctx context.Context, res []*models.Benchmark) ([]*models.Benchmark, error)

	Fetch(ctx context.Context, options ...QueryOption) ([]*models.Benchmark, error)
	FetchByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*models.Benchmark, error)
	FetchByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*models.Benchmark, error)
	FetchGroupIDs(ctx context.Context, options ...QueryOption) ([]string, error)
	Count(ctx context.Context, options ...QueryOption) (uint64, error)
	CountAll(ctx context.Context) (uint64, error)

	RemoveByIDs(ctx context.Context, ids []string) error
	RemoveByGroupIDs(ctx context.Context, ids []string) error
}

// DB is a struct that represents a database connection.
type DB struct {
	client      *ent.Client
	connectOnce sync.Once
	err         error
}

// New creates a new DB instance. Call Connect() to establish a connection.
func New() *DB {
	return &DB{}
}

// Connect establishes a connection to the database.
// It uses the sync.Once.Do function to ensure that the connection is only established once.
func (db *DB) Connect(_ context.Context, dsn string) (*DB, error) {
	db.connectOnce.Do(func() {
		dsn += "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"

		// Open database connection
		db.client, db.err = ent.Open(dialect.SQLite, dsn)
		if db.err != nil {
			return
		}
	})

	return db, db.err
}

// ShouldMigrate checks if the database needs to be migrated.
func (db *DB) ShouldMigrate(ctx context.Context, appVersion string) (bool, error) {
	if db.client == nil {
		return false, fmt.Errorf("no database connection")
	}

	config, err := db.client.AppConfig.Query().
		Select(appconfig.FieldVersion).
		Only(ctx)
	if err != nil {
		// Error most likely indicates that the database is empty. If it's something else, the migration will fail.
		return true, nil
	}

	// 'dev' indicates that this is a development build. AutoMigrate.
	if appVersion == "dev" || config.Version == "dev" {
		return true, nil
	}

	// AutoMigrate if the app version is greater than the database version. Otherwise, don't migrate.
	// Examples:
	//  - appVersion: 1.0.0, dbVersion: 0.9.0 -> migrate
	//  - appVersion: 1.0.0, dbVersion: 1.0.0 -> don't migrate
	_appVersion, err := semver.NewVersion(appVersion)
	if err != nil {
		return false, fmt.Errorf("invalid app version: %w", err) // What's the best way to handle this?
	}

	dbVersion, err := semver.NewVersion(config.Version)
	if err != nil {
		return false, fmt.Errorf("invalid database version: %w", err) // What's the best way to handle this?
	}

	if _appVersion.GreaterThan(dbVersion) {
		return true, nil
	}

	// Otherwise, don't migrate.
	return false, nil
}

// AutoMigrate runs the database migration.
func (db *DB) AutoMigrate(ctx context.Context, appVersion string) error {
	err := db.client.Schema.Create(ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	if err != nil {
		return err
	}

	// Upsert app config
	err = db.client.AppConfig.Create().
		SetID(1).
		SetVersion(appVersion).
		OnConflict().
		UpdateNewValues().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("create app config: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.client != nil {
		return db.client.Close()
	}

	return nil
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}

// convertToPULID is a function that converts a list of string IDs to a list of PULIDs.
func convertToPULID(ids []string) ([]pulid.ID, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}

	pulids := make([]pulid.ID, len(ids))
	for i, id := range ids {
		pulids[i] = pulid.ID(id)
	}

	return pulids, nil
}
