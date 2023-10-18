//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportProject = errors.New("project related functions are enterprise version functions")

func (d *ProjectUsecase) CreateProject(ctx context.Context, project *Project, createUserUID string) (err error) {
	return errNotSupportProject
}

func (d *ProjectUsecase) GetProjectByName(ctx context.Context, projectName string) (*Project, error) {
	return nil, errNotSupportProject
}

func (d *ProjectUsecase) UpdateProjectDesc(ctx context.Context, currentUserUid, projectUid string, desc *string) (err error) {

	return errNotSupportProject
}

func (d *ProjectUsecase) ArchivedProject(ctx context.Context, currentUserUid, projectUid string, archived bool) (err error) {

	return errNotSupportProject
}

func (d *ProjectUsecase) DeleteProject(ctx context.Context, currentUserUid, projectUid string) (err error) {
	return errNotSupportProject
}

func (d *ProjectUsecase) isProjectActive(ctx context.Context, projectUid string) error {
	return nil
}
