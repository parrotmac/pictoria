package main

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/parrotmac/pictorial/ent"
	"github.com/parrotmac/pictorial/ent/photo"
	"github.com/parrotmac/pictorial/ent/session"
	"github.com/parrotmac/pictorial/ent/user"
)

type PostgresStorage struct {
	client *ent.Client
}

func NewPostgresStorage(databaseURL string) (Storage, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create an ent driver from database/sql
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	// Run auto migration
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &PostgresStorage{client: client}, nil
}

// Photo operations
func (ps *PostgresStorage) AddPhoto(p Photo) error {
	ctx := context.Background()
	
	userUUID, err := uuid.Parse(p.UploadedBy)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	photoUUID, err := uuid.Parse(p.ID)
	if err != nil {
		return fmt.Errorf("invalid photo ID: %w", err)
	}

	_, err = ps.client.Photo.Create().
		SetID(photoUUID).
		SetOriginalName(p.OriginalName).
		SetMimeType(p.MimeType).
		SetUploadedAt(p.UploadedAt).
		SetUploaderID(userUUID).
		Save(ctx)

	return err
}

func (ps *PostgresStorage) GetPhoto(id string) (Photo, error) {
	ctx := context.Background()
	
	photoUUID, err := uuid.Parse(id)
	if err != nil {
		return Photo{}, fmt.Errorf("invalid photo ID: %w", err)
	}

	p, err := ps.client.Photo.Query().
		Where(photo.ID(photoUUID)).
		WithUploader().
		Only(ctx)
	
	if err != nil {
		if ent.IsNotFound(err) {
			return Photo{}, ErrNotFound
		}
		return Photo{}, err
	}

	return Photo{
		ID:           p.ID.String(),
		OriginalName: p.OriginalName,
		MimeType:     p.MimeType,
		UploadedAt:   p.UploadedAt,
		UploadedBy:   p.Edges.Uploader.ID.String(),
	}, nil
}

func (ps *PostgresStorage) GetAllPhotos() ([]Photo, error) {
	ctx := context.Background()
	
	photos, err := ps.client.Photo.Query().
		WithUploader().
		All(ctx)
	
	if err != nil {
		return nil, err
	}

	result := make([]Photo, len(photos))
	for i, p := range photos {
		result[i] = Photo{
			ID:           p.ID.String(),
			OriginalName: p.OriginalName,
			MimeType:     p.MimeType,
			UploadedAt:   p.UploadedAt,
			UploadedBy:   p.Edges.Uploader.ID.String(),
		}
	}

	return result, nil
}

func (ps *PostgresStorage) DeletePhoto(id string) error {
	ctx := context.Background()
	
	photoUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid photo ID: %w", err)
	}

	count, err := ps.client.Photo.Delete().
		Where(photo.ID(photoUUID)).
		Exec(ctx)
	
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNotFound
	}

	return nil
}

// User operations
func (ps *PostgresStorage) CreateUser(u User) error {
	ctx := context.Background()
	
	userUUID, err := uuid.Parse(u.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	_, err = ps.client.User.Create().
		SetID(userUUID).
		SetName(u.Name).
		SetCreatedAt(u.CreatedAt).
		Save(ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			return ErrExists
		}
		return err
	}

	return nil
}

func (ps *PostgresStorage) GetUser(id string) (User, error) {
	ctx := context.Background()
	
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return User{}, fmt.Errorf("invalid user ID: %w", err)
	}

	u, err := ps.client.User.Query().
		Where(user.ID(userUUID)).
		Only(ctx)
	
	if err != nil {
		if ent.IsNotFound(err) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	return User{
		ID:        u.ID.String(),
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (ps *PostgresStorage) GetAllUsers() ([]User, error) {
	ctx := context.Background()
	
	users, err := ps.client.User.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]User, len(users))
	for i, u := range users {
		result[i] = User{
			ID:        u.ID.String(),
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
		}
	}

	return result, nil
}

func (ps *PostgresStorage) UpdateUser(u User) error {
	ctx := context.Background()
	
	userUUID, err := uuid.Parse(u.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	count, err := ps.client.User.Update().
		Where(user.ID(userUUID)).
		SetName(u.Name).
		Save(ctx)
	
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNotFound
	}

	return nil
}

func (ps *PostgresStorage) DeleteUser(id string) error {
	ctx := context.Background()
	
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	count, err := ps.client.User.Delete().
		Where(user.ID(userUUID)).
		Exec(ctx)
	
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNotFound
	}

	return nil
}

// Session operations
func (ps *PostgresStorage) CreateSession(s Session) error {
	ctx := context.Background()
	
	userUUID, err := uuid.Parse(s.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	_, err = ps.client.Session.Create().
		SetID(s.ID).
		SetCreatedAt(s.CreatedAt).
		SetUserID(userUUID).
		Save(ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			return ErrExists
		}
		return err
	}

	return nil
}

func (ps *PostgresStorage) GetSession(id string) (Session, error) {
	ctx := context.Background()
	
	s, err := ps.client.Session.Query().
		Where(session.ID(id)).
		WithUser().
		Only(ctx)
	
	if err != nil {
		if ent.IsNotFound(err) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}

	return Session{
		ID:        s.ID,
		UserID:    s.Edges.User.ID.String(),
		CreatedAt: s.CreatedAt,
	}, nil
}

func (ps *PostgresStorage) DeleteSession(id string) error {
	ctx := context.Background()
	
	count, err := ps.client.Session.Delete().
		Where(session.ID(id)).
		Exec(ctx)
	
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNotFound
	}

	return nil
}

func (ps *PostgresStorage) GetAllSessions() ([]Session, error) {
	ctx := context.Background()
	
	sessions, err := ps.client.Session.Query().
		WithUser().
		All(ctx)
	
	if err != nil {
		return nil, err
	}

	result := make([]Session, len(sessions))
	for i, s := range sessions {
		result[i] = Session{
			ID:        s.ID,
			UserID:    s.Edges.User.ID.String(),
			CreatedAt: s.CreatedAt,
		}
	}

	return result, nil
}

// Persistence operations
func (ps *PostgresStorage) Save() error {
	// No-op for PostgreSQL - data is persisted immediately
	return nil
}

func (ps *PostgresStorage) Load() error {
	// No-op for PostgreSQL - data is loaded on demand
	return nil
}