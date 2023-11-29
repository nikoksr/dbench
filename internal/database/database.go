package database

import (
	"context"
	"fmt"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/models"
)

type BenchmarkRepo interface {
	Save(ctx context.Context, res *models.Benchmark) (*models.Benchmark, error)
	Fetch(ctx context.Context, options ...QueryOption) ([]*models.Benchmark, error)
	FetchByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*models.Benchmark, error)
	FetchByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*models.Benchmark, error)
	RemoveByIDs(ctx context.Context, ids []string) error
	RemoveByGroupIDs(ctx context.Context, ids []string) error
	SaveSystemDetails(ctx context.Context, res *models.SystemDetails) (*models.SystemDetails, error)
	Close() error
}

type FilterFunc func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery

type QueryOption func(*QueryOptions)

type QueryOptions struct {
	// You can add more fields as necessary to handle different options
	OrderBy func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery
	Filters []FilterFunc
	Limit   int
	Offset  int
}

func WithOrderBy(orderFunc func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery) QueryOption {
	return func(opts *QueryOptions) {
		opts.OrderBy = orderFunc
	}
}

func WithFilter(filterFunc FilterFunc) QueryOption {
	return func(opts *QueryOptions) {
		opts.Filters = append(opts.Filters, filterFunc)
	}
}

func WithLimit(limit int) QueryOption {
	return func(opts *QueryOptions) {
		opts.Limit = limit
	}
}

func WithOffset(offset int) QueryOption {
	return func(opts *QueryOptions) {
		opts.Offset = offset
	}
}

func applyQueryOptions(query *ent.BenchmarkQuery, opts ...QueryOption) *ent.BenchmarkQuery {
	qo := &QueryOptions{} // Initialize with default options
	for _, opt := range opts {
		opt(qo) // Apply each option to the options
	}

	// Apply OrderBy, Limit, and Offset if they are set
	if qo.OrderBy != nil {
		query = qo.OrderBy(query)
	}
	if qo.Limit > 0 {
		query = query.Limit(qo.Limit)
	}
	if qo.Offset > 0 {
		query = query.Offset(qo.Offset)
	}

	// Apply Filters
	for _, filter := range qo.Filters {
		query = filter(query)
	}

	return query
}

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
