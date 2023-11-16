package database

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	_ "github.com/xiaoqidun/entps" // Modernc wrapper for ent

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
	"github.com/nikoksr/dbench/pkg/models"
)

var _ Database = (*EntDatabase)(nil) // Ensure that EntDatabase implements the Database interface

type EntDatabase struct {
	client *ent.Client
}

func NewEntDatabase(ctx context.Context, dsn string) (*EntDatabase, error) {
	client, err := ent.Open(dialect.SQLite, dsn)
	if err != nil {
		return nil, err
	}

	if err := client.Schema.Create(ctx); err != nil {
		return nil, fmt.Errorf("create schema resources: %v", err)
	}

	return &EntDatabase{client: client}, nil
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}

func (db *EntDatabase) SaveBenchmark(ctx context.Context, benchmark *models.Benchmark) error {
	tx, err := db.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	_benchmark, err := tx.Benchmark.Create().
		SetGroupID(benchmark.GroupID).
		SetVersion(benchmark.Version).
		SetCommand(benchmark.Command).
		SetTransactionType(benchmark.TransactionType).
		SetScalingFactor(benchmark.ScalingFactor).
		SetQueryMode(benchmark.QueryMode).
		SetClients(benchmark.Clients).
		SetThreads(benchmark.Threads).
		Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("save benchmark: %w", err))
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
		return rollback(tx, fmt.Errorf("save benchmark result: %w", err))
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
		return rollback(tx, fmt.Errorf("save system metrics: %w", err))
	}

	return tx.Commit()
}

func (db *EntDatabase) FetchBenchmarks(ctx context.Context, options ...QueryOption) ([]*models.Benchmark, error) {
	query, err := applyQueryOptions(db.client.Benchmark.Query(), options...)
	if err != nil {
		return nil, err
	}

	return query.
		WithResult().
		WithSystemMetric().
		All(ctx)
}

func (db *EntDatabase) FetchBenchmarksByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Benchmark, error) {
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
	return db.FetchBenchmarks(ctx, options...)
}

func (db *EntDatabase) FetchBenchmarksByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Benchmark, error) {
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
	return db.FetchBenchmarks(ctx, options...)
}

func (db *EntDatabase) Close() error {
	return db.client.Close()
}
