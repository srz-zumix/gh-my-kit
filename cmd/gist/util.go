/*
Copyright © 2025 srz_zumix
*/
package gist

import (
	"context"
	"fmt"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

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
