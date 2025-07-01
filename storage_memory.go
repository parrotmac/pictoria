package main

import (
	"sync"
)

// MemoryStorage implements Storage interface with in-memory storage
// This is useful for testing or temporary deployments
type MemoryStorage struct {
	mu       sync.RWMutex
	photos   []Photo
	users    map[string]User
	sessions map[string]Session
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		photos:   []Photo{},
		users:    make(map[string]User),
		sessions: make(map[string]Session),
	}
}

func (ms *MemoryStorage) Save() error {
	// No-op for memory storage
	return nil
}

func (ms *MemoryStorage) Load() error {
	// No-op for memory storage
	return nil
}

// Photo operations
func (ms *MemoryStorage) AddPhoto(photo Photo) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.photos = append(ms.photos, photo)
	return nil
}

func (ms *MemoryStorage) GetPhoto(id string) (Photo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, photo := range ms.photos {
		if photo.ID == id {
			return photo, nil
		}
	}
	return Photo{}, ErrNotFound
}

func (ms *MemoryStorage) GetAllPhotos() ([]Photo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]Photo, len(ms.photos))
	copy(result, ms.photos)
	return result, nil
}

func (ms *MemoryStorage) DeletePhoto(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	photos := []Photo{}
	found := false
	for _, photo := range ms.photos {
		if photo.ID != id {
			photos = append(photos, photo)
		} else {
			found = true
		}
	}

	if !found {
		return ErrNotFound
	}

	ms.photos = photos
	return nil
}

// User operations
func (ms *MemoryStorage) CreateUser(user User) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.users[user.ID]; exists {
		return ErrExists
	}

	ms.users[user.ID] = user
	return nil
}

func (ms *MemoryStorage) GetUser(id string) (User, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	user, exists := ms.users[id]
	if !exists {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (ms *MemoryStorage) GetAllUsers() ([]User, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	users := make([]User, 0, len(ms.users))
	for _, user := range ms.users {
		users = append(users, user)
	}
	return users, nil
}

func (ms *MemoryStorage) UpdateUser(user User) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.users[user.ID]; !exists {
		return ErrNotFound
	}

	ms.users[user.ID] = user
	return nil
}

func (ms *MemoryStorage) DeleteUser(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.users[id]; !exists {
		return ErrNotFound
	}

	delete(ms.users, id)
	return nil
}

// Session operations
func (ms *MemoryStorage) CreateSession(session Session) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.sessions[session.ID]; exists {
		return ErrExists
	}

	ms.sessions[session.ID] = session
	return nil
}

func (ms *MemoryStorage) GetSession(id string) (Session, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	session, exists := ms.sessions[id]
	if !exists {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (ms *MemoryStorage) DeleteSession(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.sessions[id]; !exists {
		return ErrNotFound
	}

	delete(ms.sessions, id)
	return nil
}

func (ms *MemoryStorage) GetAllSessions() ([]Session, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	sessions := make([]Session, 0, len(ms.sessions))
	for _, session := range ms.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}