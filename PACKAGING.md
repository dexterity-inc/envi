# Package Distribution Guide

This document outlines the steps to distribute Envi through various package managers.

## Prerequisites

- A GitHub account with access to create repositories
- Access to the Envi repository
- GitHub Personal Access Token with `repo` scope
- For Chocolatey: An account on chocolatey.org

## Setup Steps

### 1. Homebrew Tap Setup

The Homebrew tap has already been created at `dexterity-inc/homebrew-tap` with a formula for Envi:

```ruby
class Envi < Formula
  desc "CLI tool to push and pull .env files to/from GitHub Gists"
  homepage "https://github.com/dexterity-inc/envi"
  url "https://github.com/dexterity-inc/envi/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "f775b026898dcba5c18d332f838785f0084c15c3ac3aedebdada4e1d83e158d0"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args
  end

  test do
    assert_match "environment variable management", shell_output("#{bin}/envi --help")
  end
end
```

When a new version is released, GoReleaser will automatically update this formula with the new version and SHA hash.

### 2. Scoop Bucket Setup

1. **Create the Scoop bucket repository**:

   ```bash
   # Create a new public repository named 'scoop-bucket'
   ```

2. **Initialize the repository**:

   ```bash
   echo "# Scoop Bucket for Dexterity Inc tools" > README.md
   git add README.md
   git commit -m "Initial commit"
   git push
   ```

3. **Create the bucket JSON file**:
   ```bash
   # See packaging/scoop/envi.json for the template
   ```

### 3. Chocolatey Package Setup

1. **Create the Chocolatey packages repository**:

   ```bash
   # Create a new repository named 'chocolatey-packages'
   ```

2. **Set up the package structure**:

   ```bash
   mkdir -p envi/tools
   # Copy packaging/chocolatey/envi.nuspec to envi/
   # Copy packaging/chocolatey/tools/* to envi/tools/
   ```

3. **Register on Chocolatey.org**:
   - Create an account at https://chocolatey.org
   - Generate an API key

## Automated Release Process

The project is configured to use GoReleaser for automated releases to all package managers.

### Setting Up Environment Variables

Add these secrets to your GitHub repository:

- `GITHUB_TOKEN`: GitHub Personal Access Token with repo scope
- `HOMEBREW_TAP_GITHUB_TOKEN`: GitHub token for Homebrew tap
- `SCOOP_BUCKET_GITHUB_TOKEN`: GitHub token for Scoop bucket
- `CHOCOLATEY_API_KEY`: API key from chocolatey.org

### Creating a Release

1. **Tag a new version**:

   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **Automated Release with GitHub Actions**:

   Simply push the tag, and the GitHub Actions workflow will automatically:

   - Validate the GoReleaser configuration
   - Build binaries for all platforms
   - Create a GitHub release with assets
   - Update the Homebrew formula
   - Update the Scoop manifest
   - Create and publish a Chocolatey package

3. **Manual Release with GoReleaser**:

   If you prefer to run the release process manually:

   ```bash
   # Set environment variables
   export GITHUB_TOKEN=your_github_token
   export HOMEBREW_TAP_GITHUB_TOKEN=your_github_token
   export SCOOP_BUCKET_GITHUB_TOKEN=your_github_token
   export CHOCOLATEY_API_KEY=your_chocolatey_api_key

   # Run GoReleaser
   goreleaser release --clean
   ```

## Manual Package Updates

If you need to manually update packages:

### Homebrew

```bash
# Update the formula in your tap
cd homebrew-tap
vi Formula/envi.rb

# Update the version and SHA hash
# You can get the SHA hash with:
# curl -sL https://github.com/dexterity-inc/envi/archive/refs/tags/v1.0.0.tar.gz | shasum -a 256

# Commit and push the changes
git commit -am "Update envi to v1.0.0"
git push
```

### Scoop

```bash
# Update the JSON manifest with the correct version and hash
# See packaging/scoop/envi.json for reference
```

### Chocolatey

```bash
# Pack the package
cd envi
choco pack

# Push to Chocolatey.org
choco push envi.1.0.0.nupkg --api-key=your_api_key
```

## Testing Packages

### Homebrew

```bash
# Install from tap
brew install dexterity-inc/tap/envi

# Verify installation
envi --version
```

### Scoop

```bash
# Add the bucket
scoop bucket add dexterity-inc https://github.com/dexterity-inc/scoop-bucket

# Install the package
scoop install envi

# Verify installation
envi --version
```

### Chocolatey

```bash
# Install locally first to test
choco install envi -source .

# Install from Chocolatey.org
choco install envi

# Verify installation
envi --version
```
