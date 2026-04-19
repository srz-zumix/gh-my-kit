# gh-my-kit

My personal [GitHub CLI](https://cli.github.com/) extension kit.

## Installation

```sh
gh extension install srz-zumix/gh-my-kit
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GH_MY_KIT_NO_DOTENV` | Set to any non-empty value to disable automatic loading of the `.env` file |

## Commands

### `gist`

Commands for managing GitHub Gists.

#### `gist convert <gist-id...>`

Convert one or more gists to regular GitHub repositories, preserving the full git history via `git clone --mirror` + `git push --mirror`.

The repository name is derived from the gist description (or the gist ID if the description is empty).
The repository visibility defaults to the gist's own visibility (public/private).
By default, the default branch is renamed from `master` to `main` after conversion.

```
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

**Examples:**

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

#### `gist copy [gist-id...]`

Copy gists from one GitHub host to another (latest file content only, no git history).

```
gh my-kit gist copy [gist-id...] [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--src <host>` | `-s` | Source GitHub host (default: current host from `gh auth`) |
| `--dst <host>` | `-d` | Destination GitHub host (default: current host from `gh auth`) |
| `--src-token <token>` | | Token for the source GitHub host |
| `--dst-token <token>` | | Token for the destination GitHub host |
| `--dryrun` | `-n` | Show what would be copied without making changes |

**Examples:**

```sh
# Copy all gists from github.com to a GHES instance
gh my-kit gist copy --dst ghes.example.com --dst-token <token>

# Copy specific gists
gh my-kit gist copy abc123 def456 --dst ghes.example.com --dst-token <token>

# Copy between two GHES instances
gh my-kit gist copy \
  --src src.example.com --src-token <src-token> \
  --dst dst.example.com --dst-token <dst-token>
```

#### `gist migrate [gist-id...]`

Migrate gists from one GitHub host to another, preserving the full git history via `git clone --mirror` + `git push --mirror`.

```
gh my-kit gist migrate [gist-id...] [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--src <host>` | `-s` | Source GitHub host (default: current host from `gh auth`) |
| `--dst <host>` | `-d` | Destination GitHub host (default: current host from `gh auth`) |
| `--src-token <token>` | | Token for the source GitHub host |
| `--dst-token <token>` | | Token for the destination GitHub host |
| `--dryrun` | `-n` | Show what would be migrated without making changes |

**Examples:**

```sh
# Migrate all gists from github.com to a GHES instance
gh my-kit gist migrate --dst ghes.example.com --dst-token <token>

# Migrate specific gists
gh my-kit gist migrate abc123 def456 --dst ghes.example.com --dst-token <token>

# Migrate between two GHES instances
gh my-kit gist migrate \
  --src src.example.com --src-token <src-token> \
  --dst dst.example.com --dst-token <dst-token>
```

> **copy vs migrate**
>
> | | `gist copy` | `gist migrate` |
> |-|-------------|----------------|
> | File content | ✅ | ✅ |
> | Git history | ❌ | ✅ |

### `completion`

Generate shell completion scripts.

```sh
gh my-kit completion --help
```
