package eventsourcing

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/caos/zitadel/internal/api/authz"
	caos_errs "github.com/caos/zitadel/internal/errors"
	es_models "github.com/caos/zitadel/internal/eventstore/models"
	iam_model "github.com/caos/zitadel/internal/iam/model"
	"github.com/caos/zitadel/internal/iam/repository/eventsourcing/model"
)

func TestIamByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es  *IamEventstore
		iam *model.Iam
	}
	type res struct {
		iam     *model.Iam
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "iam from events, ok",
			args: args{
				es:  GetMockIamByIDOK(ctrl),
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}},
			},
			res: res{
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}},
			},
		},
		{
			name: "iam from events, no events",
			args: args{
				es:  GetMockIamByIDNoEvents(ctrl),
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}},
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
		{
			name: "iam from events, no id",
			args: args{
				es:  GetMockIamByIDNoEvents(ctrl),
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "", Sequence: 1}},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.IamByID(nil, tt.args.iam.AggregateID)

			if tt.res.errFunc == nil && result.AggregateID != tt.res.iam.AggregateID {
				t.Errorf("got wrong result name: expected: %v, actual: %v ", tt.res.iam.AggregateID, result.AggregateID)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestSetUpStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es    *IamEventstore
		ctx   context.Context
		iamID string
	}
	type res struct {
		iam     *model.Iam
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "setup started iam, ok",
			args: args{
				es:    GetMockManipulateIamNotExisting(ctrl),
				ctx:   authz.NewMockContext("orgID", "userID"),
				iamID: "iamID",
			},
			res: res{
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "iamID", Sequence: 1}, SetUpStarted: true},
			},
		},
		{
			name: "setup already started",
			args: args{
				es:    GetMockManipulateIam(ctrl),
				ctx:   authz.NewMockContext("orgID", "userID"),
				iamID: "iamID",
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "setup iam no id",
			args: args{
				es:  GetMockManipulateIam(ctrl),
				ctx: authz.NewMockContext("orgID", "userID"),
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.StartSetup(tt.args.ctx, tt.args.iamID)

			if tt.res.errFunc == nil && result.AggregateID == "" {
				t.Errorf("result has no id")
			}
			if tt.res.errFunc == nil && result.SetUpStarted != tt.res.iam.SetUpStarted {
				t.Errorf("got wrong result setupStarted: expected: %v, actual: %v ", tt.res.iam.SetUpStarted, result.SetUpStarted)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestSetUpDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es    *IamEventstore
		ctx   context.Context
		iamID string
	}
	type res struct {
		iam     *model.Iam
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "setup done iam, ok",
			args: args{
				es:    GetMockManipulateIam(ctrl),
				ctx:   authz.NewMockContext("orgID", "userID"),
				iamID: "iamID",
			},
			res: res{
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "iamID", Sequence: 1}, SetUpStarted: true, SetUpDone: true},
			},
		},
		{
			name: "setup iam no id",
			args: args{
				es:  GetMockManipulateIam(ctrl),
				ctx: authz.NewMockContext("orgID", "userID"),
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "iam not found",
			args: args{
				es:    GetMockManipulateIamNotExisting(ctrl),
				ctx:   authz.NewMockContext("orgID", "userID"),
				iamID: "iamID",
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.SetupDone(tt.args.ctx, tt.args.iamID)

			if tt.res.errFunc == nil && result.AggregateID == "" {
				t.Errorf("result has no id")
			}
			if tt.res.errFunc == nil && result.SetUpDone != tt.res.iam.SetUpDone {
				t.Errorf("got wrong result SetUpDone: expected: %v, actual: %v ", tt.res.iam.SetUpDone, result.SetUpDone)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestSetGlobalOrg(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es        *IamEventstore
		ctx       context.Context
		iamID     string
		globalOrg string
	}
	type res struct {
		iam     *model.Iam
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "global org set, ok",
			args: args{
				es:        GetMockManipulateIam(ctrl),
				ctx:       authz.NewMockContext("orgID", "userID"),
				iamID:     "iamID",
				globalOrg: "globalOrg",
			},
			res: res{
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "iamID", Sequence: 1}, SetUpStarted: true, GlobalOrgID: "globalOrg"},
			},
		},
		{
			name: "no iam id",
			args: args{
				es:        GetMockManipulateIam(ctrl),
				ctx:       authz.NewMockContext("orgID", "userID"),
				globalOrg: "",
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "no global org",
			args: args{
				es:    GetMockManipulateIam(ctrl),
				ctx:   authz.NewMockContext("orgID", "userID"),
				iamID: "iamID",
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "iam not found",
			args: args{
				es:        GetMockManipulateIamNotExisting(ctrl),
				ctx:       authz.NewMockContext("orgID", "userID"),
				iamID:     "iamID",
				globalOrg: "globalOrg",
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.SetGlobalOrg(tt.args.ctx, tt.args.iamID, tt.args.globalOrg)

			if tt.res.errFunc == nil && result.AggregateID == "" {
				t.Errorf("result has no id")
			}
			if tt.res.errFunc == nil && result.GlobalOrgID != tt.res.iam.GlobalOrgID {
				t.Errorf("got wrong result GlobalOrgID: expected: %v, actual: %v ", tt.res.iam.GlobalOrgID, result.GlobalOrgID)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestSetIamProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es           *IamEventstore
		ctx          context.Context
		iamID        string
		iamProjectID string
	}
	type res struct {
		iam     *model.Iam
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "iam project set, ok",
			args: args{
				es:           GetMockManipulateIam(ctrl),
				ctx:          authz.NewMockContext("orgID", "userID"),
				iamID:        "iamID",
				iamProjectID: "iamProjectID",
			},
			res: res{
				iam: &model.Iam{ObjectRoot: es_models.ObjectRoot{AggregateID: "iamID", Sequence: 1}, SetUpStarted: true, IamProjectID: "iamProjectID"},
			},
		},
		{
			name: "no iam id",
			args: args{
				es:           GetMockManipulateIam(ctrl),
				ctx:          authz.NewMockContext("orgID", "userID"),
				iamProjectID: "",
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "no global org",
			args: args{
				es:    GetMockManipulateIam(ctrl),
				ctx:   authz.NewMockContext("orgID", "userID"),
				iamID: "iamID",
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "iam not found",
			args: args{
				es:           GetMockManipulateIamNotExisting(ctrl),
				ctx:          authz.NewMockContext("orgID", "userID"),
				iamID:        "iamID",
				iamProjectID: "iamProjectID",
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.SetIamProject(tt.args.ctx, tt.args.iamID, tt.args.iamProjectID)

			if tt.res.errFunc == nil && result.AggregateID == "" {
				t.Errorf("result has no id")
			}
			if tt.res.errFunc == nil && result.IamProjectID != tt.res.iam.IamProjectID {
				t.Errorf("got wrong result IamProjectID: expected: %v, actual: %v ", tt.res.iam.IamProjectID, result.IamProjectID)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestAddIamMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es     *IamEventstore
		ctx    context.Context
		member *iam_model.IamMember
	}
	type res struct {
		result  *iam_model.IamMember
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "add iam member, ok",
			args: args{
				es:     GetMockManipulateIam(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
			res: res{
				result: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
		},
		{
			name: "no userid",
			args: args{
				es:     GetMockManipulateIam(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, Roles: []string{"Roles"}},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "no roles",
			args: args{
				es:     GetMockManipulateIam(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID"},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "member already existing",
			args: args{
				es:     GetMockManipulateIamWithMember(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
			res: res{
				errFunc: caos_errs.IsErrorAlreadyExists,
			},
		},
		{
			name: "existing iam not found",
			args: args{
				es:     GetMockManipulateIamNotExisting(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.AddIamMember(tt.args.ctx, tt.args.member)

			if tt.res.errFunc == nil && result.AggregateID == "" {
				t.Errorf("result has no id")
			}
			if tt.res.errFunc == nil && result.UserID != tt.res.result.UserID {
				t.Errorf("got wrong result userid: expected: %v, actual: %v ", tt.res.result.UserID, result.UserID)
			}
			if tt.res.errFunc == nil && len(result.Roles) != len(tt.res.result.Roles) {
				t.Errorf("got wrong result roles: expected: %v, actual: %v ", tt.res.result.Roles, result.Roles)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestChangeIamMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es     *IamEventstore
		ctx    context.Context
		member *iam_model.IamMember
	}
	type res struct {
		result  *iam_model.IamMember
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "add iam member, ok",
			args: args{
				es:     GetMockManipulateIamWithMember(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"ChangeRoles"}},
			},
			res: res{
				result: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
		},
		{
			name: "no userid",
			args: args{
				es:     GetMockManipulateIam(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, Roles: []string{"ChangeRoles"}},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "no roles",
			args: args{
				es:     GetMockManipulateIam(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID"},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "member not existing",
			args: args{
				es:     GetMockManipulateIam(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "existing not found",
			args: args{
				es:     GetMockManipulateIamNotExisting(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"ChangeRoles"}},
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.args.es.ChangeIamMember(tt.args.ctx, tt.args.member)

			if tt.res.errFunc == nil && result.AggregateID == "" {
				t.Errorf("result has no id")
			}
			if tt.res.errFunc == nil && result.UserID != tt.res.result.UserID {
				t.Errorf("got wrong result userid: expected: %v, actual: %v ", tt.res.result.UserID, result.UserID)
			}
			if tt.res.errFunc == nil && len(result.Roles) != len(tt.res.result.Roles) {
				t.Errorf("got wrong result roles: expected: %v, actual: %v ", tt.res.result.Roles, result.Roles)
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestRemoveIamMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		es       *IamEventstore
		ctx      context.Context
		existing *model.Iam
		member   *iam_model.IamMember
	}
	type res struct {
		result  *iam_model.IamMember
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "remove iam member, ok",
			args: args{
				es:  GetMockManipulateIamWithMember(ctrl),
				ctx: authz.NewMockContext("orgID", "userID"),
				existing: &model.Iam{
					ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1},
					Members:    []*model.IamMember{{UserID: "UserID", Roles: []string{"Roles"}}},
				},
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID"},
			},
			res: res{
				result: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
		},
		{
			name: "no userid",
			args: args{
				es:  GetMockManipulateIam(ctrl),
				ctx: authz.NewMockContext("orgID", "userID"),
				existing: &model.Iam{
					ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1},
					Members:    []*model.IamMember{{UserID: "UserID", Roles: []string{"Roles"}}},
				},
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, Roles: []string{"ChangeRoles"}},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "member not existing",
			args: args{
				es:  GetMockManipulateIam(ctrl),
				ctx: authz.NewMockContext("orgID", "userID"),
				existing: &model.Iam{
					ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1},
				},
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"Roles"}},
			},
			res: res{
				errFunc: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "existing not found",
			args: args{
				es:     GetMockManipulateIamNotExisting(ctrl),
				ctx:    authz.NewMockContext("orgID", "userID"),
				member: &iam_model.IamMember{ObjectRoot: es_models.ObjectRoot{AggregateID: "AggregateID", Sequence: 1}, UserID: "UserID", Roles: []string{"ChangeRoles"}},
			},
			res: res{
				errFunc: caos_errs.IsNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.es.RemoveIamMember(tt.args.ctx, tt.args.member)

			if tt.res.errFunc == nil && err != nil {
				t.Errorf("should not get err")
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}
