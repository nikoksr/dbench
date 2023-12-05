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

// SystemConfig struct extends ent.Schema, defines the SystemConfig table in the database.
type SystemConfig struct {
	ent.Schema
}

// SystemConfigMixin is a struct with embedded mixin.Schema.
type SystemConfigMixin struct {
	mixin.Schema
}

// Fields method defines the fields within the SystemConfig database table.
func (SystemConfigMixin) Fields() []ent.Field {
	return []ent.Field{
		// The benchmark this result belongs to.
		field.String("benchmark_id").
			GoType(pulid.ID("")).
			Immutable().
			Unique(),
		// The unique identifier of the machine this system info belongs to. We can't set this field as unique,
		// because it's nullable.
		field.String("machine_id").
			Optional().
			Nillable().
			Immutable().
			NotEmpty(),
		// The following fields are optional and contain information about the system.
		field.String("os_name").
			Optional().
			Nillable().
			NotEmpty(),
		field.String("os_arch").
			Optional().
			Nillable().
			NotEmpty(),
		field.String("cpu_vendor").
			Optional().
			Nillable().
			NotEmpty().
			Comment("In case of a multi-CPU system, this is the vendor of the first CPU."),
		field.String("cpu_model").
			Optional().
			Nillable().
			NotEmpty().
			Comment("In case of a multi-CPU system, this is the model of the first CPU."),
		field.Uint32("cpu_count").
			Optional().
			Nillable().
			Positive(),
		field.Uint32("cpu_cores").
			Optional().
			Nillable().
			Positive(),
		field.Uint32("cpu_threads").
			Optional().
			Nillable().
			Positive(),
		field.Uint64("ram_physical").
			Optional().
			Nillable().
			Positive().
			Comment("The amount of RAM in bytes that is contained in all memory banks (DIMMs) that are attached to the motherboard. This is the total amount of RAM that is physically installed in the system."),
		field.Uint64("ram_usable").
			Optional().
			Nillable().
			Positive().
			Comment("The amount of physical RAM in bytes that is usable by the operating system. This is the total amount of RAM that is physically installed in the system minus the amount of RAM that is reserved for the system."),
		field.Uint32("disk_count").
			Optional().
			Nillable().
			Positive(),
		field.Uint64("disk_space_total").
			Optional().
			Nillable().
			Positive(),
	}
}

// Mixin function defines the mixins to be incorporated into the SystemConfig schema.
func (SystemConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// Primary key using PULIDs
		pulid.NewMixinWithPrefix("id", "sdet"),
		// The SystemConfig itself
		SystemConfigMixin{},
		// CreatedAt and UpdatedAt timestamps
		datetime.NewMixin(),
	}
}

// Edges function defines the relations/edges of the SystemConfig schema.
func (SystemConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("benchmark", Benchmark.Type).
			Ref("system").
			Field("benchmark_id").
			Unique().
			Required().
			Immutable().
			Comment("The benchmark that was run against this system config."),
	}
}

// Indexes function defines the indexed fields for faster queries on the SystemConfig schema.
func (SystemConfig) Indexes() []ent.Index {
	return []ent.Index{
		// Index the machine_id column for faster lookups on system config of a specific machine.
		//
		// Note: machine_id can't be unique, as it's nullable.
		index.Fields("machine_id"),
	}
}

// Annotations function adds annotations to the SystemConfig schema.
func (SystemConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("SystemConfig holds detailed information about the system the benchmark was run on."),
		edge.Annotation{StructTag: `json:"-"`},
	}
}
