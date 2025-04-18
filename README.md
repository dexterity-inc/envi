# Envi CLI

A simple CLI tool to push and pull `.env` files to/from GitHub Gists for secure storage and sharing.

## Installation

1. Make sure you have Go installed
2. Clone this repository
3. Build the binary:

```bash
go build -o envi
```

4. Move the binary to your PATH (optional):

```bash
sudo mv envi /usr/local/bin/
```

## Usage

### Set up GitHub Personal Access Token

You have two options to set your GitHub token:

1. Using the config command (recommended):

```bash
envi config --token your_github_personal_access_token
```

This will save your token in `~/.envi/config.json` and use it for all future commands.

2. Using an environment variable (temporary):

```bash
export GITHUB_TOKEN=your_github_personal_access_token
```

In both cases, you'll need a GitHub token with the `gist` scope.

### Push .env file to GitHub Gist

To push the `.env` file in your current directory to a new GitHub Gist:

```bash
envi push
```

The command will create a new Gist, output the Gist ID and URL, and save the ID for future updates.

#### Update an existing Gist

To update an existing Gist instead of creating a new one:

```bash
envi push --id GIST_ID
```

Replace `GIST_ID` with the ID of the Gist you want to update.

If you've previously pushed a Gist with `envi push`, the tool will remember the last Gist ID and offer to update it:

```bash
envi push
```

It will prompt you whether to use the saved Gist ID.

#### Disable saving Gist ID

If you don't want to save the Gist ID for future updates:

```bash
envi push --save=false
```

### Pull .env file from GitHub Gist

There are several ways to pull an `.env` file from a GitHub Gist:

1. Using a specific Gist ID:

```bash
envi pull --id GIST_ID
```

Replace `GIST_ID` with the ID of the Gist you want to pull from.

2. Using the saved Gist ID (if you've previously pushed or pulled):

```bash
envi pull
```

This will use the Gist ID saved in your config file, after asking for confirmation.

#### Additional Pull Options

Create a backup of your existing `.env` file before overwriting:

```bash
envi pull --backup
```

This creates a timestamped backup file (e.g., `.env.backup.20230615-123045`).

Skip confirmation when overwriting an existing `.env` file:

```bash
envi pull --force
```

Combine options as needed:

```bash
envi pull --id GIST_ID --backup --force
```

### List Your Gists

View all your GitHub Gists containing `.env` files:

```bash
envi list
```

This displays a table of your Gists, highlighting the currently selected one.

#### Additional List Options

Show all Gists, not just those containing `.env` files:

```bash
envi list --all
```

Limit the number of Gists displayed:

```bash
envi list --limit 5
```

### Compare Local and Remote .env Files

Compare your local `.env` file with a remote one:

```bash
envi diff
```

This shows variables that are only in local, only in remote, or have different values.

To see the actual values (not just variable names):

```bash
envi diff --values
```

⚠️ Note: Using `--values` will display the actual values which may contain sensitive data.

### Merge Local and Remote .env Files

Merge your local `.env` file with a remote one:

```bash
envi merge
```

This will interactively prompt you to resolve conflicts between local and remote values.

#### Additional Merge Options

Automatically prefer local values for conflicts:

```bash
envi merge --local
```

Automatically prefer remote values for conflicts:

```bash
envi merge --remote
```

Push the merged result back to the Gist:

```bash
envi merge --push
```

Disable creating a backup (not recommended):

```bash
envi merge --backup=false
```

### Validate .env file against .env.example

To check if your `.env` file contains all the variables defined in `.env.example`:

```bash
envi validate
```

This will show you any missing variables. To automatically add the missing variables to your `.env` file:

```bash
envi validate --fix
```

Additional options:

- `--example`, `-e`: Specify a custom path to the example file (default: `.env.example`)

### View and Manage Configuration

To view your current configuration:

```bash
envi config
```

This will display your GitHub token status and any saved Gist ID.

To clear the saved Gist ID:

```bash
envi config --clear-gist
```

## Command Help

For more information about the commands:

```bash
# General help
envi --help

# Push command help
envi push --help

# Pull command help
envi pull --help

# List command help
envi list --help

# Diff command help
envi diff --help

# Merge command help
envi merge --help

# Config command help
envi config --help

# Validate command help
envi validate --help
```
