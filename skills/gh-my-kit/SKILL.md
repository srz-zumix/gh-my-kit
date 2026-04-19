---
name: gh-my-kit
description: gh-my-kit GitHub CLI extension for managing GitHub Gists — including converting gists to repositories, copying gist files across hosts, and migrating gist history between GitHub instances (github.com and GitHub Enterprise Server).
---

# gh-my-kit

Personal [GitHub CLI](https://cli.github.com/) extension kit for advanced gist management.

## Installation

```sh
gh extension install srz-zumix/gh-my-kit
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GH_MY_KIT_NO_DOTENV` | Set to any non-empty value to disable automatic loading of the `.env` file |

## CLI Structure

```
gh my-kit
├── completion          # Shell completion scripts
└── gist                # Gist management commands
    ├── convert         # Convert gists to repositories
    ├── copy            # Copy gists between hosts (file content only)
    └── migrate         # Migrate gists between hosts (with git history)
```

## Commands

### gist convert

Convert one or more gists to regular GitHub repositories, preserving the full git history via `git clone --mirror` + `git push --mirror`.

The repository name is derived from the gist description (or the gist ID if the description is empty).
The repository visibility defaults to the gist's own visibility (public/private).
By default, the default branch is renamed from `master` to `main` after conversion.

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

### gist copy

Copy gists from one GitHub host to another (latest file content only, no git history).

```sh
gh my-kit gist copy [gist-id...] [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--src <host>` | `-s` | current host from `gh auth` | Source GitHub host |
| `--dst <host>` | `-d` | current host from `gh auth` | Destination GitHub host |
| `--src-token <token>` | | | Token for the source GitHub host |
| `--dst-token <token>` | | | Token for the destination GitHub host |
| `--dryrun` | `-n` | false | Show what would be copied without making changes |

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
gh my-kit gist copy --dst ghes.example.com --dryrun
```

### gist migrate

Migrate gists from one GitHub host to another, preserving the full git history via `git clone --mirror` + `git push --mirror`.

```sh
gh my-kit gist migrate [gist-id...] [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--src <host>` | `-s` | current host from `gh auth` | Source GitHub host |
| `--dst <host>` | `-d` | current host from `gh auth` | Destination GitHub host |
| `--src-token <token>` | | | Token for the source GitHub host |
| `--dst-token <token>` | | | Token for the destination GitHub host |
| `--dryrun` | `-n` | false | Show what would be migrated without making changes |

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
gh my-kit gist migrate --dst ghes.example.com --dryrun
```

### copy vs migrate

| | `gist copy` | `gist migrate` |
|-|-------------|----------------|
| File content | ✅ | ✅ |
| Git history | ❌ | ✅ |

### completion

Generate shell completion scripts.

```sh
gh my-kit completion -s bash  > ~/.gh-my-kit-complete.bash
gh my-kit completion -s zsh   > ~/.gh-my-kit-complete.zsh
gh my-kit completion -s fish  > ~/.gh-my-kit-complete.fish
```

## Common Workflows

### Migrate all gists to GitHub Enterprise Server

```sh
# Migrate all gists with full history
gh my-kit gist migrate \
  --dst ghes.example.com \
  --dst-token <ghes-token>
```

### Convert a gist to an organization repository

```sh
# Convert gist to a private repository under an org
gh my-kit gist convert abc123 \
  --owner my-org \
  --visibility private
```

### Cross-host full gist migration

```sh
# Full history migration between two GHES instances
gh my-kit gist migrate \
  --src src.example.com --src-token <src-token> \
  --dst dst.example.com --dst-token <dst-token>
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

- Repository: https://github.com/srz-zumix/gh-my-kit
- GitHub CLI Extensions: https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions
