# Envi CLI

A simple CLI tool to push and pull `.env` files to/from GitHub Gists for secure storage and sharing.

## Installation

### macOS and Linux

#### Using Homebrew (Recommended)

```bash
brew tap dexterity-inc/tap
brew install envi
```

#### Manual Installation

1. Download the latest binary from the [Releases page](https://github.com/dexterity-inc/envi/releases)
2. Make it executable: `chmod +x envi`
3. Move it to your PATH: `sudo mv envi /usr/local/bin/`

### Windows (Coming Soon)

Windows installation options are in progress:

- Scoop package
- Chocolatey package
- MSI installer

For now, Windows users can:

1. Download the latest Windows binary from the [Releases page](https://github.com/dexterity-inc/envi/releases)
2. Add the executable to your PATH

### Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/dexterity-inc/envi.git
cd envi

# Build the binary
go build -o envi

# Optional: Move to PATH
sudo mv envi /usr/local/bin/
```

## Usage

### Set up GitHub Personal Access Token

You have two options to set your GitHub token:

1. Using the config command (recommended):

```bash
envi config --token your_github_personal_access_token
```

This will save your token in your system's secure credential store.

2. Using an environment variable (temporary):

```bash
export GITHUB_TOKEN=your_github_personal_access_token
```

In both cases, you'll need a GitHub token with the `gist` scope.

## Security Features

Envi includes several security features to protect your sensitive data:

### Secure Token Storage

Your GitHub token is stored in your system's secure credential manager:

- macOS: Keychain
- Linux: Secret Service API/libsecret
- Windows: Windows Credential Manager

This provides much better security than storing tokens in plaintext files.

### Token Protection

- Your token is never displayed in full in terminal output
- If secure credential storage isn't available, a fallback mechanism stores tokens with proper file permissions (0600)
- Token removal command (`envi config --clear-token`) ensures tokens can be completely removed from storage
- Warning when providing token via command line to discourage this less secure method

### File Handling

- Proper file permissions (0600) ensure only the file owner can read or write the configuration
- Automatic backup creation for critical operations (merge, pull) allows recovery if needed
- Automatic fixing of insecure permissions on key files

### Encryption

Envi provides strong encryption for your .env files before storing them in GitHub Gists:

- Industry-standard AES-256-GCM encryption
- Two authentication options:
  - Password-based encryption (uses SHA-256 for key derivation) with confirmation prompt
  - File-based key storage (more secure for automated workflows)
- Two encryption modes:
  - Full encryption: Encrypts the entire file content
  - Masked encryption: Encrypts only the values while keeping variable names visible
- Encrypted data is base64-encoded for safe storage
- Automatic encryption/decryption with minimal configuration
- Strict validation of key file length and format
- Automatic permission fixing for key files

#### Encryption Usage

To encrypt your .env file when pushing:

```bash
# Encrypt with password (will prompt for password and confirmation)
envi push --encrypt

# Mask values only, keeping variable names visible
envi push --mask

# Encrypt with a key file
envi push --encrypt --use-key-file --key-file ~/.envi.key

# Generate a new random key file and use it
envi push --encrypt --generate-key --key-file ~/.envi.key

# Provide password via command line (not recommended)
envi push --encrypt --password "mypassword"
```

To decrypt when pulling:

```bash
# Auto-detect encryption and decrypt (will prompt for password if needed)
envi pull

# Explicitly decrypt with password
envi pull --decrypt

# Explicitly unmask values
envi pull --unmask

# Decrypt with key file
envi pull --decrypt --use-key-file --key-file ~/.envi.key
```

#### Setting Encryption Defaults

Configure encryption defaults to streamline your workflow:

```bash
# Always encrypt when pushing
envi config --encrypt-by-default

# Always use masked encryption (only encrypt values)
envi config --encrypt-by-default --mask-by-default

# Set default key file
envi config --default-key-file ~/.envi.key

# Use key file by default instead of password
envi config --use-key-file

# Disable encryption by default
envi config --disable-encryption
```

### Gist Security

Remember that private Gists:

- Are not visible in public listings
- Are not truly private - anyone with the URL can access them
- Should not be used for highly sensitive information without encryption

For highly sensitive environments, always use encryption with envi or consider a dedicated secrets management service.

## Command Reference

### Push Command

Push your local `.env` file to a GitHub Gist.

```bash
envi push
```

**Flags**:

- `--id`, `-i` [string]: GitHub Gist ID to update (if not provided, a new Gist will be created)
- `--save`, `-s` [boolean]: Save the Gist ID to config for future updates (default: true)
- `--desc`, `-d` [string]: Custom description for the Gist (default: auto-generated description with project and date)
- `--encrypt`, `-e` [boolean]: Encrypt .env file before pushing to Gist (default: false, unless configured)
- `--mask`, `-m` [boolean]: Mask the values in the .env file before pushing (default: false, unless configured)
- `--use-key-file` [boolean]: Use a key file for encryption instead of password (default: false, unless configured)
- `--key-file`, `-k` [string]: Path to the encryption key file (default: .envi.key, unless configured)
- `--password`, `-p` [string]: Password for encryption (not recommended, use key file instead)
- `--generate-key` [boolean]: Auto-generate a strong encryption key and save to key file (default: false)

**Examples**:

```bash
# Create a new Gist
envi push

# Update an existing Gist
envi push --id abc123def456

# Create a new Gist with custom description
envi push --desc "Production environment variables"

# Create a new Gist without saving the ID
envi push --save=false

# Push with encryption (password-based)
envi push --encrypt

# Push with masked encryption (only values encrypted)
envi push --mask

# Push with encryption using a key file
envi push --encrypt --use-key-file --key-file ~/.envi.key

# Push with encryption and generate a new key file
envi push --encrypt --generate-key
```

### Pull Command

Pull an `.env` file from a GitHub Gist and save it to the current directory.

```bash
envi pull
```

**Flags**:

- `--id`, `-i` [string]: GitHub Gist ID to pull from (if not specified, uses the saved ID)
- `--backup`, `-b` [boolean]: Create a backup of the existing `.env` file before overwriting (default: true)
- `--force`, `-f` [boolean]: Force overwrite without confirmation if `.env` file exists (default: false)
- `--decrypt`, `-d` [boolean]: Decrypt .env file after pulling from Gist (default: false, unless configured)
- `--unmask` [boolean]: Unmask values in .env file (for masked encryption) (default: false, unless configured)
- `--use-key-file` [boolean]: Use a key file for decryption instead of password (default: false, unless configured)
- `--key-file`, `-k` [string]: Path to the encryption key file (default: .envi.key, unless configured)
- `--password`, `-p` [string]: Password for decryption (not recommended, use key file instead)

**Examples**:

```bash
# Pull using saved Gist ID (with confirmation)
envi pull

# Pull from specific Gist ID
envi pull --id abc123def456

# Pull and create backup
envi pull --backup

# Pull with force overwrite (no confirmation)
envi pull --force

# Pull and decrypt (password-based)
envi pull --decrypt

# Pull and unmask values
envi pull --unmask

# Pull and decrypt with key file
envi pull --decrypt --use-key-file --key-file ~/.envi.key

# Pull with all options
envi pull --id abc123def456 --backup --force --decrypt
```

### List Command

List available Gists containing `.env` files from your GitHub account.

```bash
envi list
```

**Flags**:

- `--full-id` [boolean]: Display full Gist IDs instead of shortened IDs (default: false)
- `--format` [string]: Custom format for displaying gists (e.g. 'id,desc,date')
- `--sort` [string]: Sort gists by: 'date', 'name', or 'id' (default: 'date')

**Examples**:

```bash
# List all .env gists
envi list

# Show full Gist IDs
envi list --full-id

# Sort by project name
envi list --sort=name

# Sort by Gist ID
envi list --sort=id

# Custom format showing only IDs and descriptions
envi list --format="id,desc"

# Custom format showing only IDs and update dates
envi list --format="id,date"
```

**Output**:

- Displays a color-coded list of your Gists containing `.env` files
- Currently selected Gist (last used) is highlighted with a green asterisk (\*)
- Shows project names, dates, and Gist descriptions
- Indicates encryption status with [encrypted] or [masked] labels
- Shows shortened Gist IDs by default (first 7 characters)
- Provides human-readable time formats (e.g., "2 hours ago", "yesterday")

### Diff Command

Compare your local `.env` file with a remote one stored in a GitHub Gist.

```bash
envi diff
```

**Flags**:

- `--id`, `-i` [string]: GitHub Gist ID to compare with (if not specified, uses the saved ID)
- `--values`, `-v` [boolean]: Show values in the diff output (may expose sensitive data) (default: false)

**Examples**:

```bash
# Compare with saved Gist ID
envi diff

# Compare with specific Gist ID
envi diff --id abc123def456

# Compare and show actual values
envi diff --values
```

**Output**:

- Variables only in local file
- Variables only in remote file
- Variables with different values
- Summary of differences

### Merge Command

Merge your local `.env` file with a remote one stored in a GitHub Gist.

```bash
envi merge
```

**Flags**:

- `--id`, `-i` [string]: GitHub Gist ID to merge with (if not specified, uses the saved ID)
- `--local`, `-l` [boolean]: Prefer local values when there are conflicts (default: false)
- `--remote`, `-r` [boolean]: Prefer remote values when there are conflicts (default: false)
- `--backup`, `-b` [boolean]: Create a backup of the local `.env` file before merging (default: true)
- `--push`, `-p` [boolean]: Push the merged result back to the Gist (default: false)

**Examples**:

```bash
# Interactive merge with saved Gist ID
envi merge

# Merge with specific Gist ID
envi merge --id abc123def456

# Merge preferring local values for conflicts
envi merge --local

# Merge preferring remote values for conflicts
envi merge --remote

# Merge without creating a backup
envi merge --backup=false

# Merge and push result back to Gist
envi merge --push

# Combine options
envi merge --id abc123def456 --local --backup --push
```

**Note**: You cannot use both `--local` and `--remote` flags together.

### Validate Command

Compare your `.env` file with `.env.example` to find missing variables.

```bash
envi validate
```

**Flags**:

- `--fix`, `-f` [boolean]: Automatically add missing environment variables from `.env.example` to `.env` (default: false)
- `--example`, `-e` [string]: Path to the example env file (default: ".env.example")

**Examples**:

```bash
# Check for missing variables
envi validate

# Auto-fix missing variables
envi validate --fix

# Use custom example file
envi validate --example=.env.template

# Use custom example file and auto-fix
envi validate --example=.env.template --fix
```

**Output**:

- List of missing variables
- Success message if all variables exist
- Automatic addition of missing variables when using `--fix`

### Config Command

Configure Envi CLI settings including your GitHub token and default Gist ID.

```bash
envi config
```

**Flags**:

- `--token`, `-t` [string]: Set your GitHub personal access token
- `--clear-gist`, `-c` [boolean]: Clear the saved Gist ID (default: false)
- `--clear-token` [boolean]: Remove the GitHub token from secure storage (default: false)
- `--encrypt-by-default` [boolean]: Enable encryption by default for all push operations (default: false)
- `--mask-by-default` [boolean]: Use masked encryption by default (only encrypt values) (default: false)
- `--disable-encryption` [boolean]: Disable encryption by default (default: false)
- `--default-key-file` [string]: Set the default encryption key file path
- `--use-key-file` [boolean]: Use key file by default instead of password for encryption (default: false)
- `--force-file-storage` [boolean]: Force token storage in file instead of system credential manager (not recommended)

**Examples**:

```bash
# View current configuration
envi config

# Set GitHub token
envi config --token=your_github_personal_access_token

# Clear saved Gist ID
envi config --clear-gist

# Remove stored GitHub token
envi config --clear-token

# Enable encryption by default
envi config --encrypt-by-default

# Enable masked encryption by default
envi config --encrypt-by-default --mask-by-default

# Set default key file
envi config --default-key-file ~/.envi.key

# Use key file by default
envi config --use-key-file

# Disable default encryption
envi config --disable-encryption
```

**Output**:

- GitHub token status (securely stored or partially masked)
- Default Gist ID (if set)
- Encryption settings
- Usage examples for the saved Gist ID
- Security information about token storage

## Best Practices

### Secure Usage

1. **Always use encryption** for sensitive environment variables, especially production credentials
2. **Use key files** instead of passwords for automated workflows
3. **Store key files securely** and ensure they have proper permissions (0600)
4. **Back up your encryption keys** in a secure location
5. **Never commit key files** to version control
6. **Avoid using the `--password` flag** on the command line, as it may be visible in command history

### Gist Management

1. **Use descriptive Gist descriptions** to easily identify different environments
2. **Use the `--sort` option** when listing Gists to organize them by project, date, or ID
3. **Use the `--format` option** to customize the Gist list display for your needs
4. **Always check the Gist ID** when pulling or pushing to avoid accidentally using the wrong environment

## Command Help

For more information about any command:

```bash
# General help
envi --help

# Command-specific help
envi push --help
envi pull --help
envi list --help
envi diff --help
envi merge --help
envi validate --help
envi config --help
```

## Security Notes

- Envi uses AES-256-GCM encryption, which is a secure authenticated encryption mode
- Password-based encryption uses SHA-256 for key derivation
- File permissions are automatically checked and fixed for key files
- Secure validation is performed for key files and passwords
- Always keep your encryption keys secure - if you lose them, you won't be able to decrypt your data
- Consider using masked encryption when you need variable names to remain visible but values encrypted
