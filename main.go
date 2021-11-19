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

