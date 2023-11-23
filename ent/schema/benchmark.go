package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/nikoksr/dbench/ent/schema/datetime"
	"github.com/nikoksr/dbench/ent/schema/pulid"
)

// Benchmark struct extends ent.Schema to describe the Benchmark database table.
type Benchmark struct {
	ent.Schema
}

// BenchmarkMixin is a struct with embedded mixin.Schema (a predefined structure)
type BenchmarkMixin struct {
	mixin.Schema
}

// Fields method defines the fields within the Benchmark database table.
func (BenchmarkMixin) Fields() []ent.Field {
	return []ent.Field{
		// The group this benchmark belongs to.
		field.String("group_id").
			GoType(pulid.ID("")).
			Immutable(),
		// An optional comment for the benchmark.
		field.String("comment").
			Optional(),
		// Remaining fields are optional and cannot be changed once set.
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
		// The information about the system this benchmark was run on. Optional, as this feature is opt-in.
		field.String("system_id").
			GoType(pulid.ID("")).
			Immutable().
			Optional().
			Nillable(),
	}
}

// Mixin function defines the mixins to be incorporated into the Benchmark schema.
func (Benchmark) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// Primary key using PULIDs
		pulid.NewMixinWithPrefix("id", "bmk"),
		// The Benchmark itself
		BenchmarkMixin{},
		// CreatedAt and UpdatedAt timestamps
		datetime.NewMixin(),
	}
}

// Edges function defines the relations/edges of the Benchmark schema.
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
		edge.From("system", SystemDetails.Type).
			Ref("benchmarks").
			Field("system_id").
			Unique().
			Immutable().
			Comment("The system this benchmark was run on."),
	}
}

// Indexes function defines the indexed fields for faster queries on the Benchmark schema.
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

// Annotations function adds annotations to the Benchmark schema.
func (Benchmark) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Benchmarks are the main entity of this application. They represent a single pgbench run. This table stores the config of the run."),
	}
}
