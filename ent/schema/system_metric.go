package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/nikoksr/dbench/ent/schema/datetime"
	"github.com/nikoksr/dbench/ent/schema/pulid"
)

// SystemMetric struct extends ent.Schema, defines the SystemMetric table in the database.
type SystemMetric struct {
	ent.Schema
}

// SystemMetricMixin is a struct with embedded mixin.Schema.
type SystemMetricMixin struct {
	mixin.Schema
}

// newMetricField function creates a float field parameterized by the given name.
// This float field includes restrictions on the input value (positive only,
// minimum value of 0), so it seems it's used to represent some statistic measures.
func newMetricField(name string) ent.Field {
	return field.Float(name).
		// Define custom schema types for each dialect. Metrics would otherwise with a huge precision which is not needed.
		// We pick a precision of 7 and a scale of 2, since I'm not sure whether values could exceed 100.00.
		SchemaType(map[string]string{
			dialect.SQLite:   "real",
			dialect.Postgres: "numeric(7, 2)",
			dialect.MySQL:    "numeric(7, 2)",
		}).
		Positive().
		Min(0).
		Immutable()
}

// Fields method defines the fields within the SystemMetric database table.
func (SystemMetricMixin) Fields() []ent.Field {
	return []ent.Field{
		// The benchmark this system metric belongs to.
		field.String("benchmark_id").
			GoType(pulid.ID("")).
			Immutable().
			Unique(),
		// Other fields are optional and cannot be changed once set.
		// CPU
		newMetricField("cpu_min_load"),
		newMetricField("cpu_max_load"),
		newMetricField("cpu_average_load"),
		newMetricField("cpu_50th_load"),
		newMetricField("cpu_75th_load"),
		newMetricField("cpu_90th_load"),
		newMetricField("cpu_95th_load"),
		newMetricField("cpu_99th_load"),
		// Memory
		newMetricField("memory_min_load"),
		newMetricField("memory_max_load"),
		newMetricField("memory_average_load"),
		newMetricField("memory_50th_load"),
		newMetricField("memory_75th_load"),
		newMetricField("memory_90th_load"),
		newMetricField("memory_95th_load"),
		newMetricField("memory_99th_load"),
	}
}

// Mixin function defines the mixins to be incorporated into the SystemMetric schema.
func (SystemMetric) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// Primary key using PULIDs
		pulid.NewMixinWithPrefix("id", "smet"),
		// The SystemMetric itself
		SystemMetricMixin{},
		// CreatedAt and UpdatedAt timestamps
		datetime.NewMixin(),
	}
}

// Edges function defines the relations/edges of the SystemMetric schema.
func (SystemMetric) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("benchmark", Benchmark.Type).
			Ref("system_metric").
			Field("benchmark_id").
			Unique().
			Required().
			Immutable().
			Comment("The benchmark this system metric belong to."),
	}
}

// Annotations function adds annotations to the SystemMetric schema.
func (SystemMetric) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("System metrics are the system metrics of the host system we polled while the benchmark was running. It is a one-to-one relation to the benchmark table. We probe for CPU and memory usage while the benchmark is running, calculate the median and percentiles and store them here."),
		entsql.Annotation{Table: "system_metrics"},
		edge.Annotation{StructTag: `json:"-"`},
	}
}
