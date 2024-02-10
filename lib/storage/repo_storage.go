package storage

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sivukhin/gobughunt/lib/dto"
)

type RepoList []dto.Repo

func (l RepoList) SelectWithInstances() RepoList {
	repos := make(RepoList, 0)
	for _, repo := range l {
		if repo.Instance != nil {
			repos = append(repos, repo)
		}
	}
	return repos
}

type RepoStorage interface {
	AddOrUpdate(ctx context.Context, repo dto.Repo, updatedAt time.Time) error
	Get(ctx context.Context, repoId string) (dto.Repo, error)
	Delete(ctx context.Context, repoId string) error
	List(ctx context.Context) (RepoList, error)
}

type PgRepoStorage PgStorage

var _ RepoStorage = PgRepoStorage{}

//go:embed queries/repos.sql
var repoTableSql string

//go:embed queries/repos_add_or_update.sql
var repoAddOrUpdate string

//go:embed queries/repos_get.sql
var repoGetSql string

//go:embed queries/repos_list.sql
var repoListSql string

//go:embed queries/repos_delete.sql
var repoDeleteSql string

var NoRepoErr = errors.New("repo not found")

func (p PgRepoStorage) InitTables(ctx context.Context) error {
	_, repoTableErr := p.Exec(ctx, repoTableSql)
	return repoTableErr
}

func (p PgRepoStorage) AddOrUpdate(ctx context.Context, repo dto.Repo, updatedAt time.Time) error {
	var gitCommitHash *string
	if repo.Instance != nil {
		gitCommitHash = &repo.Instance.GitCommitHash
	}
	_, err := p.Exec(
		ctx,
		repoAddOrUpdate,
		repo.Meta.RepoId,
		repo.Meta.RepoGitUrl,
		repo.Meta.RepoGitBranch,
		gitCommitHash,
		updatedAt,
	)
	return err
}

func (p PgRepoStorage) Get(ctx context.Context, repoId string) (dto.Repo, error) {
	rows, err := p.Query(ctx, repoGetSql, repoId)
	if err != nil {
		return dto.Repo{}, err
	}
	for rows.Next() {
		return scanRepoRow(rows)
	}
	return dto.Repo{}, NoRepoErr
}

func (p PgRepoStorage) List(ctx context.Context) (RepoList, error) {
	rows, err := p.Query(ctx, repoListSql)
	if err != nil {
		return nil, err
	}
	repos := make(RepoList, 0)
	for rows.Next() {
		repo, err := scanRepoRow(rows)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func (p PgRepoStorage) Delete(ctx context.Context, repoId string) error {
	_, err := p.Exec(ctx, repoDeleteSql, repoId)
	return err
}

func scanRepoRow(rows pgx.Row) (dto.Repo, error) {
	var (
		repoId                string
		repoGitUrl            string
		repoGitBranch         string
		repoLastGitCommitHash *string
	)
	err := rows.Scan(&repoId, &repoGitUrl, &repoGitBranch, &repoLastGitCommitHash)
	if err != nil {
		return dto.Repo{}, err
	}
	meta := dto.RepoMeta{
		RepoId:        repoId,
		RepoGitUrl:    repoGitUrl,
		RepoGitBranch: repoGitBranch,
	}
	var instance *dto.RepoInstance
	if repoLastGitCommitHash != nil {
		instance = &dto.RepoInstance{
			RepoId:        repoId,
			GitUrl:        repoGitUrl,
			GitCommitHash: *repoLastGitCommitHash,
		}
	}
	return dto.Repo{Meta: meta, Instance: instance}, nil
}