package activity

import (
	"fmt"
	"strings"
)

// kindDescriptions documents what each kind file holds, for the generated AGENTS.md.
var kindDescriptions = map[Kind]string{
	KindProfile:       "The user's GitHub profile: name, bio, company, location and public counts.",
	KindFollowers:     "Users who follow the collected user.",
	KindFollowing:     "Users the collected user follows.",
	KindOrgs:          "Organizations the collected user is a public member of.",
	KindEvents:        "GitHub events (pushes, pull request/issue actions, releases, etc.) created in the period.",
	KindContributions: "Commit contribution totals per repository for the period.",
	KindPulls:         "Pull requests authored by the user and created in the period.",
	KindIssues:        "Issues authored by the user and created in the period.",
	KindReviews:       "Pull requests the user reviewed that were updated in the period.",
	KindComments:      "Issue and pull request comments authored by the user in the period, including the comment body.",
	KindRepos:         "Repositories owned by the user.",
	KindStarred:       "Repositories the user starred in the period.",
	KindWatching:      "Repositories the user watches.",
	KindGists:         "Gists the user created in the period.",
	KindPackages:      "GitHub Packages published under the owner.",
	KindProjects:      "Projects (v2) owned by the user or their organizations.",
	KindDiscussions:   "Discussions the user participated in.",
	KindNotifications: "Notifications updated in the period.",
}

// periodBoundedKinds are limited to [Since, Until]: each record carries a
// timestamp (created/updated/starred) that the collector filters on.
var periodBoundedKinds = []Kind{
	KindEvents, KindContributions, KindPulls, KindIssues, KindReviews,
	KindComments, KindGists, KindStarred, KindNotifications,
}

// unboundedKinds reflect the current state and are NOT limited to the period,
// because GitHub does not expose a usable timestamp for the collector to filter
// on. This covers snapshot relationships (followers/following/orgs/watching),
// the profile, and owner-scoped listings the collector does not date-filter
// (repos/packages/projects/discussions).
var unboundedKinds = []Kind{
	KindProfile, KindFollowers, KindFollowing, KindOrgs, KindWatching,
	KindRepos, KindPackages, KindProjects, KindDiscussions,
}

// RenderAgentsGuide renders an AGENTS.md that explains the layout and the
// meaning of the dumped files to an AI agent reading the output directory.
func RenderAgentsGuide(r *Result, mode Mode, kinds []Kind, skipEmpty bool) string {
	var b strings.Builder

	b.WriteString("# AGENTS.md\n\n")
	fmt.Fprintf(&b, "This directory is a dump of GitHub activity for `%s`, produced by `gh my-kit activity`.\n", r.Username)
	b.WriteString("It is intended to be read by AI agents as source data for summaries, reports and reviews.\n\n")

	b.WriteString("## Metadata\n\n")
	fmt.Fprintf(&b, "- User: `%s`\n", r.Username)
	fmt.Fprintf(&b, "- Period: `%s` .. `%s`\n", r.Since.Format(dateLayout), r.Until.Format(dateLayout))
	fmt.Fprintf(&b, "- Render mode: `%s`\n", mode)
	fmt.Fprintf(&b, "- Collection warnings: %d\n\n", len(r.Warnings))

	b.WriteString("## Directory layout\n\n")
	b.WriteString("```text\n")
	b.WriteString(".\n")
	b.WriteString("├── AGENTS.md     # this file\n")
	b.WriteString("├── summary.md    # record counts per kind and scope, plus collection warnings\n")
	b.WriteString("├── <kind>.md     # account-scoped data, not tied to a repository owner\n")
	b.WriteString("├── <kind>.json\n")
	b.WriteString("└── <owner>/      # one directory per repository owner (user or organization)\n")
	b.WriteString("    ├── <kind>.md\n")
	b.WriteString("    └── <kind>.json\n")
	b.WriteString("```\n\n")

	b.WriteString("## How to read this dump\n\n")
	b.WriteString("- Start from `summary.md`. It lists the record count for every kind and scope, so you can skip empty files.\n")
	b.WriteString("- `<kind>.md` is a human readable list: a `# <kind>` heading, a `Count: N` line, then one bullet per record. It is a condensed view and drops most fields.\n")
	b.WriteString("- `<kind>.json` is `{\"kind\", \"owner\", \"count\", \"data\"}`, where `data` holds the records exactly as the GitHub API returned them, with every available field. Use the JSON files whenever you need more than a title and a link.\n")
	fmt.Fprintf(&b, "- This dump was rendered in `%s` mode: %s\n", mode, modeDescription(mode))
	if skipEmpty {
		b.WriteString("- Kinds with no collected data were skipped, so a file listed below may be absent. `summary.md` still reports their count.\n")
	}
	fmt.Fprintf(&b, "- A directory named after a repository owner groups activity in that owner's repositories. `%s/` is the collected user's own account.\n\n", r.Username)

	writeKindTable(&b, "Account-scoped files", "", kinds, isRootKind)
	writeKindTable(&b, "Owner-scoped files", "<owner>/", kinds, func(k Kind) bool { return !isRootKind(k) })

	if owners := r.OwnerLogins(); len(owners) > 0 {
		b.WriteString("## Owner directories in this dump\n\n")
		for _, owner := range owners {
			if owner == "" {
				continue
			}
			fmt.Fprintf(&b, "- `%s/`\n", owner)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Caveats\n\n")
	fmt.Fprintf(&b, "- These kinds are limited to the period above: %s.\n", quotedKindList(periodBoundedKinds))
	fmt.Fprintf(&b, "- These kinds reflect the current state and are **not** limited to the period above, because GitHub does not expose a usable timestamp for them: %s.\n", quotedKindList(unboundedKinds))
	b.WriteString("- `watching` and `notifications` are only available for the authenticated user; for any other user they are skipped.\n")
	b.WriteString("- `gists` includes secret gists only for the authenticated user; for any other user only public gists are collected.\n")
	fmt.Fprintf(&b, "- `comments` is collected from at most %d matched issues/pull requests, so it can be incomplete for very active users.\n", maxCommentIssues)
	b.WriteString("- A count of `0` means nothing was collected, which may be a collection failure rather than an absence of activity. Check the `Warnings` section of `summary.md` before concluding that a kind is empty.\n")

	return b.String()
}

// writeKindTable writes a Markdown table describing the files for the kinds in
// kinds that match the scope predicate. Nothing is written when none match.
func writeKindTable(b *strings.Builder, title, prefix string, kinds []Kind, match func(Kind) bool) {
	var rows []Kind
	for _, kind := range kinds {
		if match(kind) {
			rows = append(rows, kind)
		}
	}
	if len(rows) == 0 {
		return
	}

	fmt.Fprintf(b, "## %s\n\n", title)
	b.WriteString("| File | Description |\n")
	b.WriteString("|------|-------------|\n")
	for _, kind := range rows {
		fmt.Fprintf(b, "| `%s%s.md` / `%s%s.json` | %s |\n", prefix, kind, prefix, kind, kindDescriptions[kind])
	}
	b.WriteString("\n")
}

func modeDescription(mode Mode) string {
	if mode == ModeDetail {
		return "every record is a section with its date, link and full body text (issue/PR/comment bodies, descriptions, review comments)."
	}
	return "each record is a single titled link, without its date or body."
}

func quotedKindList(kinds []Kind) string {
	names := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		names = append(names, fmt.Sprintf("`%s`", kind))
	}
	return strings.Join(names, ", ")
}
