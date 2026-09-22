# gh-my-kit

My personal [GitHub CLI](https://cli.github.com/) extension kit.

## Installation

```sh
gh extension install srz-zumix/gh-my-kit
```

## Shell Completion

**Workaround Available!** While gh CLI doesn't natively support extension completion, we provide a patch script that enables it.

**Prerequisites:** Before setting up gh-my-kit completion, ensure gh CLI completion is configured for your shell. See [gh completion documentation](https://cli.github.com/manual/gh_completion) for setup instructions.

For detailed installation instructions and setup for each shell, see the [Shell Completion Guide](https://github.com/srz-zumix/go-gh-extension/blob/main/docs/shell-completion.md).

## Agent Skills

gh-my-kit bundles agent skills for AI. Use the `skills` subcommand to install and manage them.

```sh
gh my-kit skills [subcommand] [args...]
```

For details, see [Songmu/skillsmith](https://github.com/Songmu/skillsmith).

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GH_MY_KIT_NO_DOTENV` | Set to any non-empty value to disable automatic loading of the `.env` file |

## Commands

### `activity [user]`

Collect a GitHub user's activity (profile, followers/following, organizations, events, contributions, pull requests, issues, reviews, comments, repositories, starred/watched repositories, gists, packages, projects, discussions, and notifications) and write it to a directory as Markdown and JSON files, one pair per activity kind. The Markdown files are a condensed, human readable list, while the JSON files keep every field returned by the GitHub API.

Data scoped to a repository owner (events, contributions, pulls, issues, reviews, comments, repos, packages, projects, discussions) is grouped under `<output>/<owner>/`. Data that is not owner-scoped (profile, followers, following, orgs, gists, starred, watching, notifications) is written directly under `<output>/`. A `summary.md` overview listing counts per kind and an `AGENTS.md` guide describing the output layout, the meaning of each file and the collection caveats for AI agents are always written at the top of `<output>/`.

If no user is given, the authenticated user is used. Notifications, private events, and watched repositories are only available for the authenticated user and are skipped with a warning when a different user is specified.

The command fails when the output directory already contains files. Pass `--force` to empty the directory before writing.

`--period`/`--since`/`--until` bound events, contributions, pull requests, issues, reviews, comments, gists (by creation date), and starred repositories (by star date). Followers, following, organizations, and watched repositories are always the current list, since GitHub does not expose when those relationships were created.

```sh
gh my-kit activity [user] [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host <host>` | | current host from `gh auth` | GitHub host to query |
| `--output <dir>` | `-o` | `./activity-<user>-<YYYYMMDD>` | Output directory |
| `--mode <summary\|detail>` | | `detail` | Markdown detail level: `detail` renders each record as a section with its date, link and body text (issue/PR/comment bodies, descriptions); `summary` renders one titled link per record |
| `--period <N>d\|w\|m\|y` | | `30d` | Period to collect: relative (`30d`, `4w`, `6m`, `1y`) or a fiscal year (`FY26`, `FY26H1`, `FY26Q1`..`FY26Q4`; fiscal year starts in April); mutually exclusive with `--since`/`--until` |
| `--since <date>` | | | Collect activity since this date (RFC3339 or `YYYY-MM-DD`); mutually exclusive with `--period` |
| `--until <date>` | | | Collect activity until this date (RFC3339 or `YYYY-MM-DD`); mutually exclusive with `--period` |
| `--include <kind,...>` | | all kinds | Only collect these activity kinds; mutually exclusive with `--exclude` |
| `--exclude <kind,...>` | | | Exclude these activity kinds; mutually exclusive with `--include` |
| `--skip-empty` | | `false` | Do not write output files for activity kinds with no collected data |
| `--force` | | `false` | Empty the output directory before writing instead of failing when it is not empty |
| `--format json` | | | Print the full collected result as JSON to stdout instead of writing files |

**Examples:**

```sh
# Dump the authenticated user's activity for the last 30 days
gh my-kit activity

# Dump another user's activity for the last year, as a condensed link list
gh my-kit activity octocat --period 1y --mode summary

# Dump activity for a fiscal year (April 2026 - March 2027)
gh my-kit activity octocat --period FY26

# Dump activity for the first half of a fiscal year (April - September 2026)
gh my-kit activity octocat --period FY26H1

# Dump activity for a specific date range to a custom directory
gh my-kit activity octocat --since 2024-01-01 --until 2024-03-31 --output ./out

# Skip files for activity kinds that collected no data
gh my-kit activity octocat --skip-empty

# Overwrite an existing output directory
gh my-kit activity octocat --output ./out --force

# Only collect pull requests and issues
gh my-kit activity octocat --include pulls,issues

# Print the full result as JSON instead of writing files
gh my-kit activity octocat --format json
```

### `gist`

Commands for managing GitHub Gists.

#### `gist convert <gist-id...>`

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

```sh
gh my-kit gist migrate [gist-id...] [flags]
```

| Flag | Short | Description |
| ------ | ------- | ------------- |
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
> | - | ------------- | ---------------- |
> | File content | ✅ | ✅ |
> | Git history | ❌ | ✅ |

### `completion`

Generate shell completion scripts.

```sh
gh my-kit completion --help
```
