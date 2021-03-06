package view

import (
	org_model "github.com/caos/zitadel/internal/org/model"
	"github.com/caos/zitadel/internal/org/repository/view"
	"github.com/caos/zitadel/internal/org/repository/view/model"
	"github.com/caos/zitadel/internal/view/repository"
)

const (
	orgMemberTable = "management.org_members"
)

func (v *View) OrgMemberByIDs(orgID, userID string) (*model.OrgMemberView, error) {
	return view.OrgMemberByIDs(v.Db, orgMemberTable, orgID, userID)
}

func (v *View) SearchOrgMembers(request *org_model.OrgMemberSearchRequest) ([]*model.OrgMemberView, int, error) {
	return view.SearchOrgMembers(v.Db, orgMemberTable, request)
}

func (v *View) OrgMembersByUserID(userID string) ([]*model.OrgMemberView, error) {
	return view.OrgMembersByUserID(v.Db, orgMemberTable, userID)
}

func (v *View) PutOrgMember(org *model.OrgMemberView, sequence uint64) error {
	err := view.PutOrgMember(v.Db, orgMemberTable, org)
	if err != nil {
		return err
	}
	return v.ProcessedOrgMemberSequence(sequence)
}

func (v *View) DeleteOrgMember(orgID, userID string, eventSequence uint64) error {
	err := view.DeleteOrgMember(v.Db, orgMemberTable, orgID, userID)
	if err != nil {
		return nil
	}
	return v.ProcessedOrgMemberSequence(eventSequence)
}

func (v *View) GetLatestOrgMemberSequence() (*repository.CurrentSequence, error) {
	return v.latestSequence(orgMemberTable)
}

func (v *View) ProcessedOrgMemberSequence(eventSequence uint64) error {
	return v.saveCurrentSequence(orgMemberTable, eventSequence)
}

func (v *View) GetLatestOrgMemberFailedEvent(sequence uint64) (*repository.FailedEvent, error) {
	return v.latestFailedEvent(orgMemberTable, sequence)
}

func (v *View) ProcessedOrgMemberFailedEvent(failedEvent *repository.FailedEvent) error {
	return v.saveFailedEvent(failedEvent)
}
