package database

import (
	"context"
	"fmt"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/models"
)

// Store is an interface that defines the methods that a database store should implement.
type Store interface {
	Save(ctx context.Context, res *models.Benchmark) (*models.Benchmark, error)
	Fetch(ctx context.Context, options ...QueryOption) ([]*models.Benchmark, error)
	FetchByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*models.Benchmark, error)
	FetchByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*models.Benchmark, error)
	FetchGroupIDs(ctx context.Context, options ...QueryOption) ([]string, error)
	RemoveByIDs(ctx context.Context, ids []string) error
	RemoveByGroupIDs(ctx context.Context, ids []string) error
	SaveSystemDetails(ctx context.Context, res *models.SystemDetails) (*models.SystemDetails, error)
	Count(ctx context.Context, options ...QueryOption) (int, error)
}

// FilterFunc is a type that defines a function that modifies a BenchmarkQuery.
type FilterFunc func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery

// QueryOption is a type that defines a function that modifies a QueryOptions.
type QueryOption func(*QueryOptions)

// QueryOptions is a struct that holds options for a query.
type QueryOptions struct {
	OrderBy func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery
	Filters []FilterFunc
	Limit   int
	Offset  int
}

// WithOrderBy is a function that returns a QueryOption that sets the OrderBy field of a QueryOptions.
func WithOrderBy(orderFunc func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery) QueryOption {
	return func(opts *QueryOptions) {
		opts.OrderBy = orderFunc
	}
}

// WithFilter is a function that returns a QueryOption that adds a FilterFunc to the Filters field of a QueryOptions.
func WithFilter(filterFunc FilterFunc) QueryOption {
	return func(opts *QueryOptions) {
		opts.Filters = append(opts.Filters, filterFunc)
	}
}

// WithLimit is a function that returns a QueryOption that sets the Limit field of a QueryOptions.
func WithLimit(limit int) QueryOption {
	return func(opts *QueryOptions) {
		opts.Limit = limit
	}
}

// WithOffset is a function that returns a QueryOption that sets the Offset field of a QueryOptions.
func WithOffset(offset int) QueryOption {
	return func(opts *QueryOptions) {
		opts.Offset = offset
	}
}

// applyQueryOptions is a function that applies a list of QueryOptions to a BenchmarkQuery.
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
