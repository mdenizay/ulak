// Package git provides GitHub API integration and git operations.
package git

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client.
type Client struct {
	gh    *github.Client
	owner string
	repo  string
}

// NewClient creates a GitHub API client from a personal access token.
func NewClient(token, owner, repo string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		gh:    github.NewClient(tc),
		owner: owner,
		repo:  repo,
	}
}

// AddDeployKey adds a read-only deploy key to the repository.
// Returns the created key ID (useful for later deletion).
func (c *Client) AddDeployKey(title, pubKey string) (int64, error) {
	ctx := context.Background()
	key, _, err := c.gh.Repositories.CreateKey(ctx, c.owner, c.repo, &github.Key{
		Title:    github.String(title),
		Key:      github.String(pubKey),
		ReadOnly: github.Bool(true),
	})
	if err != nil {
		return 0, fmt.Errorf("github: add deploy key: %w", err)
	}
	return key.GetID(), nil
}

// DeleteDeployKey removes a deploy key by ID.
func (c *Client) DeleteDeployKey(keyID int64) error {
	ctx := context.Background()
	_, err := c.gh.Repositories.DeleteKey(ctx, c.owner, c.repo, keyID)
	if err != nil {
		return fmt.Errorf("github: delete deploy key: %w", err)
	}
	return nil
}

// GetSSHCloneURL returns the SSH clone URL for the repository.
func (c *Client) GetSSHCloneURL() (string, error) {
	ctx := context.Background()
	repo, _, err := c.gh.Repositories.Get(ctx, c.owner, c.repo)
	if err != nil {
		return "", fmt.Errorf("github: get repo: %w", err)
	}
	return repo.GetSSHURL(), nil
}

// ListBranches returns all branch names for the repository.
func (c *Client) ListBranches() ([]string, error) {
	ctx := context.Background()
	branches, _, err := c.gh.Repositories.ListBranches(ctx, c.owner, c.repo, nil)
	if err != nil {
		return nil, fmt.Errorf("github: list branches: %w", err)
	}
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.GetName()
	}
	return names, nil
}
