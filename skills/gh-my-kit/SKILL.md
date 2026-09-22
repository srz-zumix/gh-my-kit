---
name: gh-my-kit
description: gh-my-kit GitHub CLI extension for managing GitHub Gists (converting gists to repositories, copying gist files across hosts, and migrating gist history between GitHub instances) and for dumping a GitHub user's activity to Markdown/JSON files.
---

# gh-my-kit

Personal [GitHub CLI](https://cli.github.com/) extension kit for advanced gist management and user activity dumps.

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
├── activity            # Dump a user's GitHub activity to Markdown/JSON files
├── completion          # Shell completion scripts
└── gist                # Gist management commands
    ├── convert         # Convert gists to repositories
    ├── copy            # Copy gists between hosts (file content only)
    └── migrate         # Migrate gists between hosts (with git history)
```

## Commands

### activity

Collect a GitHub user's activity (profile, followers/following, organizations, events, contributions, pull requests, issues, reviews, comments, repositories, starred/watched repositories, gists, packages, projects, discussions, and notifications) and write it to a directory as Markdown and JSON files, one pair per activity kind. The Markdown files are a condensed, human readable list, while the JSON files keep every field returned by the GitHub API.

Owner-scoped data (events, contributions, pulls, issues, reviews, comments, repos, packages, projects, discussions) is grouped under `<output>/<owner>/`. Non owner-scoped data (profile, followers, following, orgs, gists, starred, watching, notifications) is written under `<output>/`. A `summary.md` overview and an `AGENTS.md` guide describing the output layout and caveats for AI agents are always written at the top of `<output>/`.

If no user is given, the authenticated user is used. Notifications, private events, and watched repositories are only available for the authenticated user and are skipped with a warning otherwise.

The command fails when the output directory already contains files. Pass `--force` to empty the directory before writing.

`--period`/`--since`/`--until` bound events, contributions, pulls, issues, reviews, comments, gists (by creation date), and starred repositories (by star date). Followers, following, organizations, and watched repositories are always the current list, since GitHub does not expose when those relationships were created.

```sh
gh my-kit activity [user] [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host <host>` | | current host from `gh auth` | GitHub host to query |
| `--output <dir>` | `-o` | `./activity-<user>-<YYYYMMDD>` | Output directory |
| `--mode <summary\|detail>` | | `detail` | Markdown detail level: `detail` renders each record as a section with its date, link and body text; `summary` renders one titled link per record |
| `--period <N>d\|w\|m\|y` | | `30d` | Period to collect: relative (`30d`, `4w`, `6m`, `1y`) or a fiscal year (`FY26`, `FY26H1`, `FY26Q1`..`FY26Q4`; fiscal year starts in April) |
| `--since <date>` | | | Collect activity since this date (RFC3339 or `YYYY-MM-DD`) |
| `--until <date>` | | | Collect activity until this date (RFC3339 or `YYYY-MM-DD`) |
| `--include <kind,...>` | | all kinds | Only collect these activity kinds |
| `--exclude <kind,...>` | | | Exclude these activity kinds |
| `--skip-empty` | | `false` | Do not write output files for activity kinds with no collected data |
| `--force` | | `false` | Empty the output directory before writing instead of failing when it is not empty |
| `--format json` | | | Print the full result as JSON to stdout instead of writing files |

```sh
# Dump the authenticated user's activity for the last 30 days
gh my-kit activity

# Dump another user's activity for the last year, as a condensed link list
gh my-kit activity octocat --period 1y --mode summary

# Dump activity for a fiscal year (April 2026 - March 2027)
gh my-kit activity octocat --period FY26

# Only collect pull requests and issues, printed as JSON
gh my-kit activity octocat --include pulls,issues --format json
```

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
| `--src-token <token>` | | | Token for the source GitHub host (optional if the source is the current `gh auth` host) |
| `--dst <host>` | `-d` | current host from `gh auth` | Destination GitHub host where the repository will be created |
| `--dst-token <token>` | | | Token for the destination GitHub host (optional if the destination is the current `gh auth` host) |
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
gh my-kit gist convert abc123 --dst ghes.example.com --dst-token <dst-token>

# Dry run: show what would be created without making changes
gh my-kit gist convert abc123 --dryrun

# Convert without renaming default branch from master to main
gh my-kit gist convert abc123 --no-rename-branch
```

### gist copy

Copy gists from one GitHub host to another (latest file content only, no git history).
If `[gist-id...]` is omitted, all gists belonging to the authenticated user are copied.

```sh
gh my-kit gist copy [gist-id...] [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--src <host>` | `-s` | current host from `gh auth` | Source GitHub host |
| `--dst <host>` | `-d` | current host from `gh auth` | Destination GitHub host |
| `--src-token <token>` | | | Token for the source GitHub host (optional if the source is the current `gh auth` host) |
| `--dst-token <token>` | | | Token for the destination GitHub host (optional if the destination is the current `gh auth` host) |
| `--dryrun` | `-n` | false | Show what would be copied without making changes |

```sh
# Copy all gists from github.com to a GHES instance
gh my-kit gist copy --dst ghes.example.com --dst-token <dst-token>

# Copy specific gists
gh my-kit gist copy abc123 def456 --dst ghes.example.com --dst-token <dst-token>

# Copy between two GHES instances
gh my-kit gist copy \
  --src src.example.com --src-token <src-token> \
  --dst dst.example.com --dst-token <dst-token>

# Dry run: show what would be copied without making changes
gh my-kit gist copy --dst ghes.example.com --dryrun
```

### gist migrate

Migrate gists from one GitHub host to another, preserving the full git history via `git clone --mirror` + `git push --mirror`.
If `[gist-id...]` is omitted, all gists belonging to the authenticated user are migrated.

```sh
gh my-kit gist migrate [gist-id...] [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--src <host>` | `-s` | current host from `gh auth` | Source GitHub host |
| `--dst <host>` | `-d` | current host from `gh auth` | Destination GitHub host |
| `--src-token <token>` | | | Token for the source GitHub host (optional if the source is the current `gh auth` host) |
| `--dst-token <token>` | | | Token for the destination GitHub host (optional if the destination is the current `gh auth` host) |
| `--dryrun` | `-n` | false | Show what would be migrated without making changes |

```sh
# Migrate all gists from github.com to a GHES instance
gh my-kit gist migrate --dst ghes.example.com --dst-token <dst-token>

# Migrate specific gists
gh my-kit gist migrate abc123 def456 --dst ghes.example.com --dst-token <dst-token>

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
  --dst-token <dst-token>
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
