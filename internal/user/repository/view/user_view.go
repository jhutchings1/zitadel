package view

import (
	"github.com/caos/zitadel/internal/view/repository"
	"github.com/jinzhu/gorm"

	caos_errs "github.com/caos/zitadel/internal/errors"
	global_model "github.com/caos/zitadel/internal/model"
	usr_model "github.com/caos/zitadel/internal/user/model"
	"github.com/caos/zitadel/internal/user/repository/view/model"
)

func UserByID(db *gorm.DB, table, userID string) (*model.UserView, error) {
	user := new(model.UserView)
	query := repository.PrepareGetByKey(table, model.UserSearchKey(usr_model.UserSearchKeyUserID), userID)
	err := query(db, user)
	if caos_errs.IsNotFound(err) {
		return nil, caos_errs.ThrowNotFound(nil, "VIEW-sj8Sw", "Errors.User.NotFound")
	}
	return user, err
}

func UserByUserName(db *gorm.DB, table, userName string) (*model.UserView, error) {
	user := new(model.UserView)
	query := repository.PrepareGetByKey(table, model.UserSearchKey(usr_model.UserSearchKeyUserName), userName)
	err := query(db, user)
	if caos_errs.IsNotFound(err) {
		return nil, caos_errs.ThrowNotFound(nil, "VIEW-Lso9s", "Errors.User.NotFound")
	}
	return user, err
}

func UserByLoginName(db *gorm.DB, table, loginName string) (*model.UserView, error) {
	user := new(model.UserView)
	loginNameQuery := &model.UserSearchQuery{
		Key:    usr_model.UserSearchKeyLoginNames,
		Method: global_model.SearchMethodListContains,
		Value:  loginName,
	}
	query := repository.PrepareGetByQuery(table, loginNameQuery)
	err := query(db, user)
	if caos_errs.IsNotFound(err) {
		return nil, caos_errs.ThrowNotFound(nil, "VIEW-AD4qs", "Errors.User.NotFound")
	}
	return user, err
}

func UsersByOrgID(db *gorm.DB, table, orgID string) ([]*model.UserView, error) {
	users := make([]*model.UserView, 0)
	orgIDQuery := &usr_model.UserSearchQuery{
		Key:    usr_model.UserSearchKeyResourceOwner,
		Method: global_model.SearchMethodEquals,
		Value:  orgID,
	}
	query := repository.PrepareSearchQuery(table, model.UserSearchRequest{
		Queries: []*usr_model.UserSearchQuery{orgIDQuery},
	})
	_, err := query(db, &users)
	return users, err
}

func SearchUsers(db *gorm.DB, table string, req *usr_model.UserSearchRequest) ([]*model.UserView, int, error) {
	users := make([]*model.UserView, 0)
	query := repository.PrepareSearchQuery(table, model.UserSearchRequest{Limit: req.Limit, Offset: req.Offset, Queries: req.Queries})
	count, err := query(db, &users)
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

func GetGlobalUserByEmail(db *gorm.DB, table, email string) (*model.UserView, error) {
	user := new(model.UserView)
	query := repository.PrepareGetByKey(table, model.UserSearchKey(usr_model.UserSearchKeyEmail), email)
	err := query(db, user)
	if caos_errs.IsNotFound(err) {
		return nil, caos_errs.ThrowNotFound(nil, "VIEW-8uWer", "Errors.User.NotFound")
	}
	return user, err
}

func IsUserUnique(db *gorm.DB, table, userName, email string) (bool, error) {
	user := new(model.UserView)
	query := repository.PrepareGetByKey(table, model.UserSearchKey(usr_model.UserSearchKeyEmail), email)
	err := query(db, user)
	if err != nil && !caos_errs.IsNotFound(err) {
		return false, err
	}
	if user != nil {
		return false, nil
	}
	query = repository.PrepareGetByKey(table, model.UserSearchKey(usr_model.UserSearchKeyUserName), email)
	err = query(db, user)
	if err != nil && !caos_errs.IsNotFound(err) {
		return false, err
	}
	return user == nil, nil
}

func UserMfas(db *gorm.DB, table, userID string) ([]*usr_model.MultiFactor, error) {
	user, err := UserByID(db, table, userID)
	if err != nil {
		return nil, err
	}
	if user.OTPState == int32(usr_model.MfaStateUnspecified) {
		return []*usr_model.MultiFactor{}, nil
	}
	return []*usr_model.MultiFactor{{Type: usr_model.MfaTypeOTP, State: usr_model.MfaState(user.OTPState)}}, nil
}

func PutUsers(db *gorm.DB, table string, users ...*model.UserView) error {
	save := repository.PrepareBulkSave(table)
	u := make([]interface{}, len(users))
	for i, user := range users {
		u[i] = user
	}
	return save(db, u...)
}

func PutUser(db *gorm.DB, table string, user *model.UserView) error {
	save := repository.PrepareSave(table)
	return save(db, user)
}

func DeleteUser(db *gorm.DB, table, userID string) error {
	delete := repository.PrepareDeleteByKey(table, model.UserSearchKey(usr_model.UserSearchKeyUserID), userID)
	return delete(db)
}
