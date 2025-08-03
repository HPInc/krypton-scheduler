package rest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hpinc/krypton-scheduler/service/dstsclient"
	"go.uber.org/zap"
)

var (
	ErrNoAuthorizationHeader  = errors.New("request does not have an authorization header")
	ErrNoBearerTokenSpecified = errors.New("authorization header does not contain a bearer token")
)

func isValidAppAccessToken(r *http.Request) error {
	if !appTokenAuthnEnabled {
		return nil
	}

	tokenString := r.Header.Get(headerAuthorization)
	if tokenString == "" {
		schedLogger.Error("Authorization header was not provided in the request")
		return ErrNoAuthorizationHeader
	}
	if !strings.HasPrefix(tokenString, bearerToken) {
		schedLogger.Error("Authorization header specified does not contain a bearer token!")
		return ErrNoBearerTokenSpecified
	}

	_, err := dstsclient.ValidateAppAccessToken(strings.TrimPrefix(tokenString, bearerToken))
	if err != nil {
		schedLogger.Error("Provided access token is not a valid app access token!",
			zap.Error(err),
		)
		return err
	}

	return nil
}
