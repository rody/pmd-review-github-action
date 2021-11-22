package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v40/github"
	"github.com/waigani/diffparser"
	"golang.org/x/oauth2"
)

type GClient struct {
	token string
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
		token: token,
		client: github.NewClient(tc),
		Owner:  strings.Split(repository, "/")[0],
		Repo:   strings.Split(repository, "/")[1],
	}
}

func (gc *GClient) getPullRequest(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	pr, resp, err := gc.client.PullRequests.Get(ctx, gc.Owner, gc.Repo, prNumber)
	if err != nil {
		return nil, err
	}

	if pr.GetNumber() == 0 {
		return nil, fmt.Errorf("could not find pull request with number '%d', %+v", prNumber, resp)
	}

	return pr, nil
}

func (gc *GClient) getDiff(ctx context.Context, prNumber int) (*diffparser.Diff, error) {
	diff, _, err := gc.client.PullRequests.GetRaw(ctx, gc.Owner, gc.Repo, prNumber, github.RawOptions{Type: github.Diff})
	if err != nil {
		return nil, err
	}

	return diffparser.Parse(diff)
}
