package database

import (
	"context"

	"github.com/nikoksr/dbench/ent/benchmark"
)

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
