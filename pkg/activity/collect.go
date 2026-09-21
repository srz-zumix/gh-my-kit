package activity

import (
	"context"
	"fmt"

	"github.com/google/go-github/v90/github"
	extgh "github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// dateLayout is the format used in GitHub search qualifiers such as "created:".
const dateLayout = "2006-01-02"

// Collect gathers activity data for opts.Username within [opts.Since, opts.Until]
// using client g. Data that is unavailable for the requested user (e.g. private
// events, notifications, or watched repositories when opts.Self is false) is
// skipped with a warning instead of failing the whole collection.
func Collect(ctx context.Context, g *extgh.GitHubClient, opts Options) (*Result, error) {
	res := &Result{
		Username: opts.Username,
		Since:    opts.Since,
		Until:    opts.Until,
	}
	logger.Info("collecting activity", "user", opts.Username, "since", opts.Since.Format(dateLayout), "until", opts.Until.Format(dateLayout), "kinds", len(opts.Kinds))

	if hasKind(opts.Kinds, KindProfile) {
		logger.Info("collecting profile", "user", opts.Username)
		profile, err := extgh.FindUser(ctx, g, opts.Username)
		if err != nil {
			res.warn("failed to get profile for '%s': %v", opts.Username, err)
		} else {
			res.Profile = profile
		}
	}

	if hasKind(opts.Kinds, KindFollowers) {
		// GitHub does not expose when a follow relationship was created, so this
		// is always the current follower list, not limited to the requested period.
		logger.Info("collecting followers", "user", opts.Username)
		followers, err := extgh.ListUserFollowers(ctx, g, opts.Username)
		if err != nil {
			res.warn("failed to list followers for '%s': %v", opts.Username, err)
		} else {
			res.Followers = followers
			logger.Info("collected followers", "count", len(followers))
		}
	}

	if hasKind(opts.Kinds, KindFollowing) {
		// GitHub does not expose when a follow relationship was created, so this
		// is always the current following list, not limited to the requested period.
		logger.Info("collecting following", "user", opts.Username)
		following, err := extgh.ListUserFollowing(ctx, g, opts.Username)
		if err != nil {
			res.warn("failed to list following for '%s': %v", opts.Username, err)
		} else {
			res.Following = following
			logger.Info("collected following", "count", len(following))
		}
	}

	if hasKind(opts.Kinds, KindOrgs) {
		// GitHub does not expose when an organization membership was created, so
		// this is always the current membership list, not limited to the requested period.
		logger.Info("collecting organizations", "user", opts.Username)
		orgs, err := extgh.ListUserOrganizations(ctx, g, opts.Username)
		if err != nil {
			res.warn("failed to list organizations for '%s': %v", opts.Username, err)
		} else {
			res.Orgs = orgs
			logger.Info("collected organizations", "count", len(orgs))
		}
	}

	if hasKind(opts.Kinds, KindGists) && opts.Self {
		logger.Info("collecting gists", "user", opts.Username)
		gists, err := extgh.ListGists(ctx, g)
		if err != nil {
			res.warn("failed to list gists for '%s': %v", opts.Username, err)
		} else {
			for _, gist := range gists {
				createdAt := gist.GetCreatedAt().Time
				if createdAt.Before(opts.Since) || createdAt.After(opts.Until) {
					continue
				}
				res.Gists = append(res.Gists, gist)
			}
			logger.Info("collected gists", "count", len(res.Gists))
		}
	}

	if hasKind(opts.Kinds, KindStarred) {
		logger.Info("collecting starred repositories", "user", opts.Username)
		starred, err := extgh.ListUserStarredRepositories(ctx, g, opts.Username)
		if err != nil {
			res.warn("failed to list starred repositories for '%s': %v", opts.Username, err)
		} else {
			for _, s := range starred {
				starredAt := s.GetStarredAt().Time
				if starredAt.Before(opts.Since) || starredAt.After(opts.Until) {
					continue
				}
				res.Starred = append(res.Starred, s)
			}
			logger.Info("collected starred repositories", "count", len(res.Starred))
		}
	}

	if hasKind(opts.Kinds, KindWatching) {
		// GitHub does not expose when a repository was watched, so this is
		// always the current watch list, not limited to the requested period.
		if opts.Self {
			logger.Info("collecting watched repositories", "user", opts.Username)
			watching, err := extgh.ListUserWatchedRepositories(ctx, g, opts.Username)
			if err != nil {
				res.warn("failed to list watched repositories for '%s': %v", opts.Username, err)
			} else {
				res.Watching = watching
				logger.Info("collected watched repositories", "count", len(watching))
			}
		} else {
			res.warn("skipped watched repositories: only available for the authenticated user")
		}
	}

	if hasKind(opts.Kinds, KindNotifications) {
		if opts.Self {
			logger.Info("collecting notifications", "user", opts.Username)
			notifications, err := extgh.ListUserNotifications(ctx, g, &opts.Since)
			if err != nil {
				res.warn("failed to list notifications for '%s': %v", opts.Username, err)
			} else {
				res.Notifications = notifications
				logger.Info("collected notifications", "count", len(notifications))
			}
		} else {
			res.warn("skipped notifications: only available for the authenticated user")
		}
	}

	if hasKind(opts.Kinds, KindRepos) {
		logger.Info("collecting repositories", "user", opts.Username)
		repos, err := g.ListUserRepositories(ctx, opts.Username, "owner")
		if err != nil {
			res.warn("failed to list repositories for '%s': %v", opts.Username, err)
		} else {
			for _, repo := range repos {
				owner := ownerOf(repo, opts.Username)
				oa := res.owner(owner)
				oa.Repos = append(oa.Repos, repo)
			}
			logger.Info("collected repositories", "count", len(repos))
		}
	}

	if hasKind(opts.Kinds, KindEvents) {
		logger.Info("collecting events", "user", opts.Username)
		events, err := extgh.ListUserEvents(ctx, g, opts.Username, !opts.Self)
		if err != nil {
			res.warn("failed to list events for '%s': %v", opts.Username, err)
		} else {
			logger.Info("collected events", "count", len(events))
			for _, event := range events {
				createdAt := event.GetCreatedAt().Time
				if createdAt.Before(opts.Since) || createdAt.After(opts.Until) {
					continue
				}
				owner := opts.Username
				if repoName := event.GetRepo().GetName(); repoName != "" {
					if idx := indexOfSlash(repoName); idx >= 0 {
						owner = repoName[:idx]
					}
				}
				oa := res.owner(owner)
				oa.Events = append(oa.Events, event)
			}
		}
	}

	if hasKind(opts.Kinds, KindContributions) {
		logger.Info("collecting contributions", "user", opts.Username)
		contributions, err := extgh.GetUserContributions(ctx, g, opts.Username, opts.Since, opts.Until)
		if err != nil {
			res.warn("failed to get contributions for '%s': %v", opts.Username, err)
		} else if contributions != nil {
			logger.Info("collected contributions", "total_commits", contributions.TotalCommitContributions)
			for _, repoContrib := range contributions.CommitContributionsByRepository {
				owner := opts.Username
				if idx := indexOfSlash(repoContrib.NameWithOwner); idx >= 0 {
					owner = repoContrib.NameWithOwner[:idx]
				}
				oa := res.owner(owner)
				if oa.Contributions == nil {
					oa.Contributions = &extgh.ContributionsCollection{}
				}
				oa.Contributions.CommitContributionsByRepository = append(oa.Contributions.CommitContributionsByRepository, repoContrib)
				oa.Contributions.TotalCommitContributions += repoContrib.Contributions
			}
		}
	}

	dateRange := fmt.Sprintf("%s..%s", opts.Since.Format(dateLayout), opts.Until.Format(dateLayout))

	if hasKind(opts.Kinds, KindPulls) {
		logger.Info("collecting pull requests", "user", opts.Username)
		query := fmt.Sprintf("is:pr author:%s created:%s", opts.Username, dateRange)
		pulls, err := g.SearchIssues(ctx, query)
		if err != nil {
			res.warn("failed to search pull requests authored by '%s': %v", opts.Username, err)
		} else {
			logger.Info("collected pull requests", "count", len(pulls))
			for _, pr := range pulls {
				owner := ownerFromRepositoryURL(pr.GetRepositoryURL(), opts.Username)
				oa := res.owner(owner)
				oa.Pulls = append(oa.Pulls, pr)
			}
		}
	}

	if hasKind(opts.Kinds, KindIssues) {
		logger.Info("collecting issues", "user", opts.Username)
		query := fmt.Sprintf("is:issue author:%s created:%s", opts.Username, dateRange)
		issues, err := g.SearchIssues(ctx, query)
		if err != nil {
			res.warn("failed to search issues authored by '%s': %v", opts.Username, err)
		} else {
			logger.Info("collected issues", "count", len(issues))
			for _, issue := range issues {
				owner := ownerFromRepositoryURL(issue.GetRepositoryURL(), opts.Username)
				oa := res.owner(owner)
				oa.Issues = append(oa.Issues, issue)
			}
		}
	}

	if hasKind(opts.Kinds, KindReviews) {
		logger.Info("collecting reviews", "user", opts.Username)
		query := fmt.Sprintf("is:pr reviewed-by:%s updated:%s", opts.Username, dateRange)
		reviews, err := g.SearchIssues(ctx, query)
		if err != nil {
			res.warn("failed to search pull requests reviewed by '%s': %v", opts.Username, err)
		} else {
			logger.Info("collected reviews", "count", len(reviews))
			for _, pr := range reviews {
				owner := ownerFromRepositoryURL(pr.GetRepositoryURL(), opts.Username)
				oa := res.owner(owner)
				oa.Reviews = append(oa.Reviews, pr)
			}
		}
	}

	if hasKind(opts.Kinds, KindComments) {
		logger.Info("collecting comments", "user", opts.Username)
		query := fmt.Sprintf("commenter:%s updated:%s", opts.Username, dateRange)
		commented, err := g.SearchIssues(ctx, query)
		if err != nil {
			res.warn("failed to search issues commented on by '%s': %v", opts.Username, err)
		} else {
			collectComments(ctx, g, res, opts, commented)
		}
	}

	if hasKind(opts.Kinds, KindPackages) {
		logger.Info("collecting packages", "user", opts.Username)
		packages, err := extgh.ListUserPackages(ctx, g, opts.Username, "", "")
		if err != nil {
			res.warn("failed to list packages for '%s': %v", opts.Username, err)
		} else {
			for _, pkg := range packages {
				owner := opts.Username
				if pkg.GetOwner() != nil && pkg.GetOwner().GetLogin() != "" {
					owner = pkg.GetOwner().GetLogin()
				}
				oa := res.owner(owner)
				oa.Packages = append(oa.Packages, pkg)
			}
			logger.Info("collected packages", "count", len(packages))
		}
	}

	if hasKind(opts.Kinds, KindProjects) {
		logger.Info("collecting projects", "user", opts.Username)
		owners := projectOwners(res, opts)
		projectCount := 0
		for _, owner := range owners {
			projects, err := extgh.ListProjectsV2(ctx, g, owner)
			if err != nil {
				res.warn("failed to list projects for '%s': %v", owner, err)
				continue
			}
			if len(projects) == 0 {
				continue
			}
			oa := res.owner(owner)
			oa.Projects = append(oa.Projects, projects...)
			projectCount += len(projects)
		}
		logger.Info("collected projects", "count", projectCount)
	}

	if hasKind(opts.Kinds, KindDiscussions) {
		logger.Info("collecting discussions", "user", opts.Username)
		discussions, err := extgh.SearchUserDiscussions(ctx, g, opts.Username, "")
		if err != nil {
			res.warn("failed to search discussions for '%s': %v", opts.Username, err)
		} else {
			logger.Info("collected discussions", "count", len(discussions))
			for _, d := range discussions {
				owner := opts.Username
				if idx := indexOfSlash(d.Repository.NameWithOwner); idx >= 0 {
					owner = d.Repository.NameWithOwner[:idx]
				}
				oa := res.owner(owner)
				oa.Discussions = append(oa.Discussions, d)
			}
		}
	}

	logger.Info("finished collecting activity", "user", opts.Username, "owners", len(res.Owners), "warnings", len(res.Warnings))
	return res, nil
}

// maxCommentIssues bounds the number of issues/PRs inspected for comment
// collection, since fetching comments requires one API call per issue.
const maxCommentIssues = 50

// collectComments fetches the comments authored by opts.Username within the
// requested period, for a limited number of issues/PRs found via search.
func collectComments(ctx context.Context, g *extgh.GitHubClient, res *Result, opts Options, issues []*github.Issue) {
	logger.Info("collecting comments", "user", opts.Username, "matched_issues", len(issues))
	total := 0
	for i, issue := range issues {
		if i >= maxCommentIssues {
			res.warn("comment collection limited to the first %d matched issues/pull requests", maxCommentIssues)
			break
		}
		owner, name := ownerRepoFromRepositoryURL(issue.GetRepositoryURL())
		if owner == "" || name == "" {
			continue
		}
		logger.Debug("fetching issue comments", "repository", owner+"/"+name, "number", issue.GetNumber())
		comments, err := g.ListIssueComments(ctx, owner, name, issue.GetNumber())
		if err != nil {
			res.warn("failed to list comments for '%s/%s#%d': %v", owner, name, issue.GetNumber(), err)
			continue
		}
		oa := res.owner(owner)
		for _, c := range comments {
			if c.GetUser().GetLogin() != opts.Username {
				continue
			}
			createdAt := c.GetCreatedAt().Time
			if createdAt.Before(opts.Since) || createdAt.After(opts.Until) {
				continue
			}
			oa.Comments = append(oa.Comments, Comment{
				Repository: owner + "/" + name,
				Number:     issue.GetNumber(),
				Title:      issue.GetTitle(),
				URL:        c.GetHTMLURL(),
				Body:       c.GetBody(),
				CreatedAt:  createdAt,
			})
			total++
		}
	}
	logger.Info("collected comments", "count", total)
}

// projectOwners returns the set of owners (the user plus any organizations
// they belong to) to query for ProjectV2s.
func projectOwners(res *Result, opts Options) []string {
	owners := []string{opts.Username}
	for _, org := range res.Orgs {
		if login := org.GetLogin(); login != "" {
			owners = append(owners, login)
		}
	}
	return owners
}

// ownerFromRepositoryURL extracts the owner login from a GitHub API
// repository URL of the form "https://api.github.com/repos/{owner}/{repo}".
// fallback is returned if the URL cannot be parsed.
func ownerFromRepositoryURL(repositoryURL, fallback string) string {
	owner, _ := ownerRepoFromRepositoryURL(repositoryURL)
	if owner == "" {
		return fallback
	}
	return owner
}

// ownerRepoFromRepositoryURL splits a GitHub API repository URL into its
// owner and repo name components.
func ownerRepoFromRepositoryURL(repositoryURL string) (owner, name string) {
	const marker = "/repos/"
	idx := indexOf(repositoryURL, marker)
	if idx < 0 {
		return "", ""
	}
	rest := repositoryURL[idx+len(marker):]
	slash := indexOfSlash(rest)
	if slash < 0 {
		return "", ""
	}
	return rest[:slash], rest[slash+1:]
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func indexOfSlash(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
