package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
)

type Storage interface {
	// Photo operations
	AddPhoto(photo Photo) error
	GetPhoto(id string) (Photo, error)
	GetAllPhotos() ([]Photo, error)
	DeletePhoto(id string) error

	// User operations
	CreateUser(ctx context.Context, user User) error
	GetUser(id string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	GetAllUsers() ([]User, error)
	UpdateUser(user User) error
	DeleteUser(id string) error

	// Session operations
	CreateSession(session Session) error
	GetSession(id string) (Session, error)
	DeleteSession(id string) error
	GetAllSessions() ([]Session, error)

	// Persistence
	Save() error
	Load() error
}

type FileStorage struct {
	mu       sync.RWMutex
	filename string
	data     *StorageData
}

type StorageData struct {
	Photos   []Photo            `json:"photos"`
	Users    map[string]User    `json:"users"`
	Sessions map[string]Session `json:"sessions"`
	Version  int                `json:"version"`
}

// func NewFileStorage(filename string) (Storage, error) {
// 	fs := &FileStorage{
// 		filename: filename,
// 		data: &StorageData{
// 			Photos:   []Photo{},
// 			Users:    make(map[string]User),
// 			Sessions: make(map[string]Session),
// 			Version:  1,
// 		},
// 	}

// 	if err := fs.Load(); err != nil {
// 		if !os.IsNotExist(err) {
// 			return nil, err
// 		}
// 		// File doesn't exist, save initial empty state
// 		if err := fs.saveInternal(); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return fs, nil
// }

func (fs *FileStorage) Load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(fs.filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &fs.data)
}

func (fs *FileStorage) Save() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.saveInternal()
}

// saveInternal saves without acquiring locks - must be called with lock held
func (fs *FileStorage) saveInternal() error {
	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.filename, data, 0644)
}

// Photo operations
func (fs *FileStorage) AddPhoto(photo Photo) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.Photos = append(fs.data.Photos, photo)
	return fs.saveInternal()
}

func (fs *FileStorage) GetPhoto(id string) (Photo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for _, photo := range fs.data.Photos {
		if photo.ID == id {
			return photo, nil
		}
	}
	return Photo{}, ErrNotFound
}

func (fs *FileStorage) GetAllPhotos() ([]Photo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]Photo, len(fs.data.Photos))
	copy(result, fs.data.Photos)
	return result, nil
}

func (fs *FileStorage) DeletePhoto(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	photos := []Photo{}
	found := false
	for _, photo := range fs.data.Photos {
		if photo.ID != id {
			photos = append(photos, photo)
		} else {
			found = true
		}
	}

	if !found {
		return ErrNotFound
	}

	fs.data.Photos = photos
	return fs.saveInternal()
}

// User operations
func (fs *FileStorage) CreateUser(user User) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.data.Users[user.ID]; exists {
		return ErrExists
	}

	fs.data.Users[user.ID] = user
	return fs.saveInternal()
}

func (fs *FileStorage) GetUser(id string) (User, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	user, exists := fs.data.Users[id]
	if !exists {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (fs *FileStorage) GetAllUsers() ([]User, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	users := make([]User, 0, len(fs.data.Users))
	for _, user := range fs.data.Users {
		users = append(users, user)
	}
	return users, nil
}

func (fs *FileStorage) UpdateUser(user User) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.data.Users[user.ID]; !exists {
		return ErrNotFound
	}

	fs.data.Users[user.ID] = user
	return fs.saveInternal()
}

func (fs *FileStorage) DeleteUser(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.data.Users[id]; !exists {
		return ErrNotFound
	}

	delete(fs.data.Users, id)
	return fs.saveInternal()
}

// Session operations
func (fs *FileStorage) CreateSession(session Session) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.data.Sessions[session.ID]; exists {
		return ErrExists
	}

	fs.data.Sessions[session.ID] = session
	return fs.saveInternal()
}

func (fs *FileStorage) GetSession(id string) (Session, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	session, exists := fs.data.Sessions[id]
	if !exists {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (fs *FileStorage) DeleteSession(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.data.Sessions[id]; !exists {
		return ErrNotFound
	}

	delete(fs.data.Sessions, id)
	return fs.saveInternal()
}

func (fs *FileStorage) GetAllSessions() ([]Session, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	sessions := make([]Session, 0, len(fs.data.Sessions))
	for _, session := range fs.data.Sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}
