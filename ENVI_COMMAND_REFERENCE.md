# ENVI Command Reference

A comprehensive guide to all available commands and flags in the ENVI tool.

## Global Flags

These flags can be used with any command:

| Flag                    | Description                                       |
| ----------------------- | ------------------------------------------------- |
| `--encrypt`             | Encrypt data using AES-256                        |
| `-k, --key-file string` | Path to encryption key file (default ".envi.key") |
| `-m, --mask`            | Mask values (keep keys visible)                   |
| `--tui`                 | Use interactive terminal UI (default true)        |
| `--use-key-file`        | Use key file instead of password                  |

## Commands

### config

Configure CLI settings including GitHub token and default Gist ID.

**Usage**: `envi config [flags]`

**Flags**:

| Flag                        | Description                                                                        |
| --------------------------- | ---------------------------------------------------------------------------------- |
| `-t, --token string`        | Set your GitHub personal access token                                              |
| `-c, --clear-gist`          | Clear the saved Gist ID                                                            |
| `--clear-token`             | Remove the GitHub token from secure storage                                        |
| `--encrypt-by-default`      | Enable full encryption by default                                                  |
| `--disable-encryption`      | Disable encryption by default                                                      |
| `--unmask-by-default`       | Automatically unmask/decrypt values when pulling                                   |
| `--default-key-file string` | Set the default encryption key file path                                           |
| `--use-key-file`            | Use key file by default instead of password                                        |
| `--force-file-storage`      | Force token storage in file instead of system credential manager (not recommended) |

**Examples**:

```bash
# Set GitHub token
envi config -t YOUR_GITHUB_TOKEN

# Enable full encryption by default
envi config --encrypt-by-default

# Clear stored GitHub token
envi config --clear-token
```

**Output Example**:

```
GitHub Token: Securely stored in system credential manager
Default Gist ID: Not set

Encryption Settings:
  ✓ Masked encryption enabled by default (variable names visible, values encrypted)
    This is the default behavior
  ✓ Values will be automatically unmasked when pulling
  • Using password-based encryption

Security Information:
  ✓ Persistent token is stored in your system's secure credential manager
  ✓ To remove the token: envi config --clear-token
```

### validate

Compare your project's .env file with .env.example to identify missing variables.

**Usage**: `envi validate [flags]`

**Flags**:

| Flag                 | Description                                       |
| -------------------- | ------------------------------------------------- |
| `--fix`              | Fix missing variables by adding them to .env file |
| `-s, --strict`       | Use strict validation (no empty values)           |
| `--required strings` | Required variables (comma-separated)              |

**Examples**:

```bash
# Basic validation
envi validate

# Validate with strict mode (no empty values)
envi validate --strict

# Validate with required variables
envi validate --required API_KEY,DB_USER

# Fix missing variables
envi validate --fix
```

**Output Example**:

```
❌ Found 1 missing variables in .env:
  SECRET_KEY4=
Run 'envi validate --fix' to add these missing variables to your .env file
⚠️  Found 1 extra variables in .env that are not in .env.example:
  EXTRA_VAR=test
You may want to add these to .env.example if they are needed
```

### list

List all your GitHub Gists containing .env files.

**Usage**: `envi list [flags]`

**Flags**:

| Flag                  | Description                                    |
| --------------------- | ---------------------------------------------- |
| `-a, --all`           | Show all Gists, not just those with .env files |
| `-f, --format string` | Output format (table, json) (default "table")  |
| `-l, --limit int`     | Limit number of Gists to show (default 10)     |
| `-u, --urls`          | Show Gist URLs in output                       |

**Examples**:

```bash
# List .env Gists
envi list

# List with URLs
envi list -u

# Show only 3 Gists
envi list -l 3

# Output as JSON
envi list -f json
```

**Output Example (Table Format)**:

```
ID                                DESCRIPTION                               FILES  CREATED
47860ee110bc477ab759e91202490270  Environment variables for envi (2025-...  .env   2025-04-19
84bc1c89fe5025173a34aab0247a0f12  Environment variables for envi (2025-...  .env   2025-04-19
ef05485ea9c6f24dcf9d279f68ff1f7d  Environment variables for envi (2025-...  .env   2025-04-18

* = current Gist
```

### push

Push your .env file to a new or existing GitHub Gist with optional encryption.

