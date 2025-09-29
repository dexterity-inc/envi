package helpers

import (
	"fmt"
	"sync"
)

// MockKeyring provides a mock implementation of the keyring interface for testing
type MockKeyring struct {
	mu    sync.RWMutex
	store map[string]map[string]string // service -> user -> password
	
	// Error injection for testing
	SetError    error
	GetError    error
	DeleteError error
}

// NewMockKeyring creates a new mock keyring
func NewMockKeyring() *MockKeyring {
	return &MockKeyring{
		store: make(map[string]map[string]string),
	}
}

// Set stores a password in the mock keyring
func (m *MockKeyring) Set(service, user, password string) error {
	if m.SetError != nil {
		return m.SetError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.store[service] == nil {
		m.store[service] = make(map[string]string)
	}
	m.store[service][user] = password
	return nil
}

// Get retrieves a password from the mock keyring
func (m *MockKeyring) Get(service, user string) (string, error) {
	if m.GetError != nil {
		return "", m.GetError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if users, exists := m.store[service]; exists {
		if password, exists := users[user]; exists {
			return password, nil
		}
	}
	return "", fmt.Errorf("password not found in keyring")
}

// Delete removes a password from the mock keyring
func (m *MockKeyring) Delete(service, user string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if users, exists := m.store[service]; exists {
		delete(users, user)
		if len(users) == 0 {
			delete(m.store, service)
		}
	}
	return nil
}

// Clear removes all entries from the mock keyring
func (m *MockKeyring) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = make(map[string]map[string]string)
}

// SetSetError injects an error for Set operations
func (m *MockKeyring) SetSetError(err error) {
	m.SetError = err
}

// SetGetError injects an error for Get operations
func (m *MockKeyring) SetGetError(err error) {
	m.GetError = err
}

// SetDeleteError injects an error for Delete operations
func (m *MockKeyring) SetDeleteError(err error) {
	m.DeleteError = err
}

// HasEntry checks if an entry exists in the keyring
func (m *MockKeyring) HasEntry(service, user string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if users, exists := m.store[service]; exists {
		_, exists := users[user]
		return exists
	}
	return false
}

// MockFileSystem provides a mock file system for testing
type MockFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
	perms map[string]uint32
	mu    sync.RWMutex
	
	// Error injection
	readError  error
	writeError error
	statError  error
	mkdirError error
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
		perms: make(map[string]uint32),
	}
}

// WriteFile writes data to a mock file
func (m *MockFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	if m.writeError != nil {
		return m.writeError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[path] = data
	m.perms[path] = perm
	return nil
}

// ReadFile reads data from a mock file
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if m.readError != nil {
		return nil, m.readError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if data, exists := m.files[path]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// Stat returns information about a mock file
func (m *MockFileSystem) Stat(path string) (bool, uint32, error) {
	if m.statError != nil {
		return false, 0, m.statError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.files[path]; exists {
		return true, m.perms[path], nil
	}
	return false, 0, fmt.Errorf("file not found: %s", path)
}

// MkdirAll creates directories in the mock file system
func (m *MockFileSystem) MkdirAll(path string, perm uint32) error {
	if m.mkdirError != nil {
		return m.mkdirError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.dirs[path] = true
	return nil
}

// FileExists checks if a file exists in the mock file system
func (m *MockFileSystem) FileExists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.files[path]
	return exists
}

// DirExists checks if a directory exists in the mock file system
func (m *MockFileSystem) DirExists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.dirs[path]
}

// SetReadError injects an error for read operations
func (m *MockFileSystem) SetReadError(err error) {
	m.readError = err
}

// SetWriteError injects an error for write operations
func (m *MockFileSystem) SetWriteError(err error) {
	m.writeError = err
}

// SetStatError injects an error for stat operations
func (m *MockFileSystem) SetStatError(err error) {
	m.statError = err
}

// SetMkdirError injects an error for mkdir operations
func (m *MockFileSystem) SetMkdirError(err error) {
	m.mkdirError = err
}

// Clear removes all files and directories from the mock file system
func (m *MockFileSystem) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files = make(map[string][]byte)
	m.dirs = make(map[string]bool)
	m.perms = make(map[string]uint32)
}