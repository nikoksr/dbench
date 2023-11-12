package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
)

var now = func() time.Time {
	return time.Now().UTC()
}

// Result holds the schema definition for the Result entity.
type Result struct {
	ent.Schema
}

// ResultMixin holds the schema definition for the Result entity.
type ResultMixin struct {
	mixin.Schema
}

// Fields of the Result.
func (ResultMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("version").
			Optional().
			Immutable(),
		field.String("command").
			Optional().
			Immutable(),
		field.String("transaction_type").
			Optional().
			Immutable(),
		field.Float("scaling_factor").
			Optional().
			Immutable(),
		field.String("query_mode").
			Optional().
			Immutable(),
		field.Int("clients").
			Optional().
			Immutable(),
		field.Int("threads").
			Optional().
			Immutable(),
		field.Int("transactions").
			Optional().
			Immutable(),
		field.Float("transactions_per_second").
			Optional().
			Immutable(),
		field.Int("failed_transactions").
			Optional().
			Immutable(),
		field.Other("average_latency", duration.Duration(0)).
			SchemaType(map[string]string{
				dialect.SQLite:   "BIGINT",
				dialect.Postgres: "BIGINT",
				dialect.MySQL:    "BIGINT",
			}).
			Optional().
			Immutable(),
		field.Other("initial_connection_time", duration.Duration(0)).
			SchemaType(map[string]string{
				dialect.SQLite:   "BIGINT",
				dialect.Postgres: "BIGINT",
				dialect.MySQL:    "BIGINT",
			}).
			Optional().
			Immutable(),
		field.Other("total_runtime", duration.Duration(0)).
			SchemaType(map[string]string{
				dialect.SQLite:   "BIGINT",
				dialect.Postgres: "BIGINT",
				dialect.MySQL:    "BIGINT",
			}).
			Optional().
			Immutable(),
		field.Time("created_at").
			Default(now).
			Immutable(),
	}
}

func (Result) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pulid.NewMixinWithPrefix("id", "res"),
		pulid.NewMixinWithPrefix("group_id", "resgrp"), // Does not actually generate the group_id column. Look at cmd/bench.go.
		ResultMixin{},
	}
}
