package main

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/google/go-github/v40/github"
	"github.com/waigani/diffparser"
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

func (gc *GClient) getDiff(ctx context.Context, prNumber int) (*diffparser.Diff, error) {
	pr, _, err := gc.client.PullRequests.Get(ctx, gc.Owner, gc.Repo, prNumber)
	if err != nil {
		return nil, err
	}


	req, err := gc.client.NewRequest("GET", pr.GetDiffURL(), nil)
	if err != nil {
		return nil, err
	}

	res, err := gc.client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return diffparser.Parse(string(b))
}
