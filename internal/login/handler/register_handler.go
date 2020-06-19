package handler

import (
	"github.com/caos/zitadel/internal/auth_request/model"
	caos_errs "github.com/caos/zitadel/internal/errors"
	"github.com/caos/zitadel/internal/eventstore/models"
	org_model "github.com/caos/zitadel/internal/org/model"
	usr_model "github.com/caos/zitadel/internal/user/model"
	"golang.org/x/text/language"
	"net/http"
)

const (
	tmplRegister          = "register"
	orgProjectCreatorRole = "ORG_PROJECT_CREATOR"
)

type registerFormData struct {
	Email     string `schema:"email"`
	Firstname string `schema:"firstname"`
	Lastname  string `schema:"lastname"`
	Language  string `schema:"language"`
	Gender    int32  `schema:"gender"`
	Password  string `schema:"password"`
	Password2 string `schema:"password2"`
}

type registerData struct {
	baseData
	registerFormData
}

func (l *Login) handleRegister(w http.ResponseWriter, r *http.Request) {
	data := new(registerFormData)
	authRequest, err := l.getAuthRequestAndParseData(r, data)
	if err != nil {
		l.renderError(w, r, authRequest, err)
		return
	}
	l.renderRegister(w, r, authRequest, data, nil)
}

func (l *Login) handleRegisterCheck(w http.ResponseWriter, r *http.Request) {
	data := new(registerFormData)
	authRequest, err := l.getAuthRequestAndParseData(r, data)
	if err != nil {
		l.renderError(w, r, authRequest, err)
		return
	}
	if data.Password != data.Password2 {
		err := caos_errs.ThrowInvalidArgument(nil, "VIEW-KaGue", "Errors.User.Password.ConfirmationWrong")
		l.renderRegister(w, r, authRequest, data, err)
		return
	}
	iam, err := l.authRepo.GetIam(r.Context())
	if err != nil {
		l.renderRegister(w, r, authRequest, data, err)
		return
	}

	member := &org_model.OrgMember{
		ObjectRoot: models.ObjectRoot{AggregateID: iam.GlobalOrgID},
		Roles:      []string{orgProjectCreatorRole},
	}
	user, err := l.authRepo.Register(setContext(r.Context(), iam.GlobalOrgID), data.toUserModel(), member, iam.GlobalOrgID)
	if err != nil {
		l.renderRegister(w, r, authRequest, data, err)
		return
	}
	if authRequest == nil {
		http.Redirect(w, r, l.zitadelURL, http.StatusFound)
		return
	}
	authRequest.LoginName = user.PreferredLoginName
	l.renderNextStep(w, r, authRequest)
}

func (l *Login) renderRegister(w http.ResponseWriter, r *http.Request, authRequest *model.AuthRequest, formData *registerFormData, err error) {
	var errType, errMessage string
	if err != nil {
		errMessage = l.getErrorMessage(r, err)
	}
	if formData == nil {
		formData = new(registerFormData)
	}
	if formData.Language == "" {
		formData.Language = l.renderer.Lang(r).String()
	}
	data := registerData{
		baseData:         l.getBaseData(r, authRequest, "Register", errType, errMessage),
		registerFormData: *formData,
	}
	funcs := map[string]interface{}{
		"selectedLanguage": func(l string) bool {
			if formData == nil {
				return false
			}
			return formData.Language == l
		},
		"selectedGender": func(g int32) bool {
			if formData == nil {
				return false
			}
			return formData.Gender == g
		},
	}
	l.renderer.RenderTemplate(w, r, l.renderer.Templates[tmplRegister], data, funcs)
}

func (d registerFormData) toUserModel() *usr_model.User {
	return &usr_model.User{
		Profile: &usr_model.Profile{
			FirstName:         d.Firstname,
			LastName:          d.Lastname,
			PreferredLanguage: language.Make(d.Language),
			Gender:            usr_model.Gender(d.Gender),
		},
		Password: &usr_model.Password{
			SecretString: d.Password,
		},
		Email: &usr_model.Email{
			EmailAddress: d.Email,
		},
	}
}