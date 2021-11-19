package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/go-github/v40/github"
	"github.com/rody/pmd-review-github-action/pmd"
	"github.com/sethvargo/go-githubactions"
	"github.com/waigani/diffparser"
)

var (
	dir         string
	reportfile  string
	githubToken string
	prNumber    int
)

func main() {
	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&reportfile, "reportfile", "", "")
	flag.StringVar(&githubToken, "github-token", "", "")
	flag.IntVar(&prNumber, "pr-number", 0, "")
	flag.Parse()

	githubactions.Debugf("prNumber: %s", prNumber)

	if reportfile == "" {
		githubactions.Fatalf("missing input 'reportfile'")
	}

	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			githubactions.Fatalf("missing github token")
		}
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		githubactions.Fatalf("missing GITHUB_REPOSITORY")
	}

	if prNumber == 0 {
		githubactions.Fatalf("missing pr-number")
	}

	gc := NewGClient(githubToken, repository)
	diff, err := gc.getDiff(context.Background(), prNumber)
	if err != nil {
		githubactions.Fatalf("%s", err)
	}

	githubactions.Debugf("diff %+v", *diff)

	violations, err := parseReport(reportfile)
	if err != nil {
		githubactions.Fatalf("could not parse reportfile: %s", err)
	}

	comments := getReviewComments(diff, violations)
	githubactions.Debugf("diff %+v", comments)

	if len(comments) == 0 {
		githubactions.Infof("no issue")
		return
	}

	msg := "Some changes are required :D"
	review := github.PullRequestReviewRequest{
		Body: &msg,
		Comments: comments,

	}

	gc.client.PullRequests.CreateReview(context.Background(), gc.Owner, gc.Repo, prNumber, &review)

}

func getReviewComments(diff *diffparser.Diff, violations map[string]pmd.LineViolations) []*github.DraftReviewComment {
	var comments []*github.DraftReviewComment

	for _, f := range diff.Files {
		if f.Mode == diffparser.DELETED {
			continue
		}

		lvs, exists := violations[f.NewName]
		if !exists {
			// no violations for this file
			continue
		}

		for _, h := range f.Hunks {
			for _, dl := range h.NewRange.Lines {
				if dl.Mode != diffparser.ADDED {
					continue
				}

				vs, exists := lvs[dl.Number]
				if !exists {
					// no violations for this line
					continue
				}

				for _, v := range vs {
					comments = append(comments, &github.DraftReviewComment{
						Path: &v.FileName,
						Position: &dl.Position,
						Body: &v.Description,
						Line: &v.BeginLine,
					})
				}
			}
		}
	}
	return comments
}

func parseReport(filename string) (map[string]pmd.LineViolations, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return pmd.Parse(f, dir)
}
