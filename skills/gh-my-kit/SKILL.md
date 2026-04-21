---
name: gh-my-kit
description: gh-my-kit GitHub CLI extension for managing gists — including converting gists to repositories, copying gists between hosts, and migrating gists with full git history. Use when performing gist operations across GitHub hosts or converting gists into regular repositories.
---

# gh-my-kit

A personal [GitHub CLI](https://cli.github.com/) extension kit providing utilities for gist management across GitHub instances.

## Prerequisites

### Installation

```sh
gh extension install srz-zumix/gh-my-kit
```

### Verify Installation

```sh
gh my-kit --help
```

## CLI Structure

```
gh my-kit
├── gist                    # Gist management
│   ├── convert             # Convert gists to repositories
│   ├── copy                # Copy gists (latest content only)
│   └── migrate             # Migrate gists (preserving git history)
├── completion              # Shell completion scripts
└── skills                  # Agent skill management
```

## Gists (gh my-kit gist)

### Convert Gist to Repository

Convert one or more gists to regular GitHub repositories, preserving the full git
history via `git clone --mirror` + `git push --mirror`.

The repository name is derived from the gist description (or the gist ID if the
description is empty). The repository visibility defaults to the gist's own
visibility (public/private). By default, the default branch is renamed from
`master` to `main` after conversion.

```sh
gh my-kit gist convert <gist-id...> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--src <host>` | `-s` | current host from `gh auth` | Source GitHub host where the gist resides |
| `--src-token <token>` | | | Token for the source GitHub host |
| `--dst <host>` | `-d` | current host from `gh auth` | Destination GitHub host where the repository will be created |
| `--dst-token <token>` | | | Token for the destination GitHub host |
| `--name <name>` | | derived from gist description | Repository name (only valid with a single gist ID) |
| `--owner <owner>` | `-o` | authenticated user | Organization to create the repository under |
| `--visibility <visibility>` | `-v` | inherit from gist | Visibility of the created repository (`public`, `private`, `internal`) |
| `--no-rename-branch` | | false | Disable renaming the default branch from `master` to `main` |
| `--dryrun` | `-n` | false | Show what would be converted without making changes |

```sh
# Convert a gist to a repository (name derived from gist description)
gh my-kit gist convert abc123

# Convert a gist with a specific repository name
gh my-kit gist convert abc123 --name my-repo

# Convert a gist and create the repository under an organization
gh my-kit gist convert abc123 --owner my-org

# Convert multiple gists (names derived automatically)
gh my-kit gist convert abc123 def456

# Convert a gist to a repository on a GHES instance
gh my-kit gist convert abc123 --dst ghes.example.com --dst-token <token>

# Dry run: show what would be created without making changes
gh my-kit gist convert abc123 --dryrun

# Convert without renaming default branch from master to main
gh my-kit gist convert abc123 --no-rename-branch
```

### Copy Gists

Copy gists from one GitHub host to another. Only the latest file content is
copied; git history is not preserved. Use `gist migrate` to preserve full git
history.

If no gist IDs are provided, all gists of the authenticated source user are
copied.

```sh
gh my-kit gist copy [gist-id...] [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--src <host>` | `-s` | Source GitHub host (default: current host from `gh auth`) |
| `--dst <host>` | `-d` | Destination GitHub host (default: current host from `gh auth`) |
| `--src-token <token>` | | Token for the source GitHub host |
| `--dst-token <token>` | | Token for the destination GitHub host |
| `--dryrun` | `-n` | Show what would be copied without making changes |

```sh
# Copy all gists from github.com to a GHES instance
gh my-kit gist copy --dst ghes.example.com --dst-token <token>

# Copy specific gists
gh my-kit gist copy abc123 def456 --dst ghes.example.com --dst-token <token>

# Copy between two GHES instances
gh my-kit gist copy \
  --src src.example.com --src-token <src-token> \
  --dst dst.example.com --dst-token <dst-token>

# Dry run: show what would be copied without making changes
gh my-kit gist copy --dst ghes.example.com --dst-token <token> --dryrun
```

### Migrate Gists

Migrate gists from one GitHub host to another, preserving the full git history
via `git clone --mirror` + `git push --mirror`.

If no gist IDs are provided, all gists of the authenticated source user are
migrated.

```sh
gh my-kit gist migrate [gist-id...] [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--src <host>` | `-s` | Source GitHub host (default: current host from `gh auth`) |
| `--dst <host>` | `-d` | Destination GitHub host (default: current host from `gh auth`) |
| `--src-token <token>` | | Token for the source GitHub host |
| `--dst-token <token>` | | Token for the destination GitHub host |
| `--dryrun` | `-n` | Show what would be migrated without making changes |

```sh
# Migrate all gists from github.com to a GHES instance
gh my-kit gist migrate --dst ghes.example.com --dst-token <token>

# Migrate specific gists
gh my-kit gist migrate abc123 def456 --dst ghes.example.com --dst-token <token>

# Migrate between two GHES instances
gh my-kit gist migrate \
  --src src.example.com --src-token <src-token> \
  --dst dst.example.com --dst-token <dst-token>

# Dry run: show what would be migrated without making changes
gh my-kit gist migrate --dst ghes.example.com --dst-token <token> --dryrun
```

### copy vs migrate

| | `gist copy` | `gist migrate` |
|-|-------------|----------------|
| File content | ✅ | ✅ |
| Git history | ❌ | ✅ |

Use `gist copy` for a quick content-only transfer.
Use `gist migrate` when you need to preserve the full commit history.

## Completion (gh my-kit completion)

Generate shell completion scripts for gh-my-kit.

**Note:** gh CLI does not natively support extension completion. gh-my-kit
provides a patch script as a workaround. See the
[Shell Completion Guide](https://github.com/srz-zumix/go-gh-extension/blob/main/docs/shell-completion.md)
for setup instructions.

```sh
# Show completion help
gh my-kit completion --help
```

## Skills (gh my-kit skills)

Manage bundled agent skills for AI assistants.

```sh
gh my-kit skills [subcommand] [args...]
```

For details, see [Songmu/skillsmith](https://github.com/Songmu/skillsmith).

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GH_MY_KIT_NO_DOTENV` | Set to any non-empty value to disable automatic loading of the `.env` file |

## Common Workflows

### Convert All Gists to Repositories

```sh
# List all gist IDs first (using gh gist list)
gh gist list --json id --jq '.[].id'

# Convert each gist (names derived from descriptions)
gh my-kit gist convert <gist-id1> <gist-id2> ...
```

### Migrate All Gists to GHES

```sh
# Migrate all gists with full history
gh my-kit gist migrate \
  --dst ghes.example.com \
  --dst-token <ghes-token>
```

### Cross-Host Gist Backup

```sh
# Dry run to verify what will be migrated
gh my-kit gist migrate \
  --dst backup.example.com \
  --dst-token <token> \
  --dryrun

# Perform the actual migration
gh my-kit gist migrate \
  --dst backup.example.com \
  --dst-token <token>
```

## Getting Help

```sh
# General help
gh my-kit --help

# Command help
gh my-kit gist --help
gh my-kit gist convert --help
gh my-kit gist copy --help
gh my-kit gist migrate --help
```

## References

- Repository: <https://github.com/srz-zumix/gh-my-kit>
- Shell Completion Guide: <https://github.com/srz-zumix/go-gh-extension/blob/main/docs/shell-completion.md>
- skillsmith: <https://github.com/Songmu/skillsmith>
