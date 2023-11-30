package database

import (
	"context"
	"entgo.io/ent/dialect"
	"fmt"
	_ "github.com/xiaoqidun/entps" // Modernc wrapper for ent
	"sync"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
	"github.com/nikoksr/dbench/ent/migrate"
	"github.com/nikoksr/dbench/ent/systemdetails"
	"github.com/nikoksr/dbench/internal/models"
)

var _ Store = (*DB)(nil)

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
func (db *DB) Connect(ctx context.Context, dsn string) (*DB, error) {
	db.connectOnce.Do(func() {
		// Open database connection
		db.client, db.err = ent.Open(dialect.SQLite, dsn)
		if db.err != nil {
			return
		}

		// If successful, run migration
		if err := db.client.Schema.Create(ctx,
			migrate.WithDropIndex(true),
			migrate.WithDropColumn(true),
		); err != nil {
			db.err = fmt.Errorf("create schema resources: %v", err)
		}

	})

	return db, db.err
}

// IsReady checks if the database client is ready.
func (db *DB) IsReady() error {
	if db.client == nil {
		return fmt.Errorf("database client is nil")
	}

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.client == nil {
		return fmt.Errorf("database client is nil")
	}

	return db.client.Close()
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}

// Count returns the count of benchmarks in the database.
func (db *DB) Count(ctx context.Context, options ...QueryOption) (int, error) {
	query := applyQueryOptions(db.client.Benchmark.Query(), options...)

	return query.Count(ctx)
}

// Save saves a benchmark to the database.
func (db *DB) Save(ctx context.Context, benchmark *models.Benchmark) (*models.Benchmark, error) {
	tx, err := db.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	_benchmark, err := tx.Benchmark.Create().
		SetGroupID(benchmark.GroupID).
		SetComment(benchmark.Comment).
		SetVersion(benchmark.Version).
		SetCommand(benchmark.Command).
		SetTransactionType(benchmark.TransactionType).
		SetScalingFactor(benchmark.ScalingFactor).
		SetQueryMode(benchmark.QueryMode).
		SetClients(benchmark.Clients).
		SetThreads(benchmark.Threads).
		SetNillableSystemID(benchmark.SystemID).
		Save(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("save benchmark: %w", err))
	}

	_, err = tx.BenchmarkResult.Create().
		SetBenchmarkID(_benchmark.ID).
		SetTransactions(benchmark.Edges.Result.Transactions).
		SetFailedTransactions(benchmark.Edges.Result.FailedTransactions).
		SetTransactionsPerSecond(benchmark.Edges.Result.TransactionsPerSecond).
		SetAverageLatency(benchmark.Edges.Result.AverageLatency).
		SetConnectionTime(benchmark.Edges.Result.ConnectionTime).
		SetTotalRuntime(benchmark.Edges.Result.TotalRuntime).
		Save(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("save benchmark result: %w", err))
	}

	_, err = tx.SystemMetric.Create().
		// Benchmark
		SetBenchmarkID(_benchmark.ID).
		// CPU
		SetCPUMinLoad(benchmark.Edges.SystemMetric.CPUMinLoad).
		SetCPUMaxLoad(benchmark.Edges.SystemMetric.CPUMaxLoad).
		SetCPUAverageLoad(benchmark.Edges.SystemMetric.CPUAverageLoad).
		SetCPU50thLoad(benchmark.Edges.SystemMetric.CPU50thLoad).
		SetCPU75thLoad(benchmark.Edges.SystemMetric.CPU75thLoad).
		SetCPU90thLoad(benchmark.Edges.SystemMetric.CPU90thLoad).
		SetCPU95thLoad(benchmark.Edges.SystemMetric.CPU95thLoad).
		SetCPU99thLoad(benchmark.Edges.SystemMetric.CPU99thLoad).
		// Memory
		SetMemoryMinLoad(benchmark.Edges.SystemMetric.MemoryMinLoad).
		SetMemoryMaxLoad(benchmark.Edges.SystemMetric.MemoryMaxLoad).
		SetMemoryAverageLoad(benchmark.Edges.SystemMetric.MemoryAverageLoad).
		SetMemory50thLoad(benchmark.Edges.SystemMetric.Memory50thLoad).
		SetMemory75thLoad(benchmark.Edges.SystemMetric.Memory75thLoad).
		SetMemory90thLoad(benchmark.Edges.SystemMetric.Memory90thLoad).
		SetMemory95thLoad(benchmark.Edges.SystemMetric.Memory95thLoad).
		SetMemory99thLoad(benchmark.Edges.SystemMetric.Memory99thLoad).
		Save(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("save system metrics: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return benchmark, nil
}

// Fetch fetches benchmarks from the database.
func (db *DB) Fetch(ctx context.Context, options ...QueryOption) ([]*models.Benchmark, error) {
	query := applyQueryOptions(db.client.Benchmark.Query(), options...)

	return query.
		WithResult().
		WithSystemMetric().
		WithSystem().
		All(ctx)
}

// FetchByIDs fetches benchmarks by their IDs.
func (db *DB) FetchByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Benchmark, error) {
	// Check if ids are given
	if len(ids) == 0 {
		return nil, fmt.Errorf("no benchmark ID provided")
	}

	// Convert string ids to pulid.ID
	pulids, err := convertToPULID(ids)
	if err != nil {
		return nil, err
	}

	// Add ID filter to options
	options = append(options, WithFilter(func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
		return query.Where(benchmark.IDIn(pulids...))
	}))

	// Fetch benchmarks
	return db.Fetch(ctx, options...)
}

