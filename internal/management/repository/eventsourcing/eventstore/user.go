package eventstore

import (
	"context"
	caos_errs "github.com/caos/zitadel/internal/errors"

	"github.com/caos/logging"

	"github.com/caos/zitadel/internal/api/authz"
	"github.com/caos/zitadel/internal/management/repository/eventsourcing/view"
	org_event "github.com/caos/zitadel/internal/org/repository/eventsourcing"
	policy_event "github.com/caos/zitadel/internal/policy/repository/eventsourcing"
	usr_model "github.com/caos/zitadel/internal/user/model"
	usr_event "github.com/caos/zitadel/internal/user/repository/eventsourcing"
	"github.com/caos/zitadel/internal/user/repository/view/model"
)

type UserRepo struct {
	SearchLimit  uint64
	UserEvents   *usr_event.UserEventstore
	PolicyEvents *policy_event.PolicyEventstore
	OrgEvents    *org_event.OrgEventstore
	View         *view.View
}

func (repo *UserRepo) UserByID(ctx context.Context, id string) (*usr_model.UserView, error) {
	user, viewErr := repo.View.UserByID(id)
	if viewErr != nil && !caos_errs.IsNotFound(viewErr) {
		return nil, viewErr
	}
	if caos_errs.IsNotFound(viewErr) {
		user = new(model.UserView)
	}
	events, esErr := repo.UserEvents.UserEventsByID(ctx, id, user.Sequence)
	if caos_errs.IsNotFound(viewErr) && len(events) == 0 {
		return nil, caos_errs.ThrowNotFound(nil, "EVENT-Lsoj7", "Errors.User.NotFound")
	}
	if esErr != nil {
		logging.Log("EVENT-PSoc3").WithError(esErr).Debug("error retrieving new events")
		return model.UserToModel(user), nil
	}
	userCopy := *user
	for _, event := range events {
		if err := userCopy.AppendEvent(event); err != nil {
			return model.UserToModel(user), nil
		}
	}
	return model.UserToModel(&userCopy), nil
}

func (repo *UserRepo) CreateUser(ctx context.Context, user *usr_model.User) (*usr_model.User, error) {
	pwPolicy, err := repo.PolicyEvents.GetPasswordComplexityPolicy(ctx, authz.GetCtxData(ctx).OrgID)
	if err != nil {
		return nil, err
	}
	orgPolicy, err := repo.OrgEvents.GetOrgIamPolicy(ctx, authz.GetCtxData(ctx).OrgID)
	if err != nil {
		return nil, err
	}
	return repo.UserEvents.CreateUser(ctx, user, pwPolicy, orgPolicy)
}

func (repo *UserRepo) RegisterUser(ctx context.Context, user *usr_model.User, resourceOwner string) (*usr_model.User, error) {
	policyResourceOwner := authz.GetCtxData(ctx).OrgID
	if resourceOwner != "" {
		policyResourceOwner = resourceOwner
	}
	pwPolicy, err := repo.PolicyEvents.GetPasswordComplexityPolicy(ctx, policyResourceOwner)
	if err != nil {
		return nil, err
	}
	orgPolicy, err := repo.OrgEvents.GetOrgIamPolicy(ctx, authz.GetCtxData(ctx).OrgID)
	if err != nil {
		return nil, err
	}
	return repo.UserEvents.RegisterUser(ctx, user, pwPolicy, orgPolicy, resourceOwner)
}

func (repo *UserRepo) DeactivateUser(ctx context.Context, id string) (*usr_model.User, error) {
	return repo.UserEvents.DeactivateUser(ctx, id)
}

func (repo *UserRepo) ReactivateUser(ctx context.Context, id string) (*usr_model.User, error) {
	return repo.UserEvents.ReactivateUser(ctx, id)
}

func (repo *UserRepo) LockUser(ctx context.Context, id string) (*usr_model.User, error) {
	return repo.UserEvents.LockUser(ctx, id)
}

func (repo *UserRepo) UnlockUser(ctx context.Context, id string) (*usr_model.User, error) {
	return repo.UserEvents.UnlockUser(ctx, id)
}

