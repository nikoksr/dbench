package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/nikoksr/dbench/ent/schema/pulid"
)

type Benchmark struct {
	ent.Schema
}

type BenchmarkMixin struct {
	mixin.Schema
}

func (BenchmarkMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("group_id").
			GoType(pulid.ID("")).
			Immutable(),
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
		field.Time("created_at").
			Default(now).
			Immutable(),
		field.Time("updated_at").
			Default(now).
			UpdateDefault(now),
	}
}

func (Benchmark) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pulid.NewMixinWithPrefix("id", "bmk"),
		BenchmarkMixin{},
	}
}

func (Benchmark) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("result", BenchmarkResult.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)).
			Comment("The result produced by the benchmark run."),
		edge.To("system_metric", SystemMetric.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)).
			Comment("The metrics that we collected from the system during the benchmark run."),
	}
}

func (Benchmark) Indexes() []ent.Index {
	return []ent.Index{
		// Index the group_id column for faster lookups
		// on benchmarks of a specific benchmark-group.
		index.Fields("group_id"),
		// Index the clients column for faster lookups
		// on benchmarks with a comparable number of clients.
		index.Fields("clients"),
	}
}

func (Benchmark) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Benchmarks are the main entity of this application. They represent a single pgbench run. This table stores the config of the run."),
	}
}
