package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/nikoksr/dbench/ent/schema/pulid"
)

type SystemMetric struct {
	ent.Schema
}

type SystemMetricMixin struct {
	mixin.Schema
}

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
		Optional().
		Immutable()
}

func (SystemMetricMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("benchmark_id").
			GoType(pulid.ID("")).
			Immutable().
			Unique(),
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
		// Misc
		field.Time("created_at").
			Default(now).
			Immutable(),
	}
}

func (SystemMetric) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pulid.NewMixinWithPrefix("id", "smet"),
		SystemMetricMixin{},
	}
}

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

func (SystemMetric) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "system_metrics"},
		schema.Comment("System metrics are the system metrics of the host system we polled while the benchmark was running. It is a one-to-one relation to the benchmark table. We probe for CPU and memory usage while the benchmark is running, calculate the median and percentiles and store them here."),
	}
}