func (repo *UserRepo) SearchUsers(ctx context.Context, request *usr_model.UserSearchRequest) (*usr_model.UserSearchResponse, error) {
	request.EnsureLimit(repo.SearchLimit)
	sequence, err := repo.View.GetLatestUserSequence()
	logging.Log("EVENT-Lcn7d").OnError(err).Warn("could not read latest user sequence")
	projects, count, err := repo.View.SearchUsers(request)
	if err != nil {
		return nil, err
	}
	result := &usr_model.UserSearchResponse{
		Offset:      request.Offset,
		Limit:       request.Limit,
		TotalResult: uint64(count),
		Result:      model.UsersToModel(projects),
	}
	if err == nil {
		result.Sequence = sequence.CurrentSequence
		result.Timestamp = sequence.CurrentTimestamp
	}
	return result, nil
}

func (repo *UserRepo) UserChanges(ctx context.Context, id string, lastSequence uint64, limit uint64, sortAscending bool) (*usr_model.UserChanges, error) {
	changes, err := repo.UserEvents.UserChanges(ctx, id, lastSequence, limit, sortAscending)
	if err != nil {
		return nil, err
	}
	for _, change := range changes.Changes {
		change.ModifierName = change.ModifierId
		user, _ := repo.UserEvents.UserByID(ctx, change.ModifierId)
		if user != nil {
			change.ModifierName = user.DisplayName
		}
	}
	return changes, nil
}

func (repo *UserRepo) GetGlobalUserByEmail(ctx context.Context, email string) (*usr_model.UserView, error) {
	user, err := repo.View.GetGlobalUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return model.UserToModel(user), nil
}

func (repo *UserRepo) IsUserUnique(ctx context.Context, userName, email string) (bool, error) {
	return repo.View.IsUserUnique(userName, email)
}

func (repo *UserRepo) UserMfas(ctx context.Context, userID string) ([]*usr_model.MultiFactor, error) {
	return repo.View.UserMfas(userID)
}

func (repo *UserRepo) SetOneTimePassword(ctx context.Context, password *usr_model.Password) (*usr_model.Password, error) {
	policy, err := repo.PolicyEvents.GetPasswordComplexityPolicy(ctx, authz.GetCtxData(ctx).OrgID)
	if err != nil {
		return nil, err
	}
	return repo.UserEvents.SetOneTimePassword(ctx, policy, password)
}

func (repo *UserRepo) RequestSetPassword(ctx context.Context, id string, notifyType usr_model.NotificationType) error {
	return repo.UserEvents.RequestSetPassword(ctx, id, notifyType)
}

func (repo *UserRepo) ProfileByID(ctx context.Context, userID string) (*usr_model.Profile, error) {
	user, err := repo.UserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.GetProfile(), nil
}

func (repo *UserRepo) ChangeProfile(ctx context.Context, profile *usr_model.Profile) (*usr_model.Profile, error) {
	return repo.UserEvents.ChangeProfile(ctx, profile)
}

func (repo *UserRepo) EmailByID(ctx context.Context, userID string) (*usr_model.Email, error) {
	user, err := repo.UserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.GetEmail(), nil
}

func (repo *UserRepo) ChangeEmail(ctx context.Context, email *usr_model.Email) (*usr_model.Email, error) {
	return repo.UserEvents.ChangeEmail(ctx, email)
}

func (repo *UserRepo) CreateEmailVerificationCode(ctx context.Context, userID string) error {
	return repo.UserEvents.CreateEmailVerificationCode(ctx, userID)
}

func (repo *UserRepo) PhoneByID(ctx context.Context, userID string) (*usr_model.Phone, error) {
	user, err := repo.UserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.GetPhone(), nil
}

func (repo *UserRepo) ChangePhone(ctx context.Context, email *usr_model.Phone) (*usr_model.Phone, error) {
	return repo.UserEvents.ChangePhone(ctx, email)
}

func (repo *UserRepo) RemovePhone(ctx context.Context, userID string) error {
	return repo.UserEvents.RemovePhone(ctx, userID)
}

func (repo *UserRepo) CreatePhoneVerificationCode(ctx context.Context, userID string) error {
	return repo.UserEvents.CreatePhoneVerificationCode(ctx, userID)
}

func (repo *UserRepo) AddressByID(ctx context.Context, userID string) (*usr_model.Address, error) {
	user, err := repo.UserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.GetAddress(), nil
}

func (repo *UserRepo) ChangeAddress(ctx context.Context, address *usr_model.Address) (*usr_model.Address, error) {
	return repo.UserEvents.ChangeAddress(ctx, address)
}
