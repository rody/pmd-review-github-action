package main

import (
	"context"
	"strings"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

type GClient struct {
	client *github.Client
	Owner  string
	Repo   string
}

func NewGClient(token, repository string) *GClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GClient{
		client: github.NewClient(tc),
		Owner:  strings.Split(repository, "/")[0],
		Repo:   strings.Split(repository, "/")[1],
	}
}

func (gc *GClient) getDiff(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	pr, _, err := gc.client.PullRequests.Get(ctx, gc.Owner, gc.Repo, prNumber)
	return pr, err
}
