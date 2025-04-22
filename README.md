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

# Or, tap first and then install
brew tap dexterity-inc/tap
brew install envi
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
- `envi version`: Display version information and build details
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

## Package Distribution

Envi is distributed through multiple package managers to make installation easy on any platform. For details on setting up your own distribution repositories or contributing to package maintenance, see [PACKAGING.md](./PACKAGING.md).

### Distributing with Homebrew

The project is already configured for Homebrew distribution using [GoReleaser](https://goreleaser.com/). To publish a new version:

1. Tag your release:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. Run GoReleaser:

```bash
# Set the GitHub token with repo scope
export GITHUB_TOKEN=your_github_token
export HOMEBREW_TAP_GITHUB_TOKEN=your_github_token

# Run GoReleaser
goreleaser release --clean
```

This will:

- Build binaries for all platforms
- Create GitHub release with assets
- Update the Homebrew tap formula

### Setting Up Windows Package Managers

#### Scoop Bucket Setup

1. Create a new GitHub repository named `scoop-bucket`

2. Add a manifest file named `envi.json`:

```json
{
  "version": "1.0.0",
  "description": "A secure tool for managing environment variables with GitHub Gists",
  "homepage": "https://github.com/dexterity-inc/envi",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/dexterity-inc/envi/releases/download/v1.0.0/envi-Windows-x86_64.zip",
      "hash": "<SHA256_HASH>",
      "extract_dir": "envi-Windows-x86_64"
    }
  },
  "bin": "envi.exe",
  "checkver": {
    "github": "https://github.com/dexterity-inc/envi"
  },
  "autoupdate": {
    "architecture": {
      "64bit": {
        "url": "https://github.com/dexterity-inc/envi/releases/download/v$version/envi-Windows-x86_64.zip"
      }
    }
  }
}
```

#### Chocolatey Package Setup

1. Create a new repository named `chocolatey-packages`

2. Add a package template:

```powershell
# envi.nuspec
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>envi</id>
    <version>1.0.0</version>
    <title>Envi</title>
    <authors>Dexterity Inc</authors>
    <projectUrl>https://github.com/dexterity-inc/envi</projectUrl>
    <licenseUrl>https://github.com/dexterity-inc/envi/blob/main/LICENSE</licenseUrl>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
    <description>A secure tool for managing environment variables with GitHub Gists</description>
    <tags>cli environment-variables github gist</tags>
  </metadata>
  <files>
    <file src="tools\**" target="tools" />
  </files>
</package>
```

3. Create `tools/chocolateyinstall.ps1`:

```powershell
$ErrorActionPreference = 'Stop'
$packageName = 'envi'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64 = 'https://github.com/dexterity-inc/envi/releases/download/v1.0.0/envi-Windows-x86_64.zip'
$checksum64 = '<SHA256_HASH>'

Install-ChocolateyZipPackage $packageName $url64 $toolsDir -checksum64 $checksum64 -checksumType64 'sha256'
```

## Development

### Building from Source

```bash
git clone https://github.com/dexterity-inc/envi.git
cd envi

# Using Make (recommended - includes version information)
make build

# Or using Go directly
go build -o envi ./cmd/envi
```

### Building with Version Information

The `version` command displays the current version, build date, and commit hash. This information is set during the build process:

```bash
# For GoReleaser (automatic during release)
goreleaser build --single-target --snapshot

# For manual builds (using Make)
make build

# Verify the version information
./bin/envi version
```

## License

MIT

## Credits

Created by Dexterity Inc.
