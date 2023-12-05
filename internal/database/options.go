package database

import (
	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
)

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
	} else {
		// No order function provided -> use default order
		query = query.Order(
			ent.Desc(benchmark.FieldGroupID),
			ent.Desc(benchmark.FieldID),
		)
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
