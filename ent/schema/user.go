package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),
		field.String("name").
			NotEmpty().
			StorageKey("name"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			StorageKey("created_at"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// A user can have many photos
		edge.To("photos", Photo.Type),
		// A user can have many sessions
		edge.To("sessions", Session.Type),
	}
}
