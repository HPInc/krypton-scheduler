package dstsclient

import (
	"errors"
	"net/http"
)

var (
	ErrMissingAssets                = errors.New("required assets are missing to create a public key")
	ErrInvalidToken                 = errors.New("invalid token provided")
	ErrInvalidTokenHeaderKid        = errors.New("invalid token signing kid specified")
	ErrInvalidTokenHeaderSigningAlg = errors.New("invalid token signing algorithm specified")
	ErrInvalidIssuerClaim           = errors.New("specified token contains an invalid issuer claim")
	ErrInvalidAudienceClaim         = errors.New("specified token contains an invalid audience claim")
	ErrNotDeviceToken               = errors.New("specified token is not a device access token")
	ErrNotAppToken                  = errors.New("specified token is not an app access token")
	ErrUnauthorized                 = errors.New(http.StatusText(http.StatusUnauthorized))
	ErrBadRequest                   = errors.New(http.StatusText(http.StatusBadRequest))
	ErrOverflow                     = errors.New("integer overflow detected")
)
