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

### File Handling

- Proper file permissions (0600) ensure only the file owner can read or write the configuration
- Automatic backup creation for critical operations (merge, pull) allows recovery if needed

### Gist Security

Remember that private Gists:

- Are not visible in public listings
- Are not truly private - anyone with the URL can access them
- Should not be used for highly sensitive information

For highly sensitive environments, consider using a dedicated secrets management service.

## Command Reference

### Push Command

Push your local `.env` file to a GitHub Gist.

```bash
envi push
```

**Flags**:

- `--id`, `-i` [string]: GitHub Gist ID to update (if not provided, a new Gist will be created)
- `--save`, `-s` [boolean]: Save the Gist ID to config for future updates (default: true)

**Examples**:

```bash
# Create a new Gist
envi push

# Update an existing Gist
envi push --id abc123def456

# Create a new Gist without saving the ID
envi push --save=false
```

### Pull Command

Pull an `.env` file from a GitHub Gist and save it to the current directory.

```bash
envi pull
```

**Flags**:

- `--id`, `-i` [string]: GitHub Gist ID to pull from (if not specified, uses the saved ID)
- `--backup`, `-b` [boolean]: Create a backup of the existing `.env` file before overwriting (default: false)
- `--force`, `-f` [boolean]: Force overwrite without confirmation if `.env` file exists (default: false)

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

# Pull with all options
envi pull --id abc123def456 --backup --force
```

### List Command

List available Gists containing `.env` files from your GitHub account.

```bash
envi list
```

**Output**:

- Displays a list of your Gists containing `.env` files
- Currently selected Gist (last used) is highlighted with an asterisk (\*)
- Shows Gist descriptions and update times

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
```

**Output**:

- GitHub token status (securely stored or partially masked)
- Default Gist ID (if set)
- Usage examples for the saved Gist ID
- Security information about token storage

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
