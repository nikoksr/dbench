package store

import (
	"context"
	"fmt"
	"github.com/nikoksr/dbench/ent/systemdetails"

	"entgo.io/ent/dialect"
	_ "github.com/xiaoqidun/entps" // Modernc wrapper for ent

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
	"github.com/nikoksr/dbench/ent/migrate"
	"github.com/nikoksr/dbench/internal/models"
)

var _ benchmarkStore = (*entStore)(nil) // Ensure that entStore implements the Store interface

type entStore struct {
	client *ent.Client
}

func newEntStore(ctx context.Context, dsn string) (*entStore, error) {
	client, err := ent.Open(dialect.SQLite, dsn)
	if err != nil {
		return nil, err
	}

	if err := client.Schema.Create(ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		return nil, fmt.Errorf("create schema resources: %v", err)
	}

	return &entStore{client: client}, nil
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}

func (db *entStore) Save(ctx context.Context, benchmark *models.Benchmark) (*models.Benchmark, error) {
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

func (db *entStore) Fetch(ctx context.Context, options ...QueryOption) ([]*models.Benchmark, error) {
	query, err := applyQueryOptions(db.client.Benchmark.Query(), options...)
	if err != nil {
		return nil, err
	}

	return query.
		WithResult().
		WithSystemMetric().
		WithSystem().
		All(ctx)
}

func (db *entStore) FetchByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Benchmark, error) {
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

func (db *entStore) FetchByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Benchmark, error) {
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

func (db *entStore) RemoveByIDs(ctx context.Context, ids []string) error {
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

func (db *entStore) RemoveByGroupIDs(ctx context.Context, ids []string) error {
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

func (db *entStore) SaveSystemDetails(ctx context.Context, systemDetails *models.SystemDetails) (*models.SystemDetails, error) {
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

func (db *entStore) Close() error {
	return db.client.Close()
}
