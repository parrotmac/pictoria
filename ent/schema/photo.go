package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Photo holds the schema definition for the Photo entity.
type Photo struct {
	ent.Schema
}

// Fields of the Photo.
func (Photo) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("original_name").
			NotEmpty().
			StorageKey("original_name"),
		field.String("mime_type").
			NotEmpty().
			StorageKey("mime_type"),
		field.Time("uploaded_at").
			Default(time.Now).
			Immutable().
			StorageKey("uploaded_at"),
	}
}

// Edges of the Photo.
func (Photo) Edges() []ent.Edge {
	return []ent.Edge{
		// A photo belongs to a user
		edge.From("uploader", User.Type).
			Ref("photos").
			Unique().
			Required(),
	}
}
