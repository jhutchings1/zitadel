package eventsourcing

import (
	"context"
	"testing"

	"github.com/caos/zitadel/internal/api/authz"
	"github.com/caos/zitadel/internal/errors"
	es_models "github.com/caos/zitadel/internal/eventstore/models"
	org_model "github.com/caos/zitadel/internal/org/model"
	"github.com/caos/zitadel/internal/org/repository/eventsourcing/model"
)

func Test_isReservedValidation(t *testing.T) {
	type res struct {
		isErr              func(error) bool
		agggregateSequence uint64
	}
	type args struct {
		aggregate *es_models.Aggregate
		eventType es_models.EventType
		Events    []*es_models.Event
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no events success",
			args: args{
				aggregate: &es_models.Aggregate{},
				eventType: "object.reserved",
				Events:    []*es_models.Event{},
			},
			res: res{
				isErr:              nil,
				agggregateSequence: 0,
			},
		},
		{
			name: "not reseved success",
			args: args{
				aggregate: &es_models.Aggregate{},
				eventType: "object.reserved",
				Events: []*es_models.Event{
					{
						AggregateID:   "asdf",
						AggregateType: "org",
						Sequence:      45,
						Type:          "object.released",
					},
				},
			},
			res: res{
				isErr:              nil,
				agggregateSequence: 45,
			},
		},
		{
			name: "reseved error",
			args: args{
				aggregate: &es_models.Aggregate{},
				eventType: "object.reserved",
				Events: []*es_models.Event{
					{
						AggregateID:   "asdf",
						AggregateType: "org",
						Sequence:      45,
						Type:          "object.reserved",
					},
				},
			},
			res: res{
				isErr:              errors.IsPreconditionFailed,
				agggregateSequence: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := isEventValidation(tt.args.aggregate, tt.args.eventType)

			err := validate(tt.args.Events...)

			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got: %v", err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got: %v", err)
			}
			if err == nil && tt.args.aggregate.PreviousSequence != tt.res.agggregateSequence {
				t.Errorf("expected sequence %d got %d", tt.res.agggregateSequence, tt.args.aggregate.PreviousSequence)
			}
		})
	}
}

func aggregateWithPrecondition() *es_models.Aggregate {
	return nil
}

func Test_uniqueNameAggregate(t *testing.T) {
	type res struct {
		expected *es_models.Aggregate
		isErr    func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		orgName    string
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no org name error",
			args: args{
				ctx:        authz.NewMockContext("orgID", "userID"),
				aggCreator: es_models.NewAggregateCreator("test"),
				orgName:    "",
			},
			res: res{
				expected: nil,
				isErr:    errors.IsPreconditionFailed,
			},
		},
		{
			name: "aggregate created",
			args: args{
				ctx:        authz.NewMockContext("orgID", "userID"),
				aggCreator: es_models.NewAggregateCreator("test"),
				orgName:    "asdf",
			},
			res: res{
				expected: aggregateWithPrecondition(),
				isErr:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reservedUniqueNameAggregate(tt.args.ctx, tt.args.aggCreator, "", tt.args.orgName)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got: %v", err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && (got.Precondition == nil || got.Precondition.Query == nil || got.Precondition.Validation == nil) {
				t.Errorf("precondition is not set correctly")
			}
		})
	}
}

func Test_uniqueDomainAggregate(t *testing.T) {
	type res struct {
		expected *es_models.Aggregate
		isErr    func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		orgDomain  string
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no org domain error",
			args: args{
				ctx:        authz.NewMockContext("orgID", "userID"),
				aggCreator: es_models.NewAggregateCreator("test"),
				orgDomain:  "",
			},
			res: res{
				expected: nil,
				isErr:    errors.IsPreconditionFailed,
			},
		},
		{
			name: "aggregate created",
			args: args{
				ctx:        authz.NewMockContext("orgID", "userID"),
				aggCreator: es_models.NewAggregateCreator("test"),
				orgDomain:  "asdf",
			},
			res: res{
				expected: aggregateWithPrecondition(),
				isErr:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reservedUniqueDomainAggregate(tt.args.ctx, tt.args.aggCreator, "", tt.args.orgDomain)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got: %v", err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && (got.Precondition == nil || got.Precondition.Query == nil || got.Precondition.Validation == nil) {
				t.Errorf("precondition is not set correctly")
			}
		})
	}
}

