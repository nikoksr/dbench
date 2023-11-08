package database

import (
	"context"
	"fmt"
	"github.com/nikoksr/dbench/ent/result"

	"entgo.io/ent/dialect"
	_ "github.com/xiaoqidun/entps" // Modernc wrapper for ent

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/pkg/models"
)

var _ Database = (*EntDatabase)(nil) // Ensure that EntDatabase implements the Database interface

type EntDatabase struct {
	client *ent.Client
}

func NewEntDatabase(ctx context.Context, dsn string) (Database, error) {
	client, err := ent.Open(dialect.SQLite, dsn)
	if err != nil {
		return nil, err
	}

	if err := client.Schema.Create(ctx); err != nil {
		return nil, fmt.Errorf("create schema resources: %v", err)
	}

	return &EntDatabase{client: client}, nil
}

func (db *EntDatabase) SaveResult(ctx context.Context, res *models.Result) error {
	_, err := db.client.Result.Create().
		SetGroupID(res.GroupID).
		SetVersion(res.Version).
		SetCommand(res.Command).
		SetTransactionType(res.TransactionType).
		SetScalingFactor(res.ScalingFactor).
		SetQueryMode(res.QueryMode).
		SetClients(res.Clients).
		SetThreads(res.Threads).
		SetTransactions(res.Transactions).
		SetTransactionsPerSecond(res.TransactionsPerSecond).
		SetTransactionsPerClient(res.TransactionsPerClient).
		SetFailedTransactions(res.FailedTransactions).
		SetAverageLatency(res.AverageLatency).
		SetInitialConnectionTime(res.InitialConnectionTime).
		SetTotalRuntime(res.TotalRuntime).
		Save(ctx)

	return err
}

func (db *EntDatabase) FetchResults(ctx context.Context, options ...QueryOption) ([]*models.Result, error) {
	query, err := applyQueryOptions(db.client.Result.Query(), options...)
	if err != nil {
		return nil, err
	}

	return query.All(ctx)
}

func (db *EntDatabase) FetchResultsByIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Result, error) {
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
	options = append(options, WithFilter(func(query *ent.ResultQuery) *ent.ResultQuery {
		return query.Where(result.IDIn(pulids...))
	}))

	// Fetch results
	return db.FetchResults(ctx, options...)
}

func (db *EntDatabase) FetchResultsByGroupIDs(ctx context.Context, ids []string, options ...QueryOption) ([]*ent.Result, error) {
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
	options = append(options, WithFilter(func(query *ent.ResultQuery) *ent.ResultQuery {
		return query.Where(result.GroupIDIn(pulids...))
	}))

	// Fetch results
	return db.FetchResults(ctx, options...)
}

func (db *EntDatabase) Close() error {
	return db.client.Close()
}
