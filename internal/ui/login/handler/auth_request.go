package handler

import (
	"net/http"

	"github.com/caos/zitadel/internal/auth_request/model"
)

const (
	queryAuthRequestID = "authRequestID"
)

func (l *Login) getAuthRequest(r *http.Request) (*model.AuthRequest, error) {
	authRequestID := r.FormValue(queryAuthRequestID)
	if authRequestID == "" {
		return nil, nil
	}
	return l.authRepo.AuthRequestByID(r.Context(), authRequestID)
}

func (l *Login) getAuthRequestAndParseData(r *http.Request, data interface{}) (*model.AuthRequest, error) {
	authReq, err := l.getAuthRequest(r)
	if err != nil {
		return nil, err
	}
	err = l.parser.Parse(r, data)
	return authReq, err
}
