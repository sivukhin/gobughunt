package main

type StatDto struct {
	TotalHighlightDedup    int `json:"totalHighlightDedup"`
	PendingHighlightDedup  int `json:"pendingHighlightDedup"`
	RejectedHighlightDedup int `json:"rejectedHighlightDedup"`
	AcceptedHighlightDedup int `json:"acceptedHighlightDedup"`
}

type LinterDto struct {
	Id                 string  `json:"id"`
	GitUrl             string  `json:"gitUrl"`
	GitBranch          string  `json:"gitBranch"`
	DockerImage        *string `json:"dockerImage"`
	DockerImageShaHash *string `json:"dockerImageShaHash"`
	*StatDto
}

type RepoDto struct {
	Id            string  `json:"id"`
	GitUrl        string  `json:"gitUrl"`
	GitBranch     string  `json:"gitBranch"`
	GitCommitHash *string `json:"gitCommitHash"`
	*StatDto
}

type LintTaskDto struct {
	Id              string    `json:"id"`
	Status          string    `json:"status"`
	StatusComment   *string   `json:"statusComment"`
	LintDurationSec *float64  `json:"lintDurationSec"`
	Linter          LinterDto `json:"linter"`
	Repo            RepoDto   `json:"repo"`
}

type HighlightSnippetDto struct {
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Code      string `json:"code"`
}

type LintHighlightDto struct {
	LintId      string              `json:"lintId"`
	Linter      LinterDto           `json:"linter"`
	Repo        RepoDto             `json:"repo"`
	Status      string              `json:"status"`
	Path        string              `json:"path"`
	StartLine   int                 `json:"startLine"`
	EndLine     int                 `json:"endLine"`
	Explanation string              `json:"explanation"`
	Snippet     HighlightSnippetDto `json:"snippet"`
}
