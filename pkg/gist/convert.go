/*
Copyright © 2025 srz_zumix
*/
package gist

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/cli/v2/git"
	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// ConvertGistToRepoOptions holds options for ConvertGistToRepo.
type ConvertGistToRepoOptions struct {
	// RepoName is the target repository name. If empty, it is derived from the
	// gist description; if that is also empty, the gist ID is used.
	RepoName string
	// OrgName is the organization to create the repository under.
	// If empty, the repository is created under the authenticated user.
	OrgName string
	// Visibility controls the visibility of the created repository.
	// Accepted values: "public", "private", "internal".
	// If empty, the visibility is inherited from the gist (public/private).
	Visibility string
	// RenameMasterToMain renames the default branch from "master" to "main"
	// after pushing. If the default branch is not "master", this is a no-op.
	RenameMasterToMain bool
}

// sanitizeRepoName converts an arbitrary string into a valid GitHub repository
// name by replacing disallowed characters with hyphens and stripping leading /
// trailing dots and hyphens.
func sanitizeRepoName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, s)
	s = strings.Trim(s, ".-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

// ConvertGistToRepo converts a gist to a regular GitHub repository, preserving
// the full git history via git clone --mirror + git push --mirror.
// src is used to fetch the gist; dst is used to create the repository.
// If dst is nil, src is used for both.
func ConvertGistToRepo(ctx context.Context, src, dst *gh.GitHubClient, gistID string, opts ConvertGistToRepoOptions) (*github.Repository, error) {
	if dst == nil {
		dst = src
	}

	gistObj, err := src.GetGist(ctx, gistID)
	if err != nil {
		return nil, fmt.Errorf("get gist: %w", err)
	}

	srcURL := gistObj.GetGitPullURL()
	if srcURL == "" {
		return nil, fmt.Errorf("gist has no git pull URL")
	}

	repoName := opts.RepoName
	if repoName == "" {
		repoName = sanitizeRepoName(gistObj.GetDescription())
	}
	if repoName == "" {
		repoName = gistID
	}

	private := !gistObj.GetPublic()
	visibility := opts.Visibility
	switch visibility {
	case "":
		// inherit from gist
	case "public":
		private = false
	case "private":
		private = true
	case "internal":
		private = true
	default:
		return nil, fmt.Errorf("invalid visibility %q: must be one of public, private, internal", visibility)
	}

	tmpDir, err := os.MkdirTemp("", "gh-my-kit-convert-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			logger.Warn("failed to remove temp dir", "path", tmpDir, "error", removeErr)
		}
	}()

	mirrorDir := filepath.Join(tmpDir, "repo.git")
	cloneClient := &git.Client{Stderr: os.Stderr, Stdout: os.Stderr}
	cloneCmd, err := cloneClient.Command(ctx, "clone", "--mirror", srcURL, mirrorDir)
	if err != nil {
		return nil, fmt.Errorf("prepare git clone --mirror: %w", err)
	}
	cloneCmd.Env = gh.GitCmdEnv(src, srcURL)
	if err := cloneCmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone --mirror: %w", err)
	}

	// Rename master → main in the local mirror before pushing so that GitHub
	// picks up main as the default branch without any post-push API call.
	if opts.RenameMasterToMain {
		if headBytes, readErr := os.ReadFile(filepath.Join(mirrorDir, "HEAD")); readErr == nil {
			if strings.TrimSpace(string(headBytes)) == "ref: refs/heads/master" {
				mirrorClient := &git.Client{RepoDir: mirrorDir, Stderr: os.Stderr, Stdout: os.Stderr}
				renameCmd, err := mirrorClient.Command(ctx, "branch", "-m", "master", "main")
				if err != nil {
					return nil, fmt.Errorf("prepare git branch -m: %w", err)
				}
				if err := renameCmd.Run(); err != nil {
					return nil, fmt.Errorf("git branch -m master main: %w", err)
				}
			}
		}
	}

	newRepo := &github.Repository{
		Name:     github.Ptr(repoName),
		Private:  github.Ptr(private),
		AutoInit: github.Ptr(false),
	}
	if visibility == "internal" {
		newRepo.Visibility = github.Ptr("internal")
	}
	if desc := gistObj.GetDescription(); desc != "" {
		newRepo.Description = github.Ptr(desc)
	}

	createdRepo, err := dst.CreateRepository(ctx, opts.OrgName, newRepo)
	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}

	dstURL := createdRepo.GetCloneURL()
	if dstURL == "" {
		return nil, fmt.Errorf("created repository has no clone URL")
	}

	pushClient := &git.Client{RepoDir: mirrorDir, Stderr: os.Stderr, Stdout: os.Stderr}
	pushCmd, err := pushClient.Command(ctx, "push", "--mirror", dstURL)
	if err != nil {
		return nil, fmt.Errorf("prepare git push --mirror: %w", err)
	}
	pushCmd.Env = gh.GitCmdEnv(dst, dstURL)
	if err := pushCmd.Run(); err != nil {
		return nil, fmt.Errorf("git push --mirror: %w", err)
	}

	return createdRepo, nil
}
