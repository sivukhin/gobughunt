package main

type DashboardDto struct {
	Login   string
	Linters []LinterDto
	Repos   []RepoDto
}

type LintHighlightsDto struct {
	Login      string
	LintId     string
	RepoId     string
	LinterId   string
	Highlights []LintHighlightDto
}

type StatDto struct {
	TotalHighlight    int
	PendingHighlight  int
	RejectedHighlight int
	AcceptedHighlight int
}

type LinterDto struct {
	Id                 string
	GitUrl             string
	GitBranch          string
	DockerImage        *string
	DockerImageShaHash *string
	*StatDto
}

type RepoDto struct {
	Id            string
	GitUrl        string
	GitBranch     string
	GitCommitHash *string
	*StatDto
}

type LintTaskDto struct {
	Id              string
	Status          string
	StatusComment   *string
	LintDurationSec *float64
	Linter          LinterDto
	Repo            RepoDto
}

type HighlightSnippetDto struct {
	StartLine int
	EndLine   int
	Code      string
}

type LintHighlightDto struct {
	LintId      string
	Linter      LinterDto
	Repo        RepoDto
	Status      string
	Path        string
	StartLine   int
	EndLine     int
	Explanation string
	Snippet     HighlightSnippetDto
}