func TestOrgReactivateAggregate(t *testing.T) {
	type res struct {
		isErr func(error) bool
	}
	type args struct {
		aggCreator *es_models.AggregateCreator
		org        *model.Org
		ctx        context.Context
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "correct",
			args: args{
				aggCreator: es_models.NewAggregateCreator("test"),
				ctx:        authz.NewMockContext("org", "user"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "orgID",
						Sequence:    2,
					},
					State: int32(org_model.OrgStateInactive),
				},
			},
		},
		{
			name: "already active error",
			args: args{
				aggCreator: es_models.NewAggregateCreator("test"),
				ctx:        authz.NewMockContext("org", "user"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "orgID",
						Sequence:    2,
					},
					State: int32(org_model.OrgStateActive),
				},
			},
			res: res{
				isErr: errors.IsErrorInvalidArgument,
			},
		},
		{
			name: "org nil error",
			args: args{
				aggCreator: es_models.NewAggregateCreator("test"),
				ctx:        authz.NewMockContext("org", "user"),
				org:        nil,
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregateCreator := orgReactivateAggregate(tt.args.aggCreator, tt.args.org)
			aggregate, err := aggregateCreator(tt.args.ctx)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got: %v", err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && aggregate == nil {
				t.Error("aggregate must not be nil")
			}
		})
	}
}

func TestOrgDeactivateAggregate(t *testing.T) {
	type res struct {
		isErr func(error) bool
	}
	type args struct {
		aggCreator *es_models.AggregateCreator
		org        *model.Org
		ctx        context.Context
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "correct",
			args: args{
				aggCreator: es_models.NewAggregateCreator("test"),
				ctx:        authz.NewMockContext("org", "user"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "orgID",
						Sequence:    2,
					},
					State: int32(org_model.OrgStateActive),
				},
			},
		},
		{
			name: "already inactive error",
			args: args{
				aggCreator: es_models.NewAggregateCreator("test"),
				ctx:        authz.NewMockContext("org", "user"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "orgID",
						Sequence:    2,
					},
					State: int32(org_model.OrgStateInactive),
				},
			},
			res: res{
				isErr: errors.IsErrorInvalidArgument,
			},
		},
		{
			name: "org nil error",
			args: args{
				aggCreator: es_models.NewAggregateCreator("test"),
				ctx:        authz.NewMockContext("org", "user"),
				org:        nil,
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregateCreator := orgDeactivateAggregate(tt.args.aggCreator, tt.args.org)
			aggregate, err := aggregateCreator(tt.args.ctx)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got: %v", err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && aggregate == nil {
				t.Error("aggregate must not be nil")
			}
		})
	}
}

func TestOrgUpdateAggregates(t *testing.T) {
	type res struct {
		aggregateCount int
		isErr          func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		existing   *model.Org
		updated    *model.Org
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no existing org error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				existing:   nil,
				updated:    &model.Org{},
			},
			res: res{
				aggregateCount: 0,
				isErr:          errors.IsPreconditionFailed,
			},
		},
		{
			name: "no updated org error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				existing:   &model.Org{},
				updated:    nil,
			},
			res: res{
				aggregateCount: 0,
				isErr:          errors.IsPreconditionFailed,
			},
		},
		{
			name: "no changes",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				existing:   &model.Org{},
				updated:    &model.Org{},
			},
			res: res{
				aggregateCount: 0,
				isErr:          errors.IsPreconditionFailed,
			},
		},
		{
			name: "name changed",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				existing: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
					Name: "coas",
				},
				updated: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
					Name: "caos",
				},
			},
			res: res{
				aggregateCount: 3,
				isErr:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrgUpdateAggregates(tt.args.ctx, tt.args.aggCreator, tt.args.existing, tt.args.updated)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got: %v", err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && len(got) != tt.res.aggregateCount {
				t.Errorf("OrgUpdateAggregates() aggregate count = %d, wanted count %d", len(got), tt.res.aggregateCount)
			}
		})
	}
}

func TestOrgCreatedAggregates(t *testing.T) {
	type res struct {
		aggregateCount int
		isErr          func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		org        *model.Org
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no org error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org:        nil,
			},
			res: res{
				aggregateCount: 0,
				isErr:          errors.IsPreconditionFailed,
			},
		},
		{
			name: "org successful",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
					Name: "caos",
				},
			},
			res: res{
				aggregateCount: 2,
				isErr:          nil,
			},
		},
		{
			name: "org with domain successful",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
					Name: "caos",
					Domains: []*model.OrgDomain{{
						Domain: "caos.ch",
					}},
				},
			},
			res: res{
				aggregateCount: 2,
				isErr:          nil,
			},
		},
		{
			name: "no name error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
				},
			},
			res: res{
				aggregateCount: 2,
				isErr:          errors.IsPreconditionFailed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orgCreatedAggregates(tt.args.ctx, tt.args.aggCreator, tt.args.org)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got %T: %v", err, err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && len(got) != tt.res.aggregateCount {
				t.Errorf("OrgUpdateAggregates() aggregate count = %d, wanted count %d", len(got), tt.res.aggregateCount)
			}
		})
	}
}

