package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v40/github"
	"github.com/sethvargo/go-githubactions"
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

func (gc *GClient) getDiff(ctx context.Context, sha string) (*github.PullRequest, error) {
	prOptions := github.PullRequestListOptions{
		Head:  sha,
		State: "open",
	}

	prs, _, err := gc.client.PullRequests.ListPullRequestsWithCommit(ctx, gc.Owner, gc.Repo, sha, &prOptions)
	if err != nil {
		return nil, err
	}

	if len(prs) == 0 {
		return nil, fmt.Errorf("no open PR found containing sha %s", sha)
	}

	for _, pr := range prs {
		githubactions.Debugf("pr: %v\n", pr)
	}

	return prs[0], nil
}
