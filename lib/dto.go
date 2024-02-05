package lib

import (
	"time"
)

type ModerationStatus string

const (
	PendingStatus  ModerationStatus = "pending"
	RejectedStatus                  = "rejected"
	AcceptedStatus                  = "accepted"
)

type Moderation struct {
	ModeratorId       string
	ModerationStatus  ModerationStatus
	ModerationTime    time.Time
	ModerationComment string
}

type Prey struct {
	Moderation

	PreyId        string
	Language      string
	GitRepoUrl    string
	GitRepoBranch string
	CreationTime  time.Time
	LastCheckHash string
	LastCheckTime time.Time
}

type Hunter struct {
	Moderation

	HunterId      string
	Language      string
	DockerImage   string
	CreationTime  time.Time
	LastCheckHash string
	LastCheckTime time.Time
}

type HuntStatus string

const (
	Pending HuntStatus = "pending"
	Failed             = "failed"
	Succeed            = "succeed"
	Skipped            = "skipped"
)

type Hunt struct {
	HuntId       string
	HuntTime     time.Time
	PreyId       string
	HunterId     string
	PreyHash     string
	HunterHash   string
	HuntStatus   HuntStatus
	LockTime     time.Time
	CreationTime time.Time
}
