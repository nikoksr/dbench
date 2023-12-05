package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/Masterminds/semver/v3"

	"github.com/nikoksr/dbench/ent/schema/datetime"
)

// AppConfig struct extends ent.Schema to describe the AppConfig database table.
type AppConfig struct {
	ent.Schema
}

// AppConfigMixin is a struct with embedded mixin.Schema (a predefined structure)
type AppConfigMixin struct {
	mixin.Schema
}

func validateAppVersion(v string) error {
	if v == "dev" {
		return nil // dev version is always valid
	}

	// Check if version is valid semver
	_, err := semver.NewVersion(v)

	return err
}

// Fields method defines the fields within the AppConfig database table.
func (AppConfigMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Unique().
			Immutable(),
		field.String("version").
			Validate(validateAppVersion).
			Default("dev").
			NotEmpty().
			Comment("Version of the app itself that created this schema. We use it to check if the schema is compatible with the current version of dbench and whether a migration is required."),
		field.String("initial_version").
			Validate(validateAppVersion).
			Optional().
			Default("dev").
			NotEmpty().
			Comment("Version of the app itself that originally created this schema."),
	}
}

// Mixin function defines the mixins to be incorporated into the AppConfig schema.
func (AppConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// The AppConfig itself
		AppConfigMixin{},
		// CreatedAt and UpdatedAt timestamps
		datetime.NewMixin(),
	}
}

// Edges function defines the relations/edges of the AppConfig schema.
func (AppConfig) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes function defines the indexed fields for faster queries on the AppConfig schema.
func (AppConfig) Indexes() []ent.Index {
	return []ent.Index{
		// Index the version field for faster queries
		index.Fields("version"),
	}
}

// Annotations function adds annotations to the AppConfig schema.
func (AppConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("AppConfig holds the configuration of the application. There can only be one AppConfig in the database. It's name is always 'self'."),
	}
}
