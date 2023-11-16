package datetime

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"time"
)

// NowUTC returns the current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// NewMixin creates a Mixin that includes create_at and updated_at
func NewMixin() *Mixin {
	return &Mixin{}
}

// Mixin defines an ent Mixin
type Mixin struct {
	mixin.Schema
}

// Fields provides the created_at and updated_at field.
func (m Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(NowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(NowUTC).
			UpdateDefault(NowUTC),
	}
}
