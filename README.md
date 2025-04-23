# Envi

Envi is a CLI tool for securely managing and sharing environment variables using GitHub Gists.

## Features

- ✅ **Store Securely**: Store your `.env` files in private GitHub Gists
- 🔒 **Encrypted Storage**: Mask or fully encrypt sensitive values
- 🔄 **Easy Sharing**: Share environment variables with team members
- 🚀 **Simple Commands**: Push, pull, and manage your environment variables with ease

## Installation

### Via Homebrew (macOS and Linux)

```bash
# Install from the official tap
brew install dexterity-inc/tap/envi
```

### Windows Package Managers

#### Scoop

```bash
# Add the bucket (only needed once)
scoop bucket add dexterity-inc https://github.com/dexterity-inc/scoop-bucket.git

# Install envi
scoop install envi
```

#### Chocolatey

```bash
choco install envi
```

### Manual Installation

#### macOS / Linux

```bash
# Download the binary for your platform
curl -sSL https://github.com/dexterity-inc/envi/releases/latest/download/envi-$(uname -s)-$(uname -m) -o /usr/local/bin/envi
chmod +x /usr/local/bin/envi
```

#### Windows

Download the latest binary from the [releases page](https://github.com/dexterity-inc/envi/releases).

## Quick Start

1. Configure your GitHub token (needs Gist scope):

```bash
envi config --token YOUR_GITHUB_TOKEN
```

2. Push your .env file to a private Gist:

```bash
envi push
```

3. Pull your .env file on another machine:

```bash
envi pull --id GIST_ID
```

## Security Features

- **Token Storage**: Your GitHub token is securely stored in your system's credential manager
- **Masked Encryption**: Keep variable names visible but encrypt values (default)
- **Full Encryption**: Encrypt the entire .env file
- **Key-based Encryption**: Use a key file for enhanced security

## Core Commands

- `envi config`: Configure settings and GitHub token
- `envi push`: Push .env file to GitHub Gist
- `envi pull`: Pull .env file from GitHub Gist
- `envi list`: List your GitHub Gists with .env files
- `envi share`: Share .env files with team members
- `envi validate`: Validate .env file format and required variables
- `envi merge`: Merge multiple .env files with conflict resolution
- `envi completion`: Generate shell completion scripts for better CLI experience

## Advanced Usage

### Encryption Options

```bash
# Push with masked encryption (default)
envi push

# Push with full encryption
envi push --encrypt

# Pull encrypted values and decrypt them
envi pull --unmask

# Use a key file instead of password
envi push --encrypt --use-key-file --key-file ~/.envi.key
```

### Sharing

```bash
# Share with specific GitHub users
envi share --users user1,user2

# Generate a shareable URL
envi share --url
```

### Validating .env Files

```bash
# Basic validation of .env format
envi validate

# Strict validation (no empty values)
envi validate --strict

# Check for required variables
envi validate --required DB_HOST,API_KEY,SECRET_TOKEN

# Validate a specific file
envi validate --file .env.production
```

### Merging .env Files

```bash
# Merge multiple .env files
envi merge --files .env.base,.env.local,.env.dev

# Merge with overwrite (last file wins)
envi merge --files .env.defaults,.env.custom --overwrite

# Output to a different file and sort variables
envi merge --files .env.base,.env.test --output .env.combined --sort
```

### Version Information

To check the version of Envi CLI:

```bash
# Display version number
envi --version

# Or using the short flag
envi -v
```

### Shell Completion

Enable tab-completion for Envi commands in your terminal:

```bash
# Bash
source <(envi completion bash)

# Zsh
source <(envi completion zsh)

# Fish
envi completion fish | source

# PowerShell
envi completion powershell | Out-String | Invoke-Expression
```

For permanent installation, see the help with `envi completion --help`.

## License

MIT

## Credits

Created by Dexterity Inc.
