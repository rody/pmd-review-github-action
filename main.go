package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

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

	changed := getChanged(diff)
	githubactions.Debugf("diff %+v", changed)

	report, err := parseReport(reportfile)
	if err != nil {
		githubactions.Fatalf("could not parse reportfile: %s", err)
	}

	tr, err := GetViolationsForDiff(report, changed)
	if err != nil {
		githubactions.Fatalf("could not get violations: %s", err)
	}
	githubactions.Debugf("violations %+v", tr)
}

func getChanged(diff *diffparser.Diff) LineChanges {
	fv := make(LineChanges)

	for _, f := range diff.Files {
		if f.Mode == diffparser.DELETED {
			continue
		}

		for _, h := range f.Hunks {
			for _, dl := range h.NewRange.Lines {
				if dl.Mode == diffparser.ADDED {
					dlv := fv[f.NewName]
					dlv[dl.Number] = dl
				}
			}
		}
	}
	return fv
}

type LineChanges map[string]map[int]*diffparser.DiffLine

func GetViolationsForDiff(r pmd.Report, lc LineChanges) ([]toReport, error) {
	var tr []toReport
	for _, f := range r.Files {
		fn, err := relpath(f.Filename)
		if err != nil {
			return tr, err
		}

		lines, exists := lc[fn]
		if !exists {
			// file not present in diff
			continue
		}

		for _, v := range f.Violations {
			line, exists := lines[v.BeginLine]
			if !exists {
				continue
			}

			tr = append(tr, toReport{
				Violation: v,
				Line:      line,
			})

		}
	}

	return tr, nil
}

type toReport struct {
	Violation pmd.Violation
	Line      *diffparser.DiffLine
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

type LineViolations map[int][]pmd.Violations
type FileViolations map[string]LineViolations
