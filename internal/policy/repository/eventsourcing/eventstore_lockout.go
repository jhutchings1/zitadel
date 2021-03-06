package eventsourcing

import (
	"context"

	"github.com/caos/zitadel/internal/api/authz"
	caos_errs "github.com/caos/zitadel/internal/errors"
	es_sdk "github.com/caos/zitadel/internal/eventstore/sdk"
	pol_model "github.com/caos/zitadel/internal/policy/model"
)

func (es *PolicyEventstore) GetPasswordLockoutPolicy(ctx context.Context, id string) (*pol_model.PasswordLockoutPolicy, error) {
	policy := es.policyCache.getLockoutPolicy(id)

	query := PasswordLockoutPolicyQuery(id, policy.Sequence)
	err := es_sdk.Filter(ctx, es.FilterEvents, policy.AppendEvents, query)
	if caos_errs.IsNotFound(err) && es.passwordLockoutPolicyDefault.Description != "" {
		policy.Description = es.passwordLockoutPolicyDefault.Description
		policy.MaxAttempts = es.passwordLockoutPolicyDefault.MaxAttempts
		policy.ShowLockOutFailures = es.passwordLockoutPolicyDefault.ShowLockOutFailures
	} else if err != nil {
		return nil, err
	}
	es.policyCache.cacheLockoutPolicy(policy)
	return PasswordLockoutPolicyToModel(policy), nil
}

func (es *PolicyEventstore) CreatePasswordLockoutPolicy(ctx context.Context, policy *pol_model.PasswordLockoutPolicy) (*pol_model.PasswordLockoutPolicy, error) {
	ctxData := authz.GetCtxData(ctx)
	existingPolicy, err := es.GetPasswordLockoutPolicy(ctx, ctxData.OrgID)
	if err != nil && !caos_errs.IsNotFound(err) {
		return nil, err
	}
	if existingPolicy != nil && existingPolicy.Sequence > 0 {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "EVENT-yDJ5I", "Errors.Policy.AlreadyExists")
	}

	id, err := es.idGenerator.Next()
	if err != nil {
		return nil, err
	}
	policy.AggregateID = id

	repoPolicy := PasswordLockoutPolicyFromModel(policy)

	createAggregate := PasswordLockoutPolicyCreateAggregate(es.AggregateCreator(), repoPolicy)
	err = es_sdk.Push(ctx, es.PushAggregates, repoPolicy.AppendEvents, createAggregate)
	if err != nil {
		return nil, err
	}

	es.policyCache.cacheLockoutPolicy(repoPolicy)
	return PasswordLockoutPolicyToModel(repoPolicy), nil
}

func (es *PolicyEventstore) UpdatePasswordLockoutPolicy(ctx context.Context, policy *pol_model.PasswordLockoutPolicy) (*pol_model.PasswordLockoutPolicy, error) {
	ctxData := authz.GetCtxData(ctx)
	existingPolicy, err := es.GetPasswordLockoutPolicy(ctx, ctxData.OrgID)
	if err != nil {
		return nil, err
	}
	if existingPolicy.Sequence <= 0 {
		return es.CreatePasswordLockoutPolicy(ctx, policy)
	}
	repoExisting := PasswordLockoutPolicyFromModel(existingPolicy)
	repoNew := PasswordLockoutPolicyFromModel(policy)

	updateAggregate := PasswordLockoutPolicyUpdateAggregate(es.AggregateCreator(), repoExisting, repoNew)
	err = es_sdk.Push(ctx, es.PushAggregates, repoExisting.AppendEvents, updateAggregate)
	if err != nil {
		return nil, err
	}

	es.policyCache.cacheLockoutPolicy(repoExisting)
	return PasswordLockoutPolicyToModel(repoExisting), nil
}
