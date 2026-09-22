package activity

import (
	"fmt"
	"time"

	"github.com/google/go-github/v90/github"
	extgh "github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// Result holds every piece of activity data collected for a user.
type Result struct {
	Username string    `json:"username"`
	Since    time.Time `json:"since"`
	Until    time.Time `json:"until"`

	// Data below is not scoped to a single repository owner.
	Profile       *github.User                `json:"profile,omitempty"`
	Followers     []*github.User              `json:"followers,omitempty"`
	Following     []*github.User              `json:"following,omitempty"`
	Orgs          []*github.Organization      `json:"orgs,omitempty"`
	Gists         []*github.Gist              `json:"gists,omitempty"`
	Starred       []*github.StarredRepository `json:"starred,omitempty"`
	Watching      []*github.Repository        `json:"watching,omitempty"`
	Notifications []*github.Notification      `json:"notifications,omitempty"`

	// Owners groups data that belongs to a single repository owner, keyed by
	// owner login. The key "" holds the collected user's own account data.
	Owners map[string]*OwnerActivity `json:"owners,omitempty"`

	// Warnings records non-fatal collection failures (e.g. data unavailable
	// for the requested user, or API limits reached).
	Warnings []string `json:"warnings,omitempty"`
}

// OwnerActivity groups activity data scoped to a single repository owner
// (a GitHub user or organization).
type OwnerActivity struct {
	Owner         string                         `json:"owner"`
	Repos         []*github.Repository           `json:"repos,omitempty"`
	Events        []*github.Event                `json:"events,omitempty"`
	Contributions *extgh.ContributionsCollection `json:"contributions,omitempty"`
	Pulls         []*github.Issue                `json:"pulls,omitempty"`
	Issues        []*github.Issue                `json:"issues,omitempty"`
	Reviews       []*github.Issue                `json:"reviews,omitempty"`
	Comments      []Comment                      `json:"comments,omitempty"`
	Packages      []*github.Package              `json:"packages,omitempty"`
	Projects      []extgh.ProjectV2              `json:"projects,omitempty"`
	Discussions   []extgh.Discussion             `json:"discussions,omitempty"`
}

// Comment is a single issue or pull request comment authored by the collected user.
type Comment struct {
	Repository string    `json:"repository"`
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

// ownerOf returns the owner login of a repository, or fallback if repo is nil.
func ownerOf(repo *github.Repository, fallback string) string {
	if repo == nil || repo.Owner == nil || repo.Owner.Login == nil {
		return fallback
	}
	return repo.GetOwner().GetLogin()
}

// Data returns the raw collected records for kind at the root scope, as
// returned by the GitHub API, or nil when kind is owner-scoped.
func (r *Result) Data(kind Kind) any {
	switch kind {
	case KindProfile:
		return r.Profile
	case KindFollowers:
		return r.Followers
	case KindFollowing:
		return r.Following
	case KindOrgs:
		return r.Orgs
	case KindGists:
		return r.Gists
	case KindStarred:
		return r.Starred
	case KindWatching:
		return r.Watching
	case KindNotifications:
		return r.Notifications
	default:
		return nil
	}
}

// Data returns the raw collected records for kind scoped to this owner, as
// returned by the GitHub API, or nil when kind is not owner-scoped.
func (oa *OwnerActivity) Data(kind Kind) any {
	switch kind {
	case KindRepos:
		return oa.Repos
	case KindEvents:
		return oa.Events
	case KindContributions:
		return oa.Contributions
	case KindPulls:
		return oa.Pulls
	case KindIssues:
		return oa.Issues
	case KindReviews:
		return oa.Reviews
	case KindComments:
		return oa.Comments
	case KindPackages:
		return oa.Packages
	case KindProjects:
		return oa.Projects
	case KindDiscussions:
		return oa.Discussions
	default:
		return nil
	}
}

// owner returns (creating if necessary) the OwnerActivity bucket for login.
func (r *Result) owner(login string) *OwnerActivity {
	if r.Owners == nil {
		r.Owners = make(map[string]*OwnerActivity)
	}
	oa, ok := r.Owners[login]
	if !ok {
		oa = &OwnerActivity{Owner: login}
		r.Owners[login] = oa
	}
	return oa
}

// warn records a non-fatal collection failure and logs it.
func (r *Result) warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	r.Warnings = append(r.Warnings, msg)
	logger.Warn(msg)
}
