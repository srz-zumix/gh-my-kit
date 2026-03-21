/*
Copyright © 2025 srz_zumix
*/
package gist

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// newClientForHost creates a GitHubClient for the given host and optional token.
// If host is empty, the current authenticated host is used.
func newClientForHost(host, token string) (*gh.GitHubClient, error) {
	repo := repository.Repository{Host: host}
	if token != "" {
		return gh.NewGitHubClientWithToken(repo, token)
	}
	return gh.NewGitHubClientWithRepo(repo)
}

// newClientPair creates source and destination clients and validates that they
// do not refer to the same user on the same host.
func newClientPair(ctx context.Context, src, dst, srcToken, dstToken string) (srcClient, dstClient *gh.GitHubClient, err error) {
	srcClient, err = newClientForHost(src, srcToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create source client: %w", err)
	}

	dstClient, err = newClientForHost(dst, dstToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create destination client: %w", err)
	}

	srcUser, err := gh.GetLoginUser(ctx, srcClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get source user: %w", err)
	}
	dstUser, err := gh.GetLoginUser(ctx, dstClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get destination user: %w", err)
	}
	if srcClient.Host() == dstClient.Host() && srcUser.GetLogin() == dstUser.GetLogin() {
		return nil, nil, fmt.Errorf("source and destination user must be different (%s@%s)", srcUser.GetLogin(), srcClient.Host())
	}

	return srcClient, dstClient, nil
}

// resolveGistIDs returns the given IDs if non-empty, or lists all gists from
// srcClient and returns their IDs.
func resolveGistIDs(ctx context.Context, srcClient *gh.GitHubClient, args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}
	gists, err := gh.ListGists(ctx, srcClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list gists: %w", err)
	}
	ids := make([]string, 0, len(gists))
	for _, g := range gists {
		ids = append(ids, g.GetID())
	}
	return ids, nil
}
