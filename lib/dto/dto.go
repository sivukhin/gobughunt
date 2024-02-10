package dto

import (
	"fmt"
	"time"
)

type Linter struct {
	Meta     LinterMeta
	Instance *LinterInstance
}

type LinterMeta struct {
	LinterId        string
	LinterGitUrl    string
	LinterGitBranch string
}

type LinterInstance struct {
	LinterId           string
	DockerImage        string
	DockerImageShaHash string
}

func (linter LinterInstance) String() string {
	return fmt.Sprintf("[%v](%v@sha256:%v)", linter.LinterId, linter.DockerImage, linter.DockerImageShaHash)
}

type Repo struct {
	Meta     RepoMeta
	Instance *RepoInstance
}

type RepoMeta struct {
	RepoId        string
	RepoGitUrl    string
	RepoGitBranch string
}

type RepoInstance struct {
	RepoId        string
	GitUrl        string
	GitCommitHash string
}

func (repo RepoInstance) String() string {
	return fmt.Sprintf("[%v](%v, %v)", repo.RepoId, repo.GitUrl, repo.GitCommitHash)
}

type LintStatus string

const (
	Pending LintStatus = "pending"
	Failed             = "failed"
	Succeed            = "succeed"
	Skipped            = "skipped"
)

type LintTask struct {
	LintId string
	Linter LinterInstance
	Repo   RepoInstance
}

type LintResult struct {
	LintStatus        LintStatus
	LintStatusComment string
	LintDuration      time.Duration
	Highlights        []LintHighlightSnippet
}

type HighlightSnippet struct {
	StartLine int
	EndLine   int
	Code      string
}

type LintHighlightSnippet struct {
	LintHighlight
	Snippet HighlightSnippet
}

type LintHighlight struct {
	Path        string
	StartLine   int
	EndLine     int
	Explanation string
}

type GitRef struct {
	Branch     string
	CommitHash string
}

func (ref GitRef) String() string {
	if ref.Branch != "" {
		return "branch:" + ref.Branch
	} else {
		return "commit:" + ref.CommitHash
	}
}
