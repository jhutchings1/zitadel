package eventstore

import (
	"context"
	"github.com/caos/logging"

	"github.com/caos/zitadel/internal/api/authz"
	"github.com/caos/zitadel/internal/auth/repository/eventsourcing/view"
	authz_repo "github.com/caos/zitadel/internal/authz/repository/eventsourcing"
	caos_errs "github.com/caos/zitadel/internal/errors"
	global_model "github.com/caos/zitadel/internal/model"
	org_model "github.com/caos/zitadel/internal/org/model"
	org_view_model "github.com/caos/zitadel/internal/org/repository/view/model"
	grant_model "github.com/caos/zitadel/internal/usergrant/model"
	"github.com/caos/zitadel/internal/usergrant/repository/view/model"
)

type UserGrantRepo struct {
	SearchLimit uint64
	View        *view.View
	IamID       string
	Auth        authz.Config
	AuthZRepo   *authz_repo.EsRepository
}

func (repo *UserGrantRepo) SearchMyUserGrants(ctx context.Context, request *grant_model.UserGrantSearchRequest) (*grant_model.UserGrantSearchResponse, error) {
	request.EnsureLimit(repo.SearchLimit)
	sequence, err := repo.View.GetLatestUserGrantSequence()
	logging.Log("EVENT-Hd7s3").OnError(err).Warn("could not read latest user grant sequence")
	request.Queries = append(request.Queries, &grant_model.UserGrantSearchQuery{Key: grant_model.UserGrantSearchKeyUserID, Method: global_model.SearchMethodEquals, Value: authz.GetCtxData(ctx).UserID})
	grants, count, err := repo.View.SearchUserGrants(request)
	if err != nil {
		return nil, err
	}
	result := &grant_model.UserGrantSearchResponse{
		Offset:      request.Offset,
		Limit:       request.Limit,
		TotalResult: uint64(count),
		Result:      model.UserGrantsToModel(grants),
	}
	if err == nil {
		result.Sequence = sequence.CurrentSequence
		result.Timestamp = sequence.CurrentTimestamp
	}
	return result, nil
}

func (repo *UserGrantRepo) SearchMyProjectOrgs(ctx context.Context, request *grant_model.UserGrantSearchRequest) (*grant_model.ProjectOrgSearchResponse, error) {
	request.EnsureLimit(repo.SearchLimit)
	ctxData := authz.GetCtxData(ctx)
	if ctxData.ProjectID == "" {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "APP-7lqva", "Could not get ProjectID")
	}
	err := repo.AuthZRepo.FillIamProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if ctxData.ProjectID == repo.AuthZRepo.UserGrantRepo.IamProjectID {
		isAdmin, err := repo.IsIamAdmin(ctx)
		if err != nil {
			return nil, err
		}
		if isAdmin {
			return repo.SearchAdminOrgs(request)
		}
	}
	request.Queries = append(request.Queries, &grant_model.UserGrantSearchQuery{Key: grant_model.UserGrantSearchKeyProjectID, Method: global_model.SearchMethodEquals, Value: ctxData.ProjectID})

	grants, err := repo.SearchMyUserGrants(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(grants.Result) > 0 {
		return grantRespToOrgResp(grants), nil
	}
	user, err := repo.View.UserByID(ctxData.UserID)
	if err != nil {
		return nil, err
	}
	org, err := repo.View.OrgByID(user.ResourceOwner)
	if err != nil {
		return nil, err
	}
	return &grant_model.ProjectOrgSearchResponse{Result: []*grant_model.Org{&grant_model.Org{
		OrgID:   org.ID,
		OrgName: org.Name,
	}}}, nil
}

func (repo *UserGrantRepo) SearchMyZitadelPermissions(ctx context.Context) ([]string, error) {
	grant, err := repo.AuthZRepo.ResolveGrants(ctx)
	if err != nil {
		return nil, err
	}
	if grant == nil {
		return []string{}, nil
	}
	permissions := &grant_model.Permissions{Permissions: []string{}}
	for _, role := range grant.Roles {
		roleName, ctxID := authz.SplitPermission(role)
		for _, mapping := range repo.Auth.RolePermissionMappings {
			if mapping.Role == roleName {
				permissions.AppendPermissions(ctxID, mapping.Permissions...)
			}
		}
	}
	return permissions.Permissions, nil
}

func (repo *UserGrantRepo) SearchMyProjectPermissions(ctx context.Context) ([]string, error) {
	ctxData := authz.GetCtxData(ctx)
	usergrant, err := repo.View.UserGrantByIDs(ctxData.OrgID, ctxData.ProjectID, ctxData.UserID)
	if err != nil {
		return nil, err
	}
	permissions := make([]string, len(usergrant.RoleKeys))
	for i, role := range usergrant.RoleKeys {
		permissions[i] = role
	}
	return permissions, nil
}

func (repo *UserGrantRepo) SearchAdminOrgs(request *grant_model.UserGrantSearchRequest) (*grant_model.ProjectOrgSearchResponse, error) {
	searchRequest := &org_model.OrgSearchRequest{}
	if len(request.Queries) > 0 {
		for _, q := range request.Queries {
			if q.Key == grant_model.UserGrantSearchKeyOrgName {
				searchRequest.Queries = append(searchRequest.Queries, &org_model.OrgSearchQuery{Key: org_model.OrgSearchKeyOrgName, Method: q.Method, Value: q.Value})
			}
		}
	}
	orgs, count, err := repo.View.SearchOrgs(searchRequest)
	if err != nil {
		return nil, err
	}
	return orgRespToOrgResp(orgs, count), nil
}

func (repo *UserGrantRepo) IsIamAdmin(ctx context.Context) (bool, error) {
	grantSearch := &grant_model.UserGrantSearchRequest{
		Queries: []*grant_model.UserGrantSearchQuery{
			{Key: grant_model.UserGrantSearchKeyResourceOwner, Method: global_model.SearchMethodEquals, Value: repo.IamID},
		}}
	result, err := repo.SearchMyUserGrants(ctx, grantSearch)
	if err != nil {
		return false, err
	}
	if result.TotalResult == 0 {
		return false, nil
	}
	return true, nil
}

func (repo *UserGrantRepo) UserGrantsByProjectAndUserID(projectID, userID string) ([]*grant_model.UserGrantView, error) {
	grants, err := repo.View.UserGrantsByProjectAndUserID(projectID, userID)
	if err != nil {
		return nil, err
	}
	return model.UserGrantsToModel(grants), nil
}

func grantRespToOrgResp(grants *grant_model.UserGrantSearchResponse) *grant_model.ProjectOrgSearchResponse {
	resp := &grant_model.ProjectOrgSearchResponse{
		TotalResult: grants.TotalResult,
	}
	resp.Result = make([]*grant_model.Org, len(grants.Result))
	for i, g := range grants.Result {
		resp.Result[i] = &grant_model.Org{OrgID: g.ResourceOwner, OrgName: g.OrgName}
	}
	return resp
}

func orgRespToOrgResp(orgs []*org_view_model.OrgView, count int) *grant_model.ProjectOrgSearchResponse {
	resp := &grant_model.ProjectOrgSearchResponse{
		TotalResult: uint64(count),
	}
	resp.Result = make([]*grant_model.Org, len(orgs))
	for i, o := range orgs {
		resp.Result[i] = &grant_model.Org{OrgID: o.ID, OrgName: o.Name}
	}
	return resp
}

func mergeOrgAndAdminGrant(ctxData authz.CtxData, orgGrant, iamAdminGrant *model.UserGrantView) (grant *authz.Grant) {
	if orgGrant != nil {
		roles := orgGrant.RoleKeys
		if iamAdminGrant != nil {
			roles = addIamAdminRoles(roles, iamAdminGrant.RoleKeys)
		}
		grant = &authz.Grant{OrgID: orgGrant.ResourceOwner, Roles: roles}
	} else if iamAdminGrant != nil {
		grant = &authz.Grant{
			OrgID: ctxData.OrgID,
			Roles: iamAdminGrant.RoleKeys,
		}
	}
	return grant
}

func addIamAdminRoles(orgRoles, iamAdminRoles []string) []string {
	result := make([]string, 0)
	result = append(result, iamAdminRoles...)
	for _, role := range orgRoles {
		if !authz.ExistsPerm(result, role) {
			result = append(result, role)
		}
	}
	return result
}
