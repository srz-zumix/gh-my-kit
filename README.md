# gh-my-kit

My personal [GitHub CLI](https://cli.github.com/) extension kit.

## Installation

```sh
gh extension install srz-zumix/gh-my-kit
```

## Commands

### `gist`

Commands for managing GitHub Gists.

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
