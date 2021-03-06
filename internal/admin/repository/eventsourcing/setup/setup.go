package setup

import (
	"context"
	"time"

	"github.com/caos/logging"

	"github.com/caos/zitadel/internal/api/authz"
	"github.com/caos/zitadel/internal/config/systemdefaults"
	"github.com/caos/zitadel/internal/config/types"
	caos_errs "github.com/caos/zitadel/internal/errors"
	"github.com/caos/zitadel/internal/eventstore/models"
	iam_model "github.com/caos/zitadel/internal/iam/model"
	iam_event "github.com/caos/zitadel/internal/iam/repository/eventsourcing"
	org_model "github.com/caos/zitadel/internal/org/model"
	org_event "github.com/caos/zitadel/internal/org/repository/eventsourcing"
	policy_model "github.com/caos/zitadel/internal/policy/model"
	policy_event "github.com/caos/zitadel/internal/policy/repository/eventsourcing"
	proj_model "github.com/caos/zitadel/internal/project/model"
	proj_event "github.com/caos/zitadel/internal/project/repository/eventsourcing"
	usr_model "github.com/caos/zitadel/internal/user/model"
	usr_event "github.com/caos/zitadel/internal/user/repository/eventsourcing"
)

type Setup struct {
	repos       EventstoreRepos
	iamID       string
	setUpConfig types.IAMSetUp
}

type EventstoreRepos struct {
	IamEvents     *iam_event.IamEventstore
	OrgEvents     *org_event.OrgEventstore
	UserEvents    *usr_event.UserEventstore
	ProjectEvents *proj_event.ProjectEventstore
	PolicyEvents  *policy_event.PolicyEventstore
}

type initializer struct {
	*Setup
	createdUsers       map[string]*usr_model.User
	createdOrgs        map[string]*org_model.Org
	createdProjects    map[string]*proj_model.Project
	pwComplexityPolicy *policy_model.PasswordComplexityPolicy
}

const (
	OrgOwnerRole                   = "ORG_OWNER"
	SetupUser                      = "SETUP"
	OIDCResponseTypeCode           = "CODE"
	OIDCResponseTypeIDToken        = "ID_TOKEN"
	OIDCResponseTypeToken          = "ID_TOKEN TOKEN"
	OIDCGrantTypeAuthorizationCode = "AUTHORIZATION_CODE"
	OIDCGrantTypeImplicit          = "IMPLICIT"
	OIDCGrantTypeRefreshToken      = "REFRESH_TOKEN"
	OIDCApplicationTypeNative      = "NATIVE"
	OIDCApplicationTypeUserAgent   = "USER_AGENT"
	OIDCApplicationTypeWeb         = "WEB"
	OIDCAuthMethodTypeNone         = "NONE"
	OIDCAuthMethodTypeBasic        = "BASIC"
	OIDCAuthMethodTypePost         = "POST"
)

func StartSetup(sd systemdefaults.SystemDefaults, repos EventstoreRepos) *Setup {
	return &Setup{
		repos:       repos,
		iamID:       sd.IamID,
		setUpConfig: sd.SetUp,
	}
}

func (s *Setup) Execute(ctx context.Context) error {
	iam, err := s.repos.IamEvents.IamByID(ctx, s.iamID)
	if err != nil && !caos_errs.IsNotFound(err) {
		return err
	}
	if iam != nil && iam.SetUpDone {
		return nil
	}

	if iam != nil && iam.SetUpStarted {
		return s.waitForSetupDone(ctx)
	}

	logging.Log("SETUP-hwG32").Info("starting setup")
	ctx = setSetUpContextData(ctx, s.iamID)
	iam, err = s.repos.IamEvents.StartSetup(ctx, s.iamID)
	if err != nil {
		return err
	}

	setUp := &initializer{
		Setup:           s,
		createdUsers:    make(map[string]*usr_model.User),
		createdOrgs:     make(map[string]*org_model.Org),
		createdProjects: make(map[string]*proj_model.Project),
	}

	pwComplexityPolicy, err := s.repos.PolicyEvents.GetPasswordComplexityPolicy(ctx, policy_model.DefaultPolicy)
	if err != nil {
		logging.Log("SETUP-9osWF").WithError(err).Error("unable to read complexity policy")
		return err
	}
	setUp.pwComplexityPolicy = pwComplexityPolicy

	err = setUp.orgs(ctx, s.setUpConfig.Orgs)
	if err != nil {
		logging.Log("SETUP-p4oWq").WithError(err).Error("unable to set up orgs")
		return err
	}

	ctx = setSetUpContextData(ctx, s.iamID)
	err = setUp.iamOwners(ctx, s.setUpConfig.Owners)
	if err != nil {
		logging.Log("SETUP-WHr01").WithError(err).Error("unable to set up iam owners")
		return err
	}

	err = setUp.setGlobalOrg(ctx)
	if err != nil {
		logging.Log("SETUP-0874m").WithError(err).Error("unable to set global org")
		return err
	}

	err = setUp.setIamProject(ctx)
	if err != nil {
		logging.Log("SETUP-kaWjq").WithError(err).Error("unable to set zitadel project")
		return err
	}

	iam, err = s.repos.IamEvents.SetupDone(ctx, s.iamID)
	if err != nil {
		logging.Log("SETUP-de342").WithError(err).Error("unable to finish setup")
		return err
	}
	logging.Log("SETUP-ds31h").Info("setup done")
	return nil
}

func (s *Setup) waitForSetupDone(ctx context.Context) error {
	logging.Log("SETUP-hws22").Info("waiting for setup to be done")
	ctx, cancel := context.WithDeadline(ctx, time.Now().UTC().Add(10*time.Second))
	defer cancel()

	for {
		select {
		case <-time.After(1 * time.Second):
			iam, _ := s.repos.IamEvents.IamByID(ctx, s.iamID)
			if iam != nil && iam.SetUpDone {
				return nil
			}
			logging.Log("SETUP-d23g1").Info("setup not done yet")
		case <-ctx.Done():
			return caos_errs.ThrowInternal(ctx.Err(), "SETUP-dsjg3", "Timeout exceeded for setup")
		}
	}
}

func (setUp *initializer) orgs(ctx context.Context, orgs []types.Org) error {
	logging.Log("SETUP-dsTh3").Info("setting up orgs")
	for _, iamOrg := range orgs {
		org, err := setUp.org(ctx, iamOrg)
		if err != nil {
			logging.LogWithFields("SETUP-IlLif", "Org", iamOrg.Name).WithError(err).Error("unable to create org")
			return err
		}
		setUp.createdOrgs[iamOrg.Name] = org

		var policy *org_model.OrgIamPolicy
		if iamOrg.OrgIamPolicy {
			policy, err = setUp.iamorgpolicy(ctx, org)
			if err != nil {
				logging.LogWithFields("SETUP-IlLif", "Org Iam Policy", iamOrg.Name).WithError(err).Error("unable to create iam org policy")
				return err
			}
		} else {
			policy, err = setUp.repos.OrgEvents.GetOrgIamPolicy(ctx, policy_model.DefaultPolicy)
			if err != nil {
				logging.LogWithFields("SETUP-IS8wS", "Org Iam Policy", iamOrg.Name).WithError(err).Error("unable to get default iam org policy")
				return err
			}
		}

		ctx = setSetUpContextData(ctx, org.AggregateID)
		err = setUp.users(ctx, iamOrg.Users, policy)
		if err != nil {
			logging.LogWithFields("SETUP-8zfwz", "Org", iamOrg.Name).WithError(err).Error("unable to set up org users")
			return err
		}

		err = setUp.orgOwners(ctx, org, iamOrg.Owners)
		if err != nil {
			logging.LogWithFields("SETUP-0874m", "Org", iamOrg.Name).WithError(err).Error("unable to set up org owners")
			return err
		}

		err = setUp.projects(ctx, iamOrg.Projects)
		if err != nil {
			logging.LogWithFields("SETUP-wUzqY", "Org", iamOrg.Name).WithError(err).Error("unable to set up org projects")
			return err
		}
	}
	logging.Log("SETUP-dgjT4").Info("orgs set up")
	return nil
}

func (setUp *initializer) org(ctx context.Context, org types.Org) (*org_model.Org, error) {
	ctx = setSetUpContextData(ctx, "")
	createOrg := &org_model.Org{
		Name:    org.Name,
		Domains: []*org_model.OrgDomain{{Domain: org.Domain}},
	}
	return setUp.repos.OrgEvents.CreateOrg(ctx, createOrg)
}

func (setUp *initializer) iamorgpolicy(ctx context.Context, org *org_model.Org) (*org_model.OrgIamPolicy, error) {
	ctx = setSetUpContextData(ctx, org.AggregateID)
	policy := &org_model.OrgIamPolicy{
		ObjectRoot:            models.ObjectRoot{AggregateID: org.AggregateID},
		UserLoginMustBeDomain: false,
	}
	return setUp.repos.OrgEvents.AddOrgIamPolicy(ctx, policy)
}

func (setUp *initializer) iamOwners(ctx context.Context, owners []string) error {
	logging.Log("SETUP-dtxfj").Info("setting iam owners")
	for _, iamOwner := range owners {
		user, ok := setUp.createdUsers[iamOwner]
		if !ok {
			logging.LogWithFields("SETUP-8siew", "Owner", iamOwner).Error("unable to add user to iam members")
			return caos_errs.ThrowPreconditionFailedf(nil, "SETUP-su6L3", "unable to add user to iam members")
		}
		_, err := setUp.repos.IamEvents.AddIamMember(ctx, &iam_model.IamMember{ObjectRoot: models.ObjectRoot{AggregateID: setUp.iamID}, UserID: user.AggregateID, Roles: []string{"IAM_OWNER"}})
		if err != nil {
			logging.Log("SETUP-LM7rI").WithError(err).Error("unable to add iam administrator to iam members as owner")
			return err
		}
	}
	logging.Log("SETUP-fg5aq").Info("iam owners set")
	return nil
}

func (setUp *initializer) setGlobalOrg(ctx context.Context) error {
	logging.Log("SETUP-dsj75").Info("setting global org")
	globalOrg, ok := setUp.createdOrgs[setUp.setUpConfig.GlobalOrg]
	if !ok {
		logging.LogWithFields("SETUP-FBhs9", "GlobalOrg", setUp.setUpConfig.GlobalOrg).Error("global org not created")
		return caos_errs.ThrowPreconditionFailedf(nil, "SETUP-4GwU7", "global org not created: %v", setUp.setUpConfig.GlobalOrg)
	}

	if _, err := setUp.repos.IamEvents.SetGlobalOrg(ctx, setUp.iamID, globalOrg.AggregateID); err != nil {
		logging.Log("SETUP-uGMA3").WithError(err).Error("unable to set global org on iam")
		return err
	}
	logging.Log("SETUP-d32h1").Info("global org set")
	return nil
}

func (setUp *initializer) setIamProject(ctx context.Context) error {
	logging.Log("SETUP-HE3qa").Info("setting iam project")
	iamProject, ok := setUp.createdProjects[setUp.setUpConfig.IAMProject]
	if !ok {
		logging.LogWithFields("SETUP-SJFWP", "Iam Project", setUp.setUpConfig.IAMProject).Error("iam project created")
		return caos_errs.ThrowPreconditionFailedf(nil, "SETUP-sGmQt", "iam project not created: %v", setUp.setUpConfig.IAMProject)
	}

	if _, err := setUp.repos.IamEvents.SetIamProject(ctx, setUp.iamID, iamProject.AggregateID); err != nil {
		logging.Log("SETUP-i1pNh").WithError(err).Error("unable to set iam project on iam")
		return err
	}
	logging.Log("SETUP-d7WEU").Info("iam project set")
	return nil
}

func (setUp *initializer) users(ctx context.Context, users []types.User, orgPolicy *org_model.OrgIamPolicy) error {
	for _, user := range users {
		created, err := setUp.user(ctx, user, orgPolicy)
		if err != nil {
			logging.LogWithFields("SETUP-9soer", "Email", user.Email).WithError(err).Error("unable to create iam user")
			return err
		}
		setUp.createdUsers[user.Email] = created
	}
	return nil
}

func (setUp *initializer) user(ctx context.Context, user types.User, orgPolicy *org_model.OrgIamPolicy) (*usr_model.User, error) {
	createUser := &usr_model.User{
		Profile: &usr_model.Profile{
			UserName:  user.UserName,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		Email: &usr_model.Email{
			EmailAddress:    user.Email,
			IsEmailVerified: true,
		},
		Password: &usr_model.Password{
			SecretString: user.Password,
		},
	}
	return setUp.repos.UserEvents.CreateUser(ctx, createUser, setUp.pwComplexityPolicy, orgPolicy)
}

func (setUp *initializer) orgOwners(ctx context.Context, org *org_model.Org, owners []string) error {
	for _, orgOwner := range owners {
		user, ok := setUp.createdUsers[orgOwner]
		if !ok {
			logging.LogWithFields("SETUP-s9ilr", "Owner", orgOwner).Error("unable to add user to org members")
			return caos_errs.ThrowPreconditionFailedf(nil, "SETUP-s0prs", "unable to add user to org members: %v", orgOwner)
		}
		err := setUp.orgOwner(ctx, org, user)
		if err != nil {
			logging.Log("SETUP-s90oe").WithError(err).Error("unable to add global org admin to members of global org")
			return err
		}
	}
	return nil
}

func (setUp *initializer) orgOwner(ctx context.Context, org *org_model.Org, user *usr_model.User) error {
	addMember := &org_model.OrgMember{
		ObjectRoot: models.ObjectRoot{AggregateID: org.AggregateID},
		UserID:     user.AggregateID,
		Roles:      []string{OrgOwnerRole},
	}
	_, err := setUp.repos.OrgEvents.AddOrgMember(ctx, addMember)
	return err
}

func (setUp *initializer) projects(ctx context.Context, projects []types.Project) error {
	for _, project := range projects {
		createdProject, err := setUp.project(ctx, project)
		if err != nil {
			return err
		}
		setUp.createdProjects[createdProject.Name] = createdProject
		for _, oidc := range project.OIDCApps {
			app, err := setUp.oidcApp(ctx, createdProject, oidc)
			if err != nil {
				return err
			}
			logging.LogWithFields("SETUP-asd32f", "name", app.Name, "clientID", app.OIDCConfig.ClientID).Info("created OIDC application")
		}
	}
	return nil
}

func (setUp *initializer) project(ctx context.Context, project types.Project) (*proj_model.Project, error) {
	addProject := &proj_model.Project{
		Name: project.Name,
	}
	return setUp.repos.ProjectEvents.CreateProject(ctx, addProject)
}

func (setUp *initializer) oidcApp(ctx context.Context, project *proj_model.Project, oidc types.OIDCApp) (*proj_model.Application, error) {
	addOIDCApp := &proj_model.Application{
		ObjectRoot: models.ObjectRoot{AggregateID: project.AggregateID},
		Name:       oidc.Name,
		OIDCConfig: &proj_model.OIDCConfig{
			RedirectUris:           oidc.RedirectUris,
			ResponseTypes:          getOIDCResponseTypes(oidc.ResponseTypes),
			GrantTypes:             getOIDCGrantTypes(oidc.GrantTypes),
			ApplicationType:        getOIDCApplicationType(oidc.ApplicationType),
			AuthMethodType:         getOIDCAuthMethod(oidc.AuthMethodType),
			PostLogoutRedirectUris: oidc.PostLogoutRedirectUris,
		},
	}
	return setUp.repos.ProjectEvents.AddApplication(ctx, addOIDCApp)
}

func getOIDCResponseTypes(responseTypes []string) []proj_model.OIDCResponseType {
	types := make([]proj_model.OIDCResponseType, len(responseTypes))
	for i, t := range responseTypes {
		types[i] = getOIDCResponseType(t)
	}
	return types
}

func getOIDCResponseType(responseType string) proj_model.OIDCResponseType {
	switch responseType {
	case OIDCResponseTypeCode:
		return proj_model.OIDCResponseTypeCode
	case OIDCResponseTypeIDToken:
		return proj_model.OIDCResponseTypeIDToken
	case OIDCResponseTypeToken:
		return proj_model.OIDCResponseTypeIDTokenToken
	}
	return proj_model.OIDCResponseTypeCode
}

func getOIDCGrantTypes(grantTypes []string) []proj_model.OIDCGrantType {
	types := make([]proj_model.OIDCGrantType, len(grantTypes))
	for i, t := range grantTypes {
		types[i] = getOIDCGrantType(t)
	}
	return types
}

func getOIDCGrantType(grantTypes string) proj_model.OIDCGrantType {
	switch grantTypes {
	case OIDCGrantTypeAuthorizationCode:
		return proj_model.OIDCGrantTypeAuthorizationCode
	case OIDCGrantTypeImplicit:
		return proj_model.OIDCGrantTypeImplicit
	case OIDCGrantTypeRefreshToken:
		return proj_model.OIDCGrantTypeRefreshToken
	}
	return proj_model.OIDCGrantTypeAuthorizationCode
}

func getOIDCApplicationType(appType string) proj_model.OIDCApplicationType {
	switch appType {
	case OIDCApplicationTypeNative:
		return proj_model.OIDCApplicationTypeNative
	case OIDCApplicationTypeUserAgent:
		return proj_model.OIDCApplicationTypeUserAgent
	case OIDCApplicationTypeWeb:
		return proj_model.OIDCApplicationTypeWeb
	}
	return proj_model.OIDCApplicationTypeWeb
}

func getOIDCAuthMethod(authMethod string) proj_model.OIDCAuthMethodType {
	switch authMethod {
	case OIDCAuthMethodTypeNone:
		return proj_model.OIDCAuthMethodTypeNone
	case OIDCAuthMethodTypeBasic:
		return proj_model.OIDCAuthMethodTypeBasic
	case OIDCAuthMethodTypePost:
		return proj_model.OIDCAuthMethodTypePost
	}
	return proj_model.OIDCAuthMethodTypeNone
}

func setSetUpContextData(ctx context.Context, orgID string) context.Context {
	return authz.SetCtxData(ctx, authz.CtxData{UserID: SetupUser, OrgID: orgID})
}
