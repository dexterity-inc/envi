package helpers

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-github/v37/github"
)

// TestConfig provides common test configuration
type TestConfig struct {
	TempDir    string
	ConfigDir  string
	TestEnvFile string
	CleanupFuncs []func()
}

// SetupTestEnvironment creates a temporary test environment
func SetupTestEnvironment(t *testing.T) *TestConfig {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "envi-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create config directory
	configDir := filepath.Join(tempDir, ".envi")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create test .env file
	testEnvFile := filepath.Join(tempDir, ".env")
	testEnvContent := `# Test environment variables
DB_HOST=localhost
DB_PORT=5432
API_KEY=test-api-key-12345
SECRET_TOKEN=super-secret-token
DEBUG=true
`
	if err := os.WriteFile(testEnvFile, []byte(testEnvContent), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	config := &TestConfig{
		TempDir:     tempDir,
		ConfigDir:   configDir,
		TestEnvFile: testEnvFile,
		CleanupFuncs: []func(){},
	}

	// Add cleanup function
	config.AddCleanup(func() {
		os.RemoveAll(tempDir)
	})

	return config
}

// AddCleanup adds a cleanup function to be called during test teardown
func (tc *TestConfig) AddCleanup(fn func()) {
	tc.CleanupFuncs = append(tc.CleanupFuncs, fn)
}

// Cleanup runs all cleanup functions
func (tc *TestConfig) Cleanup() {
	for _, fn := range tc.CleanupFuncs {
		fn()
	}
}

// CreateTestEnvFile creates a test .env file with specified content
func CreateTestEnvFile(t *testing.T, dir, content string) string {
	t.Helper()

	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	return envFile
}

// CreateTestConfigFile creates a test config file
func CreateTestConfigFile(t *testing.T, configDir string, content []byte) string {
	t.Helper()

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, content, 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	return configFile
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file to not exist: %s", path)
	}
}

// AssertFilePermissions checks file permissions
func AssertFilePermissions(t *testing.T, path string, expectedMode fs.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", path, err)
	}

	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected file permissions %o, got %o", expectedMode, info.Mode().Perm())
	}
}

// AssertFileContent checks if file content matches expected
func AssertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	if string(content) != expected {
		t.Errorf("File content mismatch.\nExpected:\n%s\nGot:\n%s", expected, string(content))
	}
}

// TestTimeout provides a test timeout context
func TestTimeout(t *testing.T, duration time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), duration)
}

// MockGitHubClient creates a mock GitHub client for testing
type MockGitHubClient struct {
	Gists   *MockGistsService
	Users   *MockUsersService
}

type MockGistsService struct {
	ListFunc   func(ctx context.Context, user string, opts *github.GistListOptions) ([]*github.Gist, *github.Response, error)
	GetFunc    func(ctx context.Context, id string) (*github.Gist, *github.Response, error)
	CreateFunc func(ctx context.Context, gist *github.Gist) (*github.Gist, *github.Response, error)
	EditFunc   func(ctx context.Context, id string, gist *github.Gist) (*github.Gist, *github.Response, error)
	DeleteFunc func(ctx context.Context, id string) (*github.Response, error)
}

type MockUsersService struct {
	GetFunc func(ctx context.Context, user string) (*github.User, *github.Response, error)
}

// List implements the GitHub Gists List method
func (m *MockGistsService) List(ctx context.Context, user string, opts *github.GistListOptions) ([]*github.Gist, *github.Response, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, user, opts)
	}
	return []*github.Gist{}, &github.Response{}, nil
}

// Get implements the GitHub Gists Get method
func (m *MockGistsService) Get(ctx context.Context, id string) (*github.Gist, *github.Response, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return &github.Gist{}, &github.Response{}, nil
}

// Create implements the GitHub Gists Create method
func (m *MockGistsService) Create(ctx context.Context, gist *github.Gist) (*github.Gist, *github.Response, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, gist)
	}
	return gist, &github.Response{}, nil
}

// Edit implements the GitHub Gists Edit method
func (m *MockGistsService) Edit(ctx context.Context, id string, gist *github.Gist) (*github.Gist, *github.Response, error) {
	if m.EditFunc != nil {
		return m.EditFunc(ctx, id, gist)
	}
	return gist, &github.Response{}, nil
}

// Delete implements the GitHub Gists Delete method
func (m *MockGistsService) Delete(ctx context.Context, id string) (*github.Response, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return &github.Response{}, nil
}

// DefaultTestTimeout is the default timeout for integration tests
const DefaultTestTimeout = 30 * time.Second