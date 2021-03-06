package view

import (
	org_model "github.com/caos/zitadel/internal/org/model"
	"github.com/caos/zitadel/internal/org/repository/view"
	"github.com/caos/zitadel/internal/org/repository/view/model"
	"github.com/caos/zitadel/internal/view/repository"
)

const (
	orgDomainTable = "management.org_domains"
)

func (v *View) OrgDomainByOrgIDAndDomain(orgID, domain string) (*model.OrgDomainView, error) {
	return view.OrgDomainByOrgIDAndDomain(v.Db, orgDomainTable, orgID, domain)
}

func (v *View) OrgDomainsByOrgID(domain string) ([]*model.OrgDomainView, error) {
	return view.OrgDomainsByOrgID(v.Db, orgDomainTable, domain)
}

func (v *View) VerifiedOrgDomain(domain string) (*model.OrgDomainView, error) {
	return view.VerifiedOrgDomain(v.Db, orgDomainTable, domain)
}

func (v *View) SearchOrgDomains(request *org_model.OrgDomainSearchRequest) ([]*model.OrgDomainView, int, error) {
	return view.SearchOrgDomains(v.Db, orgDomainTable, request)
}

func (v *View) PutOrgDomain(org *model.OrgDomainView, sequence uint64) error {
	err := view.PutOrgDomain(v.Db, orgDomainTable, org)
	if err != nil {
		return err
	}
	if sequence != 0 {
		return v.ProcessedOrgDomainSequence(sequence)
	}
	return nil
}

func (v *View) DeleteOrgDomain(orgID, domain string, eventSequence uint64) error {
	err := view.DeleteOrgDomain(v.Db, orgDomainTable, orgID, domain)
	if err != nil {
		return nil
	}
	return v.ProcessedOrgDomainSequence(eventSequence)
}

func (v *View) GetLatestOrgDomainSequence() (*repository.CurrentSequence, error) {
	return v.latestSequence(orgDomainTable)
}

func (v *View) ProcessedOrgDomainSequence(eventSequence uint64) error {
	return v.saveCurrentSequence(orgDomainTable, eventSequence)
}

func (v *View) GetLatestOrgDomainFailedEvent(sequence uint64) (*repository.FailedEvent, error) {
	return v.latestFailedEvent(orgDomainTable, sequence)
}

func (v *View) ProcessedOrgDomainFailedEvent(failedEvent *repository.FailedEvent) error {
	return v.saveFailedEvent(failedEvent)
}