**Usage**: `envi push [flags]`

**Flags**:

| Flag                       | Description                                                                  |
| -------------------------- | ---------------------------------------------------------------------------- |
| `-f, --file string`        | Path to the .env file (default ".env")                                       |
| `-i, --id string`          | GitHub Gist ID to update (leave blank for new Gist)                          |
| `-p, --public`             | Make the Gist public (default private)                                       |
| `-d, --description string` | Description for the Gist (default "Environment variables created with envi") |
| `-a, --auto`               | Auto-generate a sample .env file if none exists                              |

**Examples**:

```bash
# Push current .env file to a new Gist
envi push

# Push a specific .env file
envi push -f .env.production

# Update an existing Gist
envi push -i YOUR_GIST_ID

# Push as a public Gist
envi push -p
```

### pull

Pull your .env file from a GitHub Gist with optional decryption.

**Usage**: `envi pull [flags]`

**Flags**:

| Flag                    | Description                                       |
| ----------------------- | ------------------------------------------------- |
| `-f, --force`           | Overwrite existing file without confirmation      |
| `-i, --id string`       | GitHub Gist ID to pull from                       |
| `-o, --output string`   | Output file path (default ".env")                 |
| `-k, --key-file string` | Path to encryption key file (default ".envi.key") |
| `-p, --password string` | Encryption password (not recommended)             |
| `-u, --unmask`          | Decrypt/unmask values when pulling                |
| `--use-key-file`        | Use key file instead of password                  |

**Examples**:

```bash
# Pull from default Gist
envi pull

# Pull from specific Gist
envi pull -i YOUR_GIST_ID

# Force overwrite existing file
envi pull -f

# Pull and decrypt values
envi pull -u
```

### share

Share your .env file with team members by creating a shared Gist or generating a shareable URL.

**Usage**: `envi share [flags]`

**Flags**:

| Flag                  | Description                                       |
| --------------------- | ------------------------------------------------- |
| `-i, --id string`     | GitHub Gist ID to share                           |
| `-u, --users strings` | GitHub usernames to share with (comma-separated)  |
| `-r, --readonly`      | Share with read-only access (default true)        |
| `-l, --url`           | Generate a shareable URL                          |
| `-e, --expiry int`    | Expiry time for shareable URL in days (default 7) |

**Examples**:

```bash
# Share with specific users
envi share -i YOUR_GIST_ID -u user1,user2

# Generate shareable URL
envi share -i YOUR_GIST_ID -l

# Set URL expiry to 14 days
envi share -i YOUR_GIST_ID -l -e 14
```

### merge

Merge multiple .env files or merge with a remote Gist .env file.

**Usage**: `envi merge [flags]`

**Flags**:

| Flag                    | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| `-f, --files strings`   | Paths to local .env files to merge (comma-separated)     |
| `-g, --gist string`     | GitHub Gist ID to merge with (will fetch remote .env)    |
| `-o, --output string`   | Output file path (default ".env")                        |
| `-w, --overwrite`       | Overwrite duplicates (remote file takes precedence)      |
| `-s, --skip-duplicates` | Skip duplicates (local file takes precedence)            |
| `-c, --keep-comments`   | Keep comments from all files (default true)              |
| `--backup`              | Create backup of output file if it exists (default true) |
| `--sort`                | Sort variables alphabetically                            |
| `--unmask`              | Unmask/decrypt values from remote Gist when merging      |

**Examples**:

```bash
# Merge multiple local files
envi merge -f .env.dev,.env.local -o .env.merged

# Merge local file with remote Gist
envi merge -f .env.local -g YOUR_GIST_ID

# Merge and sort alphabetically
envi merge -f .env.local -o .env.sorted --sort
```

**Output Example**:

```
Processing file: .env.local
Successfully merged .env files into .env.merged
Merged 9 variables
```

## Security and Best Practices

1. **Token Security**: Your GitHub token is stored securely in your system's credential manager.
2. **Encryption Options**:
   - Masked encryption (default): Variable names visible, values encrypted
   - Full encryption: Entire file encrypted
   - No encryption: Plain text storage (not recommended for sensitive data)
3. **Key File vs Password**: Key files provide better security than passwords for encryption.
4. **Sharing Securely**: Always use encryption when sharing environment variables.
