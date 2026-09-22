package activity

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v90/github"
)

// Entry is a single renderable record: a title/summary line plus an optional
// link, creation time and long-form body, used for both summary and detail
// rendering.
type Entry struct {
	Title     string
	URL       string
	CreatedAt string
	Body      string
}

// KindSection is the rendered content for one Kind within one scope (root or
// a specific owner).
type KindSection struct {
	Kind    Kind
	Count   int
	Entries []Entry
}

// IsEmpty reports whether the section holds no collected data.
func (s KindSection) IsEmpty() bool {
	return s.Count == 0 && len(s.Entries) == 0
}

// Section builds the KindSection for kind using the root-level Result data.
// Owner-scoped kinds are built via OwnerActivity.Section instead.
func (r *Result) Section(kind Kind) KindSection {
	switch kind {
	case KindProfile:
		if r.Profile == nil {
			return KindSection{Kind: kind}
		}
		entry := userEntry(r.Profile)
		entry.CreatedAt = r.Profile.GetCreatedAt().Format(dateLayout)
		entry.Body = profileDetails(r.Profile)
		return KindSection{Kind: kind, Count: 1, Entries: []Entry{entry}}
	case KindFollowers:
		return usersSection(kind, r.Followers)
	case KindFollowing:
		return usersSection(kind, r.Following)
	case KindOrgs:
		entries := make([]Entry, 0, len(r.Orgs))
		for _, o := range r.Orgs {
			entries = append(entries, Entry{Title: o.GetLogin(), URL: o.GetHTMLURL(), Body: o.GetDescription()})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindGists:
		entries := make([]Entry, 0, len(r.Gists))
		for _, g := range r.Gists {
			entries = append(entries, Entry{
				Title:     g.GetDescription(),
				URL:       g.GetHTMLURL(),
				CreatedAt: g.GetCreatedAt().Format(dateLayout),
				Body:      gistFileNames(g),
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindStarred:
		entries := make([]Entry, 0, len(r.Starred))
		for _, s := range r.Starred {
			repo := s.GetRepository()
			entries = append(entries, Entry{
				Title:     repo.GetFullName(),
				URL:       repo.GetHTMLURL(),
				CreatedAt: s.GetStarredAt().Format(dateLayout),
				Body:      repo.GetDescription(),
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindWatching:
		entries := make([]Entry, 0, len(r.Watching))
		for _, repo := range r.Watching {
			entries = append(entries, Entry{Title: repo.GetFullName(), URL: repo.GetHTMLURL(), Body: repo.GetDescription()})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindNotifications:
		entries := make([]Entry, 0, len(r.Notifications))
		for _, n := range r.Notifications {
			entries = append(entries, Entry{
				Title:     fmt.Sprintf("[%s] %s", n.GetRepository().GetFullName(), n.GetSubject().GetTitle()),
				URL:       n.GetSubject().GetURL(),
				CreatedAt: n.GetUpdatedAt().Format(dateLayout),
				Body: fieldList(
					"Type", n.GetSubject().GetType(),
					"Reason", n.GetReason(),
				),
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	default:
		return KindSection{Kind: kind}
	}
}

// Section builds the KindSection for kind scoped to this owner.
func (oa *OwnerActivity) Section(kind Kind) KindSection {
	switch kind {
	case KindRepos:
		entries := make([]Entry, 0, len(oa.Repos))
		for _, repo := range oa.Repos {
			entries = append(entries, Entry{
				Title:     repo.GetFullName(),
				URL:       repo.GetHTMLURL(),
				CreatedAt: repo.GetCreatedAt().Format(dateLayout),
				Body:      repo.GetDescription(),
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindEvents:
		entries := make([]Entry, 0, len(oa.Events))
		for _, e := range oa.Events {
			entries = append(entries, Entry{
				Title:     fmt.Sprintf("[%s] %s", e.GetRepo().GetName(), e.GetType()),
				CreatedAt: e.GetCreatedAt().Format(dateLayout),
				Body:      eventSummary(e),
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindContributions:
		count := 0
		var entries []Entry
		if oa.Contributions != nil {
			count = oa.Contributions.TotalCommitContributions
			for _, rc := range oa.Contributions.CommitContributionsByRepository {
				entries = append(entries, Entry{Title: fmt.Sprintf("%s (%d commits)", rc.NameWithOwner, rc.Contributions)})
			}
		}
		return KindSection{Kind: kind, Count: count, Entries: entries}
	case KindPulls:
		return issuesSection(kind, oa.Pulls)
	case KindIssues:
		return issuesSection(kind, oa.Issues)
	case KindReviews:
		return issuesSection(kind, oa.Reviews)
	case KindComments:
		entries := make([]Entry, 0, len(oa.Comments))
		for _, c := range oa.Comments {
			entries = append(entries, Entry{
				Title:     fmt.Sprintf("%s#%d %s", c.Repository, c.Number, c.Title),
				URL:       c.URL,
				CreatedAt: c.CreatedAt.Format(dateLayout),
				Body:      c.Body,
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindPackages:
		entries := make([]Entry, 0, len(oa.Packages))
		for _, p := range oa.Packages {
			entries = append(entries, Entry{
				Title:     fmt.Sprintf("%s (%s)", p.GetName(), p.GetPackageType()),
				URL:       p.GetHTMLURL(),
				CreatedAt: p.GetCreatedAt().Format(dateLayout),
				Body: fieldList(
					"Visibility", p.GetVisibility(),
					"Versions", fmt.Sprintf("%d", p.GetVersionCount()),
				),
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindProjects:
		entries := make([]Entry, 0, len(oa.Projects))
		for _, p := range oa.Projects {
			body := ""
			if p.ShortDescription != nil {
				body = *p.ShortDescription
			}
			entries = append(entries, Entry{Title: p.Title, URL: p.URL, Body: body})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	case KindDiscussions:
		entries := make([]Entry, 0, len(oa.Discussions))
		for _, d := range oa.Discussions {
			entries = append(entries, Entry{
				Title:     fmt.Sprintf("%s#%d %s", d.Repository.NameWithOwner, d.Number, d.Title),
				URL:       d.URL,
				CreatedAt: d.CreatedAt.Format(dateLayout),
				Body:      d.Body,
			})
		}
		return KindSection{Kind: kind, Count: len(entries), Entries: entries}
	default:
		return KindSection{Kind: kind}
	}
}

// OwnerLogins returns the owners in the Result, sorted alphabetically.
func (r *Result) OwnerLogins() []string {
	owners := make([]string, 0, len(r.Owners))
	for owner := range r.Owners {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners
}

func userEntry(u *github.User) Entry {
	return Entry{Title: u.GetLogin(), URL: u.GetHTMLURL()}
}

func usersSection(kind Kind, users []*github.User) KindSection {
	entries := make([]Entry, 0, len(users))
	for _, u := range users {
		entries = append(entries, userEntry(u))
	}
	return KindSection{Kind: kind, Count: len(entries), Entries: entries}
}

func issuesSection(kind Kind, issues []*github.Issue) KindSection {
	entries := make([]Entry, 0, len(issues))
	for _, i := range issues {
		entries = append(entries, Entry{
			Title:     fmt.Sprintf("%s %s", i.GetTitle(), i.GetState()),
			URL:       i.GetHTMLURL(),
			CreatedAt: i.GetCreatedAt().Format(dateLayout),
			Body:      i.GetBody(),
		})
	}
	return KindSection{Kind: kind, Count: len(entries), Entries: entries}
}

// gistFileNames returns the gist's file names, one per line, in a stable order.
func gistFileNames(g *github.Gist) string {
	names := make([]string, 0, len(g.Files))
	for name := range g.Files {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// fieldList renders label/value pairs as a Markdown list, skipping empty
// values. Pairs are given as alternating label and value arguments.
func fieldList(pairs ...string) string {
	var lines []string
	for i := 0; i+1 < len(pairs); i += 2 {
		if value := strings.TrimSpace(pairs[i+1]); value != "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", pairs[i], value))
		}
	}
	return strings.Join(lines, "\n")
}

// profileDetails renders the profile fields that the title and link drop,
// followed by the bio.
func profileDetails(u *github.User) string {
	fields := fieldList(
		"Name", u.GetName(),
		"Company", u.GetCompany(),
		"Location", u.GetLocation(),
		"Email", u.GetEmail(),
		"Blog", u.GetBlog(),
		"Twitter", u.GetTwitterUsername(),
		"Type", u.GetType(),
		"Public repositories", strconv.Itoa(u.GetPublicRepos()),
		"Public gists", strconv.Itoa(u.GetPublicGists()),
		"Followers", strconv.Itoa(u.GetFollowers()),
		"Following", strconv.Itoa(u.GetFollowing()),
		"Updated", u.GetUpdatedAt().Format(dateLayout),
	)

	bio := strings.TrimSpace(u.GetBio())
	if bio == "" {
		return fields
	}
	return fields + "\n\n" + bio
}

// eventSummary describes an event using the parts of its payload that carry
// text, since the payload shape differs per event type.
func eventSummary(e *github.Event) string {
	payload, err := e.ParsePayload()
	if err != nil {
		return ""
	}

	var lines []string
	switch p := payload.(type) {
	case *github.PushEvent:
		for _, c := range p.Commits {
			lines = append(lines, fmt.Sprintf("%s %s", c.GetSHA(), c.GetMessage()))
		}
	case *github.PullRequestEvent:
		lines = append(lines, fmt.Sprintf("%s #%d %s", p.GetAction(), p.GetNumber(), p.GetPullRequest().GetTitle()))
		lines = append(lines, p.GetPullRequest().GetBody())
	case *github.IssuesEvent:
		lines = append(lines, fmt.Sprintf("%s #%d %s", p.GetAction(), p.GetIssue().GetNumber(), p.GetIssue().GetTitle()))
		lines = append(lines, p.GetIssue().GetBody())
	case *github.IssueCommentEvent:
		lines = append(lines, fmt.Sprintf("%s #%d %s", p.GetAction(), p.GetIssue().GetNumber(), p.GetIssue().GetTitle()))
		lines = append(lines, p.GetComment().GetBody())
	case *github.PullRequestReviewEvent:
		lines = append(lines, fmt.Sprintf("%s #%d %s", p.GetAction(), p.GetPullRequest().GetNumber(), p.GetReview().GetState()))
		lines = append(lines, p.GetReview().GetBody())
	case *github.PullRequestReviewCommentEvent:
		lines = append(lines, fmt.Sprintf("%s #%d", p.GetAction(), p.GetPullRequest().GetNumber()))
		lines = append(lines, p.GetComment().GetBody())
	case *github.ReleaseEvent:
		lines = append(lines, fmt.Sprintf("%s %s", p.GetAction(), p.GetRelease().GetTagName()))
		lines = append(lines, p.GetRelease().GetBody())
	case *github.CreateEvent:
		lines = append(lines, fmt.Sprintf("created %s %s", p.GetRefType(), p.GetRef()))
	case *github.DeleteEvent:
		lines = append(lines, fmt.Sprintf("deleted %s %s", p.GetRefType(), p.GetRef()))
	}

	return strings.TrimSpace(strings.Join(nonEmpty(lines), "\n"))
}

func nonEmpty(lines []string) []string {
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			kept = append(kept, line)
		}
	}
	return kept
}

// rootKinds are kinds that are not scoped to a repository owner.
var rootKinds = []Kind{
	KindProfile, KindFollowers, KindFollowing, KindOrgs,
	KindGists, KindStarred, KindWatching, KindNotifications,
}

// isRootKind reports whether kind is rendered at the root of the output
// directory rather than grouped per owner.
func isRootKind(kind Kind) bool {
	return hasKind(rootKinds, kind)
}
