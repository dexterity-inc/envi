package utils

// File permissions
const (
	ConfigFilePerms = 0600
	KeyFilePerms    = 0600
	EnvFilePerms    = 0600
	DirPerms        = 0700
)

// Default file names and paths
const (
	DefaultConfigDir    = ".envi"
	DefaultConfigFile   = "config.yaml"
	DefaultKeyFile      = ".envi.key"
	DefaultEnvFile      = ".env"
	DefaultExampleFile  = ".env.example"
	DefaultBackupSuffix = ".backup"
	DefaultTempSuffix   = ".tmp"
)

// Encryption constants
const (
	EncryptionPrefix    = "ENVI_ENCRYPTED:"
	MaskedPrefix        = "ENVI_MASKED:"
	SelfContainedPrefix = "ENVI_SELF_CONTAINED:"
	KeyDerivationSalt   = "envi-sharing-salt-v1"
	EncryptionKeyLength = 32 // 256-bit key
)

// Application constants
const (
	ApplicationName = "envi-cli"
	TokenUsername   = "github-token"
	Version         = "1.0.0"
)

// UI constants
const (
	DefaultInputWidth = 25
	DefaultPageSize   = 10
	MaxPageSize       = 50
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
)

// GitHub API constants
const (
	GitHubAPIBaseURL = "https://api.github.com"
	GitHubGistURL    = "https://gist.github.com"
	MaxGistFileSize  = 100 * 1024 // 100KB
)

// Validation constants
const (
	MaxEnvVarLength = 1024
	MaxKeyLength    = 256
	MinPasswordLen  = 8
)

// Error messages
const (
	ErrConfigNotFound   = "configuration file not found"
	ErrInvalidToken     = "invalid GitHub token format"
	ErrFileNotFound     = "file not found"
	ErrPermissionDenied = "permission denied"
	ErrInvalidInput     = "invalid input"
	ErrEncryptionFailed = "encryption failed"
	ErrDecryptionFailed = "decryption failed"
	ErrNetworkError     = "network error"
	ErrGitHubAPIError   = "GitHub API error"
	ErrInvalidGistID    = "invalid Gist ID"
	ErrGistNotFound     = "Gist not found"
	ErrFileExists       = "file already exists"
	ErrInvalidPassword  = "invalid password"
	ErrEmptyContent     = "content cannot be empty"
	ErrInvalidFormat    = "invalid format"
)

// Success messages
const (
	MsgConfigSaved       = "Configuration saved successfully"
	MsgFileCreated       = "File created successfully"
	MsgFileUpdated       = "File updated successfully"
	MsgEncryptionSuccess = "Encryption completed successfully"
	MsgDecryptionSuccess = "Decryption completed successfully"
	MsgGistCreated       = "Gist created successfully"
	MsgGistUpdated       = "Gist updated successfully"
	MsgValidationPassed  = "Validation passed"
	MsgBackupCreated     = "Backup created successfully"
)

// Warning messages
const (
	WarnInsecurePerms   = "File has insecure permissions"
	WarnOverwriteFile   = "File will be overwritten"
	WarnTokenInConfig   = "Token stored in config file (consider using keyring)"
	WarnUnmaskedContent = "Content is encrypted but --unmask flag not specified"
	WarnDuplicateVars   = "Duplicate variables found"
	WarnMissingVars     = "Missing variables detected"
)

// Info messages
const (
	InfoUsingKeyring      = "Using keyring for token storage"
	InfoUsingConfigFile   = "Using config file for token storage"
	InfoUsingEnvVar       = "Using environment variable for token"
	InfoProcessingFile    = "Processing file"
	InfoConnectingGitHub  = "Connecting to GitHub"
	InfoValidatingContent = "Validating content"
)

// Help text
const (
	HelpTokenDescription   = "GitHub personal access token for API access"
	HelpEncryptDescription = "Encrypt data using AES-256 encryption"
	HelpMaskDescription    = "Mask values while keeping keys visible"
	HelpKeyFileDescription = "Path to encryption key file"
	HelpForceDescription   = "Force operation without confirmation"
	HelpVerboseDescription = "Enable verbose output"
	HelpDebugDescription   = "Enable debug mode"
)

// Filter options
const (
	FilterAll       = "all"
	FilterEncrypted = "encrypted"
	FilterPublic    = "public"
	FilterRecent    = "recent"
	FilterPrivate   = "private"
)

// Sort options
const (
	SortName  = "name"
	SortDate  = "date"
	SortSize  = "size"
	SortUsage = "usage"
)

// Environment types
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTesting     = "testing"
)

// Project types
const (
	ProjectWeb     = "web"
	ProjectAPI     = "api"
	ProjectMobile  = "mobile"
	ProjectDesktop = "desktop"
	ProjectCLI     = "cli"
	ProjectLibrary = "library"
)

// Time formats
const (
	TimeFormatISO   = "2006-01-02T15:04:05Z07:00"
	TimeFormatShort = "2006-01-02 15:04:05"
	TimeFormatDate  = "2006-01-02"
	TimeFormatHuman = "Jan 2, 2006 at 3:04 PM"
)

// HTTP status codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
)

// Retry constants
const (
	MaxRetries    = 3
	RetryDelay    = 1000 // milliseconds
	BackoffFactor = 2
	MaxBackoff    = 10000 // milliseconds
)

// Cache constants
const (
	CacheTTL     = 300 // seconds
	MaxCacheSize = 100 // items
)

// Security constants
const (
	MinEntropyBits = 128
	SaltLength     = 32
	NonceLength    = 12
	TagLength      = 16
)

// Validation patterns
const (
	PatternGistID     = `^[a-f0-9]{32}$`
	PatternGitHubUser = `^[a-zA-Z0-9-]+$`
	PatternEnvVar     = `^[A-Z_][A-Z0-9_]*$`
	PatternFileName   = `^[a-zA-Z0-9._-]+$`
)

// Performance constants
const (
	DefaultBufferSize    = 8192 // 8KB buffer for I/O operations
	MaxConcurrentWorkers = 10   // Maximum concurrent workers for parallel operations
	MemoryPoolSize       = 100  // Size of object pools for memory reuse
	BatchSize            = 50   // Default batch size for bulk operations
)
