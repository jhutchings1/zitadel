package view

import (
	proj_model "github.com/caos/zitadel/internal/project/model"
	"github.com/caos/zitadel/internal/project/repository/view"
	"github.com/caos/zitadel/internal/project/repository/view/model"
	"github.com/caos/zitadel/internal/view/repository"
)

const (
	grantedProjectTable = "management.project_grants"
)

func (v *View) ProjectGrantByID(grantID string) (*model.ProjectGrantView, error) {
	return view.ProjectGrantByID(v.Db, grantedProjectTable, grantID)
}

func (v *View) ProjectGrantByProjectAndOrg(projectID, orgID string) (*model.ProjectGrantView, error) {
	return view.ProjectGrantByProjectAndOrg(v.Db, grantedProjectTable, projectID, orgID)
}

func (v *View) ProjectGrantsByProjectID(projectID string) ([]*model.ProjectGrantView, error) {
	return view.ProjectGrantsByProjectID(v.Db, grantedProjectTable, projectID)
}

func (v *View) ProjectGrantsByProjectIDAndRoleKey(projectID, key string) ([]*model.ProjectGrantView, error) {
	return view.ProjectGrantsByProjectIDAndRoleKey(v.Db, grantedProjectTable, projectID, key)
}

func (v *View) SearchProjectGrants(request *proj_model.ProjectGrantViewSearchRequest) ([]*model.ProjectGrantView, int, error) {
	return view.SearchProjectGrants(v.Db, grantedProjectTable, request)
}

func (v *View) PutProjectGrant(project *model.ProjectGrantView) error {
	err := view.PutProjectGrant(v.Db, grantedProjectTable, project)
	if err != nil {
		return err
	}
	return v.ProcessedProjectGrantSequence(project.Sequence)
}

func (v *View) DeleteProjectGrant(grantID string, eventSequence uint64) error {
	err := view.DeleteProjectGrant(v.Db, grantedProjectTable, grantID)
	if err != nil {
		return err
	}
	return v.ProcessedProjectGrantSequence(eventSequence)
}

func (v *View) GetLatestProjectGrantSequence() (*repository.CurrentSequence, error) {
	return v.latestSequence(grantedProjectTable)
}

func (v *View) ProcessedProjectGrantSequence(eventSequence uint64) error {
	return v.saveCurrentSequence(grantedProjectTable, eventSequence)
}

func (v *View) GetLatestProjectGrantFailedEvent(sequence uint64) (*repository.FailedEvent, error) {
	return v.latestFailedEvent(grantedProjectTable, sequence)
}

func (v *View) ProcessedProjectGrantFailedEvent(failedEvent *repository.FailedEvent) error {
	return v.saveFailedEvent(failedEvent)
}
