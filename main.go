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

	"tailscale.com/tsnet"
	"tailscale.com/types/logger"
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
	os.MkdirAll("uploads", 0755)
	os.MkdirAll("static", 0755)

	// Initialize storage based on environment
	storageType := os.Getenv("STORAGE_TYPE")

	var err error
	switch storageType {
	case "postgres":
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			log.Fatal("DATABASE_URL environment variable is required when STORAGE_TYPE=postgres")
		}
		storage, err = NewPostgresStorage(databaseURL)
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL storage: %v", err)
		}
		fmt.Println("Using PostgreSQL storage")
	default:
		storage, err = NewFileStorage("storage.json")
		if err != nil {
			log.Fatalf("Failed to initialize file storage: %v", err)
		}
		fmt.Println("Using file-based storage")
	}

	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/direct-networking", handleDirectNetworkRouting)
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/photos", handleGetPhotos)
	http.HandleFunc("/api/user/current", handleGetCurrentUser)
	http.HandleFunc("/api/user/create", handleCreateUser)

	// Handle photo deletion with pattern matching
	http.HandleFunc("/api/photos/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/photos/") {
			handleDeletePhoto(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/dist/assets"))))
	http.Handle("/gallery", http.StripPrefix("/gallery", http.FileServer(http.Dir("frontend/dist"))))
	http.Handle("/", http.FileServer(http.Dir("frontend/dist")))

	tailscaleFunnelHostname := strings.TrimSpace(os.Getenv("TAILSCALE_FUNNEL_HOSTNAME"))
	log.Printf("TAILSCALE_FUNNEL_HOSTNAME=%s", tailscaleFunnelHostname)

	tailscaleStateLocation := strings.TrimSpace(os.Getenv("TAILSCALE_STATE_LOCATION"))
	if tailscaleStateLocation == "" {
		tailscaleStateLocation = "./ts-funnel-config"
	}
	log.Printf("TAILSCALE_STATE_LOCATION=%s", tailscaleStateLocation)

	var tailscaleLogger logger.Logf
	tailscaleEnableVerboseLogs := strings.TrimSpace(os.Getenv("TAILSCALE_VERBOSE_LOGS"))
	if tailscaleEnableVerboseLogs == "true" || tailscaleEnableVerboseLogs == "1" {
		tailscaleLogger = log.Printf
	}

	s := &tsnet.Server{
		Dir:          tailscaleStateLocation,
		Hostname:     tailscaleFunnelHostname,
		Logf:         tailscaleLogger,
		RunWebClient: true,
	}
	defer s.Close()

	ln, err := s.ListenFunnel("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Printf("Listening on https://%v\n", s.CertDomains()[0])

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on http://localhost:%s\n", port)

	go func() {
		log.Fatal(http.Serve(ln, corsMiddleware(http.DefaultServeMux)))
	}()
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(http.DefaultServeMux)))
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

type WiFiConnection struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

type NetworkRoutingHint struct {
	Wifi WiFiConnection `json:"wifi,omitempty"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleDirectNetworkRouting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	response := NetworkRoutingHint{
		Wifi: WiFiConnection{
			SSID:     "Parkers.Wedding",
			Password: "parkers2025",
		},
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode network routing hint: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user := getUserFromRequest(r)
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	r.ParseMultipartForm(128 << 20) // 128MB limit

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
	isHEIC := strings.ToLower(ext) == ".heic" || strings.ToLower(ext) == ".heif" || mimeType == "image/heic" || mimeType == "image/heif"

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

func ensureCookieDomain(request *http.Request) *http.Response {
	cookie, err := request.Cookie("session")
	if err != nil {
		return nil
	}

	if cookie.Domain == "" {
		response := &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Header:     make(http.Header),
		}
		response.Header.Set("Location", "https://photos.parkers.wedding/login")
		response.Header.Set("Set-Cookie", "session=; Path=/; Domain=photos.parkers.wedding; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly")
		response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		response.Header.Set("Pragma", "no-cache")
		response.Header.Set("Expires", "0")
		return response
	}

	return nil
}

// Get user from session cookie
func getUserFromRequest(r *http.Request) *User {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}

	session, err := storage.GetSession(cookie.Value)
	if err != nil {
		return nil
	}

	user, err := storage.GetUser(session.UserID)
	if err != nil {
		return nil
	}

	return &user
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if cookieFix := ensureCookieDomain(r); cookieFix != nil {
		http.Error(w, "Session cookie domain is not set correctly", http.StatusTemporaryRedirect)
		cookieFix.Write(w)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromRequest(r)

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
		Domain:   "photos.parkers.wedding",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleDeletePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	user := getUserFromRequest(r)
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract photo ID from URL path
	path := r.URL.Path
	prefix := "/api/photos/"
	if !strings.HasPrefix(path, prefix) {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}
	photoID := strings.TrimPrefix(path, prefix)

	// Get photo metadata
	photo, err := storage.GetPhoto(photoID)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Check if user owns the photo
	if photo.UploadedBy != user.ID {
		http.Error(w, "You can only delete photos you uploaded", http.StatusForbidden)
		return
	}

	// Delete the physical file
	filename := photoID + filepath.Ext(photo.OriginalName)
	// Handle converted HEIC files
	if strings.HasSuffix(strings.ToLower(photo.OriginalName), ".heic") || strings.HasSuffix(strings.ToLower(photo.OriginalName), ".heif") {
		filename = photoID + ".jpg"
	}
	filePath := filepath.Join("uploads", filename)

	if err := os.Remove(filePath); err != nil {
		log.Printf("Failed to delete file %s: %v", filePath, err)
		// Continue with metadata deletion even if file deletion fails
	}

	// Delete photo metadata
	if err := storage.DeletePhoto(photoID); err != nil {
		log.Printf("Failed to delete photo metadata: %v", err)
		http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
		return
	}

	log.Printf("Photo deleted: %s by user %s", photoID, user.ID)
	w.WriteHeader(http.StatusNoContent)
}