func TestOrgDomainAddedAggregates(t *testing.T) {
	type res struct {
		eventCount int
		eventType  es_models.EventType
		isErr      func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		org        *model.Org
		domain     *model.OrgDomain
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no domain error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
		{
			name: "domain successful",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
				},
				domain: &model.OrgDomain{
					Domain: "caos.ch",
				},
			},
			res: res{
				eventCount: 1,
				eventType:  model.OrgDomainAdded,
				isErr:      nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := OrgDomainAddedAggregate(tt.args.aggCreator, tt.args.org, tt.args.domain)
			got, err := agg(tt.args.ctx)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got %T: %v", err, err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}

			if tt.res.isErr == nil && got.Events[0].Type != tt.res.eventType {
				t.Errorf("OrgDomainAddedAggregate() event type = %v, wanted count %v", got.Events[0].Type, tt.res.eventType)
			}
			if tt.res.isErr == nil && len(got.Events) != tt.res.eventCount {
				t.Errorf("OrgDomainAddedAggregate() event count = %v, wanted count %v", len(got.Events), tt.res.eventCount)
			}
		})
	}
}

func TestOrgDomainVerifiedAggregates(t *testing.T) {
	type res struct {
		aggregateCount int
		isErr          func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		org        *model.Org
		domain     *model.OrgDomain
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no domain error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
		{
			name: "domain successful",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
				},
				domain: &model.OrgDomain{
					Domain: "caos.ch",
				},
			},
			res: res{
				aggregateCount: 2,
				isErr:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := OrgDomainVerifiedAggregate(tt.args.aggCreator, tt.args.org, tt.args.domain)
			got, err := agg(tt.args.ctx)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got %T: %v", err, err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && len(got) != tt.res.aggregateCount {
				t.Errorf("OrgDomainVerifiedAggregate() aggregate count = %d, wanted count %d", len(got), tt.res.aggregateCount)
			}
		})
	}
}

func TestOrgDomainSetPrimaryAggregates(t *testing.T) {
	type res struct {
		eventsCount int
		eventType   es_models.EventType
		isErr       func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		org        *model.Org
		domain     *model.OrgDomain
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no domain error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
		{
			name: "domain successful",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
				},
				domain: &model.OrgDomain{
					Domain: "caos.ch",
				},
			},
			res: res{
				eventsCount: 1,
				eventType:   model.OrgDomainPrimarySet,
				isErr:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := OrgDomainSetPrimaryAggregate(tt.args.aggCreator, tt.args.org, tt.args.domain)
			got, err := agg(tt.args.ctx)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got %T: %v", err, err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && got.Events[0].Type != tt.res.eventType {
				t.Errorf("OrgDomainSetPrimaryAggregate() event type = %v, wanted count %v", got.Events[0].Type, tt.res.eventType)
			}
			if tt.res.isErr == nil && len(got.Events) != tt.res.eventsCount {
				t.Errorf("OrgDomainSetPrimaryAggregate() event count = %d, wanted count %d", len(got.Events), tt.res.eventsCount)
			}
		})
	}
}

func TestOrgDomainRemovedAggregates(t *testing.T) {
	type res struct {
		aggregateCount int
		isErr          func(error) bool
	}
	type args struct {
		ctx        context.Context
		aggCreator *es_models.AggregateCreator
		org        *model.Org
		domain     *model.OrgDomain
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no domain error",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
			},
			res: res{
				aggregateCount: 0,
				isErr:          errors.IsPreconditionFailed,
			},
		},
		{
			name: "domain successful",
			args: args{
				ctx:        authz.NewMockContext("org", "user"),
				aggCreator: es_models.NewAggregateCreator("test"),
				org: &model.Org{
					ObjectRoot: es_models.ObjectRoot{
						AggregateID: "sdaf",
						Sequence:    5,
					},
				},
				domain: &model.OrgDomain{
					Domain: "caos.ch",
				},
			},
			res: res{
				aggregateCount: 2,
				isErr:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrgDomainRemovedAggregate(tt.args.ctx, tt.args.aggCreator, tt.args.org, tt.args.domain)
			if tt.res.isErr == nil && err != nil {
				t.Errorf("no error expected got %T: %v", err, err)
			}
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error got %T: %v", err, err)
			}
			if tt.res.isErr == nil && len(got) != tt.res.aggregateCount {
				t.Errorf("OrgDomainRemovedAggregate() aggregate count = %d, wanted count %d", len(got), tt.res.aggregateCount)
			}
		})
	}
}