// FetchByGroupIDs fetches benchmarks by their group IDs.
func (db *DB) FetchByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Benchmark, error) {
	// Check if ids are given
	if len(ids) == 0 {
		return nil, fmt.Errorf("no benchmark-group ID provided")
	}

	// Convert string ids to pulid.ID
	pulids, err := convertToPULID(ids)
	if err != nil {
		return nil, err
	}

	// Add ID filter to options
	options = append(options, WithFilter(func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
		return query.Where(benchmark.GroupIDIn(pulids...))
	}))

	// Fetch benchmarks
	return db.Fetch(ctx, options...)
}

// FetchGroupIDs fetches group IDs from the database.
func (db *DB) FetchGroupIDs(ctx context.Context, options ...QueryOption) ([]string, error) {
	if db.client == nil {
		return nil, fmt.Errorf("database client is nil")
	}

	query := applyQueryOptions(db.client.Benchmark.Query(), options...)

	return query.Select(benchmark.FieldGroupID).Strings(ctx)
}

// RemoveByIDs removes benchmarks by their IDs.
func (db *DB) RemoveByIDs(ctx context.Context, ids []string) error {
	// Convert string ids to pulid.ID
	pulids, err := convertToPULID(ids)
	if err != nil {
		return err
	}

	// Delete benchmarks
	_, err = db.client.Benchmark.Delete().
		Where(benchmark.IDIn(pulids...)).Exec(ctx)
	return err
}

// RemoveByGroupIDs removes benchmarks by their group IDs.
func (db *DB) RemoveByGroupIDs(ctx context.Context, ids []string) error {
	// Convert string ids to pulid.ID
	pulids, err := convertToPULID(ids)
	if err != nil {
		return err
	}

	// Delete benchmarks
	_, err = db.client.Benchmark.Delete().
		Where(benchmark.GroupIDIn(pulids...)).Exec(ctx)
	return err
}

// SaveSystemDetails saves system details to the database.
func (db *DB) SaveSystemDetails(ctx context.Context, systemDetails *models.SystemDetails) (*models.SystemDetails, error) {
	// Check if system details with machine ID already exists
	if systemDetails.MachineID != nil {
		_systemDetails, err := db.client.SystemDetails.Query().
			Where(systemdetails.MachineIDEQ(*systemDetails.MachineID)).
			Only(ctx)
		if err == nil {
			return _systemDetails, nil
		}
		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("check if system details already exists: %w", err)
		}
	}

	// No system details with machine ID found, create new
	_systemDetails, err := db.client.SystemDetails.Create().
		SetNillableMachineID(systemDetails.MachineID).
		SetNillableOsName(systemDetails.OsName).
		SetNillableOsArch(systemDetails.OsArch).
		SetNillableCPUVendor(systemDetails.CPUVendor).
		SetNillableCPUModel(systemDetails.CPUModel).
		SetNillableCPUCount(systemDetails.CPUCount).
		SetNillableCPUCores(systemDetails.CPUCores).
		SetNillableCPUThreads(systemDetails.CPUThreads).
		SetNillableRAMPhysical(systemDetails.RAMPhysical).
		SetNillableRAMUsable(systemDetails.RAMUsable).
		SetNillableDiskCount(systemDetails.DiskCount).
		SetNillableDiskSpaceTotal(systemDetails.DiskSpaceTotal).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return _systemDetails, nil
}
