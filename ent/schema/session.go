package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		// Session ID is a base64-encoded string, not a UUID
		field.String("id").
			Unique().
			Immutable().
			StorageKey("id"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			StorageKey("created_at"),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		// A session belongs to a user
		edge.From("user", User.Type).
			Ref("sessions").
			Unique().
			Required(),
	}
}
