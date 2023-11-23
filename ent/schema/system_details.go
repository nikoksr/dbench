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

// SystemDetails struct extends ent.Schema, defines the SystemDetails table in the database.
type SystemDetails struct {
	ent.Schema
}

// SystemDetailsMixin is a struct with embedded mixin.Schema.
type SystemDetailsMixin struct {
	mixin.Schema
}

// Fields method defines the fields within the SystemDetails database table.
func (SystemDetailsMixin) Fields() []ent.Field {
	return []ent.Field{
		// The unique identifier of the machine this system info belongs to.
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

// Mixin function defines the mixins to be incorporated into the SystemDetails schema.
func (SystemDetails) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// Primary key using PULIDs
		pulid.NewMixinWithPrefix("id", "sdet"),
		// The SystemDetails itself
		SystemDetailsMixin{},
		// CreatedAt and UpdatedAt timestamps
		datetime.NewMixin(),
	}
}

// Edges function defines the relations/edges of the SystemDetails schema.
func (SystemDetails) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("benchmarks", Benchmark.Type).
			Comment("The benchmarks that were run on this system."),
	}
}

// Indexes function defines the indexed fields for faster queries on the SystemDetails schema.
func (SystemDetails) Indexes() []ent.Index {
	return []ent.Index{
		// Index the machine_id column for faster lookups on system details of a specific machine.
		//
		// Note: machine_id can't be unique, as it's nullable.
		index.Fields("machine_id"),
	}
}

// Annotations function adds annotations to the SystemDetails schema.
func (SystemDetails) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("SystemDetails holds detailed information about the system the benchmark was run on."),
	}
}
