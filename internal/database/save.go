package database

import (
	"context"
	"fmt"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/models"
)

func (db *DB) saveBenchmark(ctx context.Context, tx *ent.Tx, bmark *models.Benchmark) (*models.Benchmark, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	if bmark == nil {
		return nil, fmt.Errorf("benchmark is nil")
	}

	bmark, err := tx.Benchmark.Create().
		SetGroupID(bmark.GroupID).
		SetNillableComment(bmark.Comment).
		SetVersion(bmark.Version).
		SetCommand(bmark.Command).
		SetTransactionType(bmark.TransactionType).
		SetScalingFactor(bmark.ScalingFactor).
		SetQueryMode(bmark.QueryMode).
		SetClients(bmark.Clients).
		SetThreads(bmark.Threads).
		SetRecordedAt(bmark.RecordedAt).
		Save(ctx)

	return bmark, err
}

func (db *DB) saveBenchmarkResult(ctx context.Context, tx *ent.Tx, bmarkID pulid.ID, result *models.BenchmarkResult) (*models.BenchmarkResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	if result == nil {
		return nil, fmt.Errorf("benchmark result is nil")
	}

	result, err := tx.BenchmarkResult.Create().
		SetBenchmarkID(bmarkID).
		SetTransactions(result.Transactions).
		SetFailedTransactions(result.FailedTransactions).
		SetTransactionsPerSecond(result.TransactionsPerSecond).
		SetAverageLatency(result.AverageLatency).
		SetConnectionTime(result.ConnectionTime).
		SetTotalRuntime(result.TotalRuntime).
		Save(ctx)

	return result, err
}

func (db *DB) saveSystemMetric(ctx context.Context, tx *ent.Tx, bmarkID pulid.ID, metric *models.SystemMetric) (*models.SystemMetric, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	if metric == nil {
		return nil, fmt.Errorf("system metric is nil")
	}

	metric, err := tx.SystemMetric.Create().
		// Benchmark
		SetBenchmarkID(bmarkID).
		// CPU
		SetCPUMinLoad(metric.CPUMinLoad).
		SetCPUMaxLoad(metric.CPUMaxLoad).
		SetCPUAverageLoad(metric.CPUAverageLoad).
		SetCPU50thLoad(metric.CPU50thLoad).
		SetCPU75thLoad(metric.CPU75thLoad).
		SetCPU90thLoad(metric.CPU90thLoad).
		SetCPU95thLoad(metric.CPU95thLoad).
		SetCPU99thLoad(metric.CPU99thLoad).
		// Memory
		SetMemoryMinLoad(metric.MemoryMinLoad).
		SetMemoryMaxLoad(metric.MemoryMaxLoad).
		SetMemoryAverageLoad(metric.MemoryAverageLoad).
		SetMemory50thLoad(metric.Memory50thLoad).
		SetMemory75thLoad(metric.Memory75thLoad).
		SetMemory90thLoad(metric.Memory90thLoad).
		SetMemory95thLoad(metric.Memory95thLoad).
		SetMemory99thLoad(metric.Memory99thLoad).
		Save(ctx)

	return metric, err
}

func (db *DB) saveSystemConfig(ctx context.Context, tx *ent.Tx, bmarkID pulid.ID, systemConfig *models.SystemConfig) (*models.SystemConfig, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	if systemConfig == nil {
		return nil, fmt.Errorf("system config are nil")
	}

	systemConfig, err := tx.SystemConfig.Create().
		SetBenchmarkID(bmarkID).
		SetNillableMachineID(systemConfig.MachineID).
		SetNillableOsName(systemConfig.OsName).
		SetNillableOsArch(systemConfig.OsArch).
		SetNillableCPUVendor(systemConfig.CPUVendor).
		SetNillableCPUModel(systemConfig.CPUModel).
		SetNillableCPUCount(systemConfig.CPUCount).
		SetNillableCPUCores(systemConfig.CPUCores).
		SetNillableCPUThreads(systemConfig.CPUThreads).
		SetNillableRAMPhysical(systemConfig.RAMPhysical).
		SetNillableRAMUsable(systemConfig.RAMUsable).
		SetNillableDiskCount(systemConfig.DiskCount).
		SetNillableDiskSpaceTotal(systemConfig.DiskSpaceTotal).
		Save(ctx)

	return systemConfig, err
}

func (db *DB) save(ctx context.Context, tx *ent.Tx, bmark *models.Benchmark) (*models.Benchmark, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	if bmark == nil {
		return nil, fmt.Errorf("benchmark is nil")
	}

	// Now we can create the benchmark
	_bmark, err := db.saveBenchmark(ctx, tx, bmark)
	if err != nil {
		return nil, err
	}

	// Set benchmark ID for foreign key compliance
	bmark.ID = _bmark.ID

	// Save benchmark result
	if _, err = db.saveBenchmarkResult(ctx, tx, bmark.ID, bmark.Edges.Result); err != nil {
		return nil, fmt.Errorf("save benchmark result: %w", err)
	}

	// Save system metric
	if _, err = db.saveSystemMetric(ctx, tx, bmark.ID, bmark.Edges.SystemMetric); err != nil {
		return nil, fmt.Errorf("save system metric: %w", err)
	}

	// System config are optional, only save if they are given
	if bmark.Edges.System != nil {
		bmark.Edges.System, err = db.saveSystemConfig(ctx, tx, bmark.ID, bmark.Edges.System)
		if err != nil {
			return nil, fmt.Errorf("save system config: %w", err)
		}
	}

	return bmark, nil
}

// Save saves a benchmark to the database.
func (db *DB) Save(ctx context.Context, bmark *models.Benchmark) (*models.Benchmark, error) {
	tx, err := db.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	bmark, err = db.save(ctx, tx, bmark)
	if err != nil {
		return nil, rollback(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return bmark, nil
}

// SaveMany saves multiple benchmarks to the database.
func (db *DB) SaveMany(ctx context.Context, bmarks []*models.Benchmark) ([]*models.Benchmark, error) {
	tx, err := db.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	for idx, bmark := range bmarks {
		bmarks[idx], err = db.save(ctx, tx, bmark)
		if err != nil {
			return nil, rollback(tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return bmarks, nil
}
