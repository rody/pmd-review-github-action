package pmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Report struct {
	FormatVersion int8   `json:"formatVersion"`
	PMDVersion    string `json:"pmdVersion"`
	Timestamp     string `json:"timestamp"`
	Files         []File `json:"files"`
}

type File struct {
	Filename   string      `json:"filename"`
	Violations []Violation `json:"violations"`
}

type Violation struct {
	FileName        string
	BeginLine       int    `json:"beginLine"`
	BeginColumn     int    `json:"beginColumn"`
	EndLine         int    `json:"endLine"`
	EndColumn       int    `json:"endColumn"`
	Description     string `json:"description"`
	Rule            string `json:"rule"`
	Ruleset         string `json:"ruleset"`
	Priority        int    `json:"priority"`
	ExternalInfoUrl string `json:"externalInfoUrl"`
}

type LineViolations map[int][]Violation

// Parse reads a PMD report in JSON format
// and returns a Report.
func Parse(pmdstring io.Reader, dir string) (map[string]LineViolations, error) {
	report := Report{}
	err := json.NewDecoder(pmdstring).Decode(&report)
	if err != nil {
		return nil, err
	}

	v := make(map[string]LineViolations)

	for _, f := range report.Files {
		fn, err := relPath(f.Filename, dir)
		if err != nil {
			return nil, err
		}

		lvs := make(LineViolations)

		for _, v := range f.Violations {
			v.FileName = fn
			lvs[v.BeginLine] = append(lvs[v.BeginLine], v)
		}

		v[fn] = lvs
	}

	return v, nil
}

func relPath(file, dir string) (string, error) {
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
