package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"image/jpeg"

	"github.com/gen2brain/heic"
	"github.com/google/uuid"
)

type Photo struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"originalName"`
	MimeType     string    `json:"mimeType"`
	UploadedAt   time.Time `json:"uploadedAt"`
	UploadedBy   string    `json:"uploadedBy"` // User ID who uploaded the photo
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

var storage Storage

// Generate a secure random session ID
func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to UUID if random fails
		return uuid.New().String()
	}
	return base64.URLEncoding.EncodeToString(b)
}

func main() {
	// Initialize HEIC decoder
	heic.Init()

	os.MkdirAll("uploads", 0755)
	os.MkdirAll("static", 0755)

	// Initialize storage
	var err error
	storage, err = NewFileStorage("storage.json")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/photos", handleGetPhotos)
	http.HandleFunc("/api/user/current", handleGetCurrentUser)
	http.HandleFunc("/api/user/create", handleCreateUser)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux)))
}

// CORS middleware to handle credentials
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, _ := getUserFromRequest(r)
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	r.ParseMultipartForm(32 << 20) // 32MB max

	file, header, err := r.FormFile("photo")
	if err != nil {
		log.Printf("Failed to get file: %v", err)
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	photoID := uuid.New().String()
	mimeType := header.Header.Get("Content-Type")
	filename := header.Filename
	ext := filepath.Ext(filename)

	log.Printf("Upload received - Filename: %s, MIME: %s, Extension: %s", filename, mimeType, ext)

	var outputPath string
	// Check if this is a HEIC file by extension (MIME type might be generic)
	isHEIC := strings.ToLower(ext) == ".heic" || strings.ToLower(ext) == ".heif"

	// Also check MIME type if available
	if !isHEIC && (mimeType == "image/heic" || mimeType == "image/heif") {
		isHEIC = true
	}

	if isHEIC {
		outputPath = filepath.Join("uploads", photoID+".jpg")

		// Read the HEIC file into memory
		heicData, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Failed to read HEIC data: %v", err)
			http.Error(w, "Failed to read HEIC file", http.StatusInternalServerError)
			return
		}

		log.Printf("Converting HEIC file: %s (size: %d bytes)", filename, len(heicData))

		// Decode HEIC from bytes
		img, err := heic.Decode(bytes.NewReader(heicData))
		if err != nil {
			log.Printf("Failed to decode HEIC: %v", err)
			http.Error(w, "Failed to decode HEIC file", http.StatusInternalServerError)
			return
		}

		outFile, err := os.Create(outputPath)
		if err != nil {
			log.Printf("Failed to create output file: %v", err)
			http.Error(w, "Failed to create output file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90}); err != nil {
			log.Printf("Failed to encode JPEG: %v", err)
			http.Error(w, "Failed to encode JPEG", http.StatusInternalServerError)
			return
		}

		log.Printf("HEIC conversion successful: %s -> %s", filename, outputPath)

		mimeType = "image/jpeg"
	} else {
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		outputPath = filepath.Join("uploads", photoID+ext)

		dst, err := os.Create(outputPath)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
	}

	photo := Photo{
		ID:           photoID,
		OriginalName: header.Filename,
		MimeType:     mimeType,
		UploadedAt:   time.Now(),
		UploadedBy:   user.ID,
	}

	if err := storage.AddPhoto(photo); err != nil {
		log.Printf("Failed to save photo metadata: %v", err)
		http.Error(w, "Failed to save photo metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("Upload successful: %s (ID: %s)", filename, photoID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photo)
}

func handleGetPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	photos, err := storage.GetAllPhotos()
	if err != nil {
		log.Printf("Failed to get photos: %v", err)
		http.Error(w, "Failed to get photos", http.StatusInternalServerError)
		return
	}

	// Enrich photos with user information
	type PhotoWithUser struct {
		Photo
		UploaderName string `json:"uploaderName"`
	}

	enrichedPhotos := make([]PhotoWithUser, 0, len(photos))
	for _, photo := range photos {
		enriched := PhotoWithUser{Photo: photo}
		if user, err := storage.GetUser(photo.UploadedBy); err == nil {
			enriched.UploaderName = user.Name
		}
		enrichedPhotos = append(enrichedPhotos, enriched)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrichedPhotos)
}

// Get user from session cookie
func getUserFromRequest(r *http.Request) (*User, *Session) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, nil
	}

	session, err := storage.GetSession(cookie.Value)
	if err != nil {
		return nil, nil
	}

	user, err := storage.GetUser(session.UserID)
	if err != nil {
		return nil, nil
	}

	return &user, &session
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := getUserFromRequest(r)

	response := struct {
		User *User `json:"user"`
	}{
		User: user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Create user
	user := User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	if err := storage.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session
	session := Session{
		ID:        generateSessionID(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
	}

	if err := storage.CreateSession(session); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
