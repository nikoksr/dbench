package database

import (
	"context"
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
	"github.com/nikoksr/dbench/internal/models"
)

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
	query := applyQueryOptions(db.client.Benchmark.Query(), options...)

	return query.Select(benchmark.FieldGroupID).Strings(ctx)
}

// Count returns the count of benchmarks in the database.
func (db *DB) Count(ctx context.Context, options ...QueryOption) (uint64, error) {
	query := applyQueryOptions(db.client.Benchmark.Query(), options...)
	count, err := query.Select(benchmark.FieldID).Count(ctx)

	return uint64(count), err
}

type counter interface {
	Count(ctx context.Context) (int, error)
}

// CountAll returns the count of all records in the database.
func (db *DB) CountAll(ctx context.Context) (uint64, error) {
	eg, ctx := errgroup.WithContext(ctx)

	// 4, so that BenchmarkResult is queried when the smaller queries are done. BenchmarkResult is the largest table.
	// eg.SetLimit(4)

	queries := []counter{
		db.client.AppConfig.Query(),
		db.client.SystemMetric.Query(),
		db.client.SystemConfig.Query(),
		db.client.Benchmark.Query(),
		db.client.BenchmarkResult.Query(),
	}

	var count atomic.Uint64

	for _, query := range queries {
		query := query
		eg.Go(func() error {
			c, err := query.Count(ctx)
			if err != nil {
				return err
			}
			count.Add(uint64(c))
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return 0, err
	}

	return count.Load(), nil
}
