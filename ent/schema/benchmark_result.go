package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
)

type BenchmarkResult struct {
	ent.Schema
}

type BenchmarkResultMixin struct {
	mixin.Schema
}

func newDurationField(name string) ent.Field {
	return field.Other(name, duration.Duration(0)).
		SchemaType(map[string]string{
			dialect.SQLite:   "BIGINT",
			dialect.Postgres: "BIGINT",
			dialect.MySQL:    "BIGINT",
		}).
		Optional().
		Immutable()
}

func (BenchmarkResultMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("benchmark_id").
			GoType(pulid.ID("")).
			Immutable().
			Unique(),
		field.Int("transactions").
			Optional().
			Immutable(),
		field.Int("failed_transactions").
			Optional().
			Immutable(),
		field.Float("transactions_per_second").
			Optional().
			Immutable(),
		newDurationField("average_latency"),
		newDurationField("connection_time"),
		newDurationField("total_runtime"),
		field.Time("created_at").
			Default(now).
			Immutable(),
	}
}

func (BenchmarkResult) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pulid.NewMixinWithPrefix("id", "bmkres"),
		BenchmarkResultMixin{},
	}
}

func (BenchmarkResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("benchmark", Benchmark.Type).
			Ref("result").
			Field("benchmark_id").
			Unique().
			Required().
			Immutable().
			Comment("The benchmark (-config) that produced this result."),
	}
}

func (BenchmarkResult) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schema.Comment("Benchmark results are the produced and parsed results of a benchmark (pgbench) run. It is a one-to-one relation with the benchmark."),
	}
}
