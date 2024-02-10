package storage

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sivukhin/gobughunt/lib/dto"
)

type LinterList []dto.Linter

func (l LinterList) SelectWithInstances() LinterList {
	linters := make(LinterList, 0)
	for _, linter := range l {
		if linter.Instance != nil {
			linters = append(linters, linter)
		}
	}
	return linters
}

type LinterStorage interface {
	AddOrUpdate(ctx context.Context, linter dto.Linter, updatedAt time.Time) error
	Get(ctx context.Context, linterId string) (dto.Linter, error)
	Delete(ctx context.Context, linterId string) error
	List(ctx context.Context) (LinterList, error)
}

type PgLinterStorage PgStorage

var _ LinterStorage = PgLinterStorage{}

//go:embed queries/linters.sql
var linterTableSql string

//go:embed queries/linters_add_or_update.sql
var linterAddOrUpdate string

//go:embed queries/linters_get.sql
var linterGetSql string

//go:embed queries/linters_list.sql
var linterListSql string

//go:embed queries/linters_delete.sql
var linterDeleteSql string

var NoLinterErr = errors.New("linter not found")

func (p PgLinterStorage) InitTables(ctx context.Context) error {
	_, linterTableErr := p.Exec(ctx, linterTableSql)
	return linterTableErr
}

func (p PgLinterStorage) AddOrUpdate(ctx context.Context, linter dto.Linter, updatedAt time.Time) error {
	var dockerImage *string
	if linter.Instance != nil {
		dockerImage = &linter.Instance.DockerImage
	}
	var dockerImageShaHash *string
	if linter.Instance != nil {
		dockerImageShaHash = &linter.Instance.DockerImageShaHash
	}
	_, err := p.Exec(
		ctx,
		linterAddOrUpdate,
		linter.Meta.LinterId,
		linter.Meta.LinterGitUrl,
		linter.Meta.LinterGitBranch,
		dockerImage,
		dockerImageShaHash,
		updatedAt,
	)
	return err
}

func (p PgLinterStorage) Get(ctx context.Context, linterId string) (dto.Linter, error) {
	rows, err := p.Query(ctx, linterGetSql, linterId)
	if err != nil {
		return dto.Linter{}, err
	}
	defer rows.Close()
	for rows.Next() {
		return scanLinterRow(rows)
	}
	if rows.Err() != nil {
		return dto.Linter{}, rows.Err()
	}
	return dto.Linter{}, NoLinterErr
}

func (p PgLinterStorage) List(ctx context.Context) (LinterList, error) {
	rows, err := p.Query(ctx, linterListSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	linters := make(LinterList, 0)
	for rows.Next() {
		linter, err := scanLinterRow(rows)
		if err != nil {
			return nil, err
		}
		linters = append(linters, linter)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return linters, nil
}

func (p PgLinterStorage) Delete(ctx context.Context, linterId string) error {
	_, err := p.Exec(ctx, linterDeleteSql, linterId)
	return err
}

func scanLinterRow(rows pgx.Row) (dto.Linter, error) {
	var (
		linterId                string
		linterGitUrl            string
		linterGitBranch         string
		linterLastDockerImage   *string
		linterLastDockerShaHash *string
	)
	err := rows.Scan(&linterId, &linterGitUrl, &linterGitBranch, &linterLastDockerImage, &linterLastDockerShaHash)
	if err != nil {
		return dto.Linter{}, err
	}
	meta := dto.LinterMeta{
		LinterId:        linterId,
		LinterGitUrl:    linterGitUrl,
		LinterGitBranch: linterGitBranch,
	}
	var instance *dto.LinterInstance
	if linterLastDockerImage != nil && linterLastDockerShaHash != nil {
		instance = &dto.LinterInstance{
			LinterId:           linterId,
			DockerImage:        *linterLastDockerImage,
			DockerImageShaHash: *linterLastDockerShaHash,
		}
	}
	return dto.Linter{Meta: meta, Instance: instance}, nil
}
