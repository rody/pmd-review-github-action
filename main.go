package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rody/pmd-review-github-action/pmd"
	"github.com/sethvargo/go-githubactions"
)

var (
	dir              string
	reportfile       string
	githubToken string
)

func main() {
	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&reportfile, "reportfile", "", "")
	flag.StringVar(&githubToken, "github-token", "", "")
	flag.Parse()

	if reportfile == "" {
		githubactions.Fatalf("missing input 'reportfile'")
	}

	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			githubactions.Fatalf("missing github token")
		}
	}

	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	if owner == "" {
		githubactions.Fatalf("missing GITHUB_REPOSITORY_OWNER")
	}

	repo := os.Getenv("GITHUB_REPOSITORY")
	if owner == "" {
		githubactions.Fatalf("missing GITHUB_REPOSITORY")
	}

	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		githubactions.Fatalf("missing GITHUB_SHA")
	}

	gc := NewGClient(githubToken, owner, repo)
	pr, err := gc.getDiff(context.Background(), sha)

	githubactions.Debugf("pr %+v", *pr)

	report, err := parseReport(reportfile)
	if err != nil {
		githubactions.Fatalf("could not parse reportfile: %s", err)
	}

	for f, _ := range report.Files {
		fmt.Println(f)
	}
}

func relpath(file string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	filename := filepath.Join(dir, file)

	if strings.HasPrefix(filename, "/") {
		return filepath.Rel(cwd, filename)
	}

	return filename, nil
}

func parseReport(filename string) (pmd.Report, error) {
	f, err := os.Open(filename)
	if err != nil {
		return pmd.Report{}, err
	}
	defer f.Close()
	return pmd.Parse(f)
}
