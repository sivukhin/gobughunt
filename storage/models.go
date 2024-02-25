// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package storage

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type HighlightStatus string

const (
	HighlightStatusPending  HighlightStatus = "pending"
	HighlightStatusAccepted HighlightStatus = "accepted"
	HighlightStatusRejected HighlightStatus = "rejected"
)

func (e *HighlightStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = HighlightStatus(s)
	case string:
		*e = HighlightStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for HighlightStatus: %T", src)
	}
	return nil
}

type NullHighlightStatus struct {
	HighlightStatus HighlightStatus
	Valid           bool // Valid is true if HighlightStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullHighlightStatus) Scan(value interface{}) error {
	if value == nil {
		ns.HighlightStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.HighlightStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullHighlightStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.HighlightStatus), nil
}

type LintStatus string

const (
	LintStatusPending LintStatus = "pending"
	LintStatusLocked  LintStatus = "locked"
	LintStatusSucceed LintStatus = "succeed"
	LintStatusFailed  LintStatus = "failed"
	LintStatusSkipped LintStatus = "skipped"
)

func (e *LintStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = LintStatus(s)
	case string:
		*e = LintStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for LintStatus: %T", src)
	}
	return nil
}

type NullLintStatus struct {
	LintStatus LintStatus
	Valid      bool // Valid is true if LintStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullLintStatus) Scan(value interface{}) error {
	if value == nil {
		ns.LintStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.LintStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullLintStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.LintStatus), nil
}

type LintHighlight struct {
	LintID            string
	Path              string
	StartLine         int32
	EndLine           int32
	Explanation       string
	SnippetStartLine  int32
	SnippetEndLine    int32
	SnippetCode       string
	ModerationStatus  HighlightStatus
	ModerationComment pgtype.Text
	ModeratedAt       pgtype.Timestamp
}

type LintTask struct {
	LintID              string
	LinterID            string
	LinterDockerImage   string
	LinterDockerShaHash string
	RepoID              string
	RepoGitUrl          string
	RepoGitCommitHash   string
	LintStatus          LintStatus
	LintStatusComment   pgtype.Text
	LintDuration        pgtype.Interval
	CreatedAt           pgtype.Timestamp
	LockedAt            pgtype.Timestamp
	LintedAt            pgtype.Timestamp
}

type Linter struct {
	LinterID                string
	LinterGitUrl            string
	LinterGitBranch         string
	LinterLastDockerImage   pgtype.Text
	LinterLastDockerShaHash pgtype.Text
	CreatedAt               pgtype.Timestamp
	UpdatedAt               pgtype.Timestamp
}

type Repo struct {
	RepoID                string
	RepoGitUrl            string
	RepoGitBranch         string
	RepoLastGitCommitHash pgtype.Text
	CreatedAt             pgtype.Timestamp
	UpdatedAt             pgtype.Timestamp
}
