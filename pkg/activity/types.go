// Package activity collects and renders a GitHub user's activity across the
// data GitHub exposes via its REST, Search and GraphQL APIs.
package activity

import "time"

// Kind identifies a category of collected activity data.
type Kind string

// Supported activity kinds.
const (
	KindProfile       Kind = "profile"
	KindFollowers     Kind = "followers"
	KindFollowing     Kind = "following"
	KindOrgs          Kind = "orgs"
	KindEvents        Kind = "events"
	KindContributions Kind = "contributions"
	KindPulls         Kind = "pulls"
	KindIssues        Kind = "issues"
	KindReviews       Kind = "reviews"
	KindComments      Kind = "comments"
	KindRepos         Kind = "repos"
	KindStarred       Kind = "starred"
	KindWatching      Kind = "watching"
	KindGists         Kind = "gists"
	KindPackages      Kind = "packages"
	KindProjects      Kind = "projects"
	KindDiscussions   Kind = "discussions"
	KindNotifications Kind = "notifications"
)

// AllKinds lists every supported activity kind in a stable, deterministic order.
// This order is also used when rendering output.
var AllKinds = []Kind{
	KindProfile, KindFollowers, KindFollowing, KindOrgs,
	KindEvents, KindContributions,
	KindPulls, KindIssues, KindReviews, KindComments,
	KindRepos, KindStarred, KindWatching,
	KindGists, KindPackages, KindProjects, KindDiscussions, KindNotifications,
}

// Mode selects how much detail is rendered for each collected kind.
type Mode string

// Supported render modes.
const (
	// ModeSummary renders aggregate counts and a titled link list per kind.
	ModeSummary Mode = "summary"
	// ModeDetail renders aggregate counts plus every collected record.
	ModeDetail Mode = "detail"
)

// ModeList lists the supported --mode values, used for flag validation.
var ModeList = []string{string(ModeSummary), string(ModeDetail)}

// Options configures a Collect call.
type Options struct {
	// Username is the GitHub login whose activity is collected.
	Username string
	// Self reports whether Username is the authenticated user, enabling
	// access to data that GitHub only exposes for the caller (private events,
	// notifications).
	Self bool
	// Since and Until bound the collected activity window.
	Since time.Time
	Until time.Time
	// Kinds is the set of activity kinds to collect.
	Kinds []Kind
}

// hasKind reports whether kinds contains k.
func hasKind(kinds []Kind, k Kind) bool {
	for _, x := range kinds {
		if x == k {
			return true
		}
	}
	return false
}
