package pulid

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// NewMixinWithPrefix creates a Mixin that encodes the provided prefix.
func NewMixinWithPrefix(columnName, prefix string) *Mixin {
	return &Mixin{
		columnName: columnName,
		prefix:     prefix,
	}
}

// Mixin defines an ent Mixin that captures the PULID prefix for a type.
type Mixin struct {
	mixin.Schema
	columnName string
	prefix     string
}

// Fields provides the id field.
func (m Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.String(m.columnName).
			GoType(ID("")).
			DefaultFunc(func() ID { return MustNew(m.prefix) }).
			Unique().
			Immutable(),
	}
}

// Annotation captures the id prefix for a type.
type Annotation struct {
	Prefix string
}

// Name implements the ent Annotation interface.
func (a Annotation) Name() string {
	return "PULID"
}

// Annotations returns the annotations for a Mixin instance.
func (m Mixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		Annotation{Prefix: m.prefix},
	}
}
