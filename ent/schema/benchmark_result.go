package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/nikoksr/dbench/ent/schema/datetime"

	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
)

// BenchmarkResult struct that extend ent.Schema, defines the BenchmarkResult
// table in the database.
type BenchmarkResult struct {
	ent.Schema
}

// BenchmarkResultMixin is a struct with embedded mixin.Schema (a predefined structure).
type BenchmarkResultMixin struct {
	mixin.Schema
}

// newDurationField function for creating the duration field with different
// schema type depending on the SQL dialect (SQLite, Postgres, MySQL).
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

// Fields method that defines the fields within the BenchmarkResult database table.
func (BenchmarkResultMixin) Fields() []ent.Field {
	return []ent.Field{
		// The benchmark this result belongs to.
		field.String("benchmark_id").
			GoType(pulid.ID("")).
			Immutable().
			Unique(),
		// Other fields are optional and cannot be changed once set.
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
	}
}

// Mixin function that defines the mixins to be incorporated into the BenchmarkResult schema.
func (BenchmarkResult) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// Primary key using PULIDs
		pulid.NewMixinWithPrefix("id", "bmkres"),
		// The BenchmarkResult itself
		BenchmarkResultMixin{},
		// CreatedAt and UpdatedAt timestamps
		datetime.NewMixin(),
	}
}

// Edges function defines the relations/edges of the BenchmarkResult schema.
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

// Annotations function adds annotations to the BenchmarkResult schema.
func (BenchmarkResult) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schema.Comment("Benchmark results are the produced and parsed results of a benchmark (pgbench) run. It is a one-to-one relation with the benchmark."),
	}
}
