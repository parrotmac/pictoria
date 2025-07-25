package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/parrotmac/pictorial/ent"
	"github.com/parrotmac/pictorial/ent/migrate"
)

type StorageData struct {
	Photos   []Photo            `json:"photos"`
	Users    map[string]User    `json:"users"`
	Sessions map[string]Session `json:"sessions"`
	Version  int                `json:"version"`
}

type Photo struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"originalName"`
	MimeType     string    `json:"mimeType"`
	UploadedAt   time.Time `json:"uploadedAt"`
	UploadedBy   string    `json:"uploadedBy"`
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}

func main() {
	var (
		databaseURL = flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
		jsonFile    = flag.String("json-file", "storage.json", "Path to JSON storage file to import")
		importData  = flag.Bool("import", false, "Import data from JSON file")
	)
	flag.Parse()

	if *databaseURL == "" {
		log.Fatal("database-url is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", *databaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create ent client
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	// Run migrations
	log.Println("Running database migrations...")
	if err := client.Schema.Create(ctx, migrate.WithForeignKeys(true)); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Import data if requested
	if *importData {
		log.Printf("Importing data from %s...\n", *jsonFile)
		if err := importFromJSON(ctx, client, *jsonFile); err != nil {
			log.Fatalf("failed to import data: %v", err)
		}
		log.Println("Data import completed successfully")
	}
}

func importFromJSON(ctx context.Context, client *ent.Client, filename string) error {
	// Read JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var storage StorageData
	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Import users first
	log.Printf("Importing %d users...\n", len(storage.Users))
	userMap := make(map[string]uuid.UUID)
	for _, user := range storage.Users {
		userUUID, err := uuid.Parse(user.ID)
		if err != nil {
			log.Printf("Skipping user with invalid UUID: %s", user.ID)
			continue
		}
		
		_, err = client.User.Create().
			SetID(userUUID).
			SetName(user.Name).
			SetCreatedAt(user.CreatedAt).
			Save(ctx)
		
		if err != nil {
			log.Printf("Failed to import user %s: %v", user.ID, err)
		} else {
			userMap[user.ID] = userUUID
			log.Printf("Imported user: %s", user.Name)
		}
	}

	// Import sessions
	log.Printf("Importing %d sessions...\n", len(storage.Sessions))
	for _, session := range storage.Sessions {
		userUUID, ok := userMap[session.UserID]
		if !ok {
			log.Printf("Skipping session %s: user %s not found", session.ID, session.UserID)
			continue
		}

		_, err = client.Session.Create().
			SetID(session.ID).
			SetCreatedAt(session.CreatedAt).
			SetUserID(userUUID).
			Save(ctx)
		
		if err != nil {
			log.Printf("Failed to import session %s: %v", session.ID, err)
		}
	}

	// Import photos
	log.Printf("Importing %d photos...\n", len(storage.Photos))
	for _, photo := range storage.Photos {
		photoUUID, err := uuid.Parse(photo.ID)
		if err != nil {
			log.Printf("Skipping photo with invalid UUID: %s", photo.ID)
			continue
		}

		userUUID, ok := userMap[photo.UploadedBy]
		if !ok {
			log.Printf("Skipping photo %s: user %s not found", photo.ID, photo.UploadedBy)
			continue
		}

		_, err = client.Photo.Create().
			SetID(photoUUID).
			SetOriginalName(photo.OriginalName).
			SetMimeType(photo.MimeType).
			SetUploadedAt(photo.UploadedAt).
			SetUploaderID(userUUID).
			Save(ctx)
		
		if err != nil {
			log.Printf("Failed to import photo %s: %v", photo.ID, err)
		} else {
			log.Printf("Imported photo: %s", photo.OriginalName)
		}
	}

	return nil
}