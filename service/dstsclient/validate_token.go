package dstsclient

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

const (
	dstsIssuerName = "HP Device Token Service"

	tokenTypeDevice = "device"
	tokenTypeApp    = "app"
)

type DstsTokenClaims struct {
	// Standard JWT claims such as 'aud', 'exp', 'jti', 'iat', 'iss', 'nbf',
	// 'sub'
	// 'sub' claim is set to the unique ID assigned to the device after enrollment.
	jwt.RegisteredClaims

	// Type of token. Possible values are:
	//  - device: device access tokens
	//  - app: app access token
	TokenType string `json:"typ"`

	// The ID of the tenant to which the device belongs.
	TenantID string `json:"tid"`

	// The device management service responsible for managing this device.
	ManagementService string `json:"ms"`
}

func ValidateDeviceAccessToken(accessToken string) (*DstsTokenClaims, error) {
	claims, err := parseAndValidateCommonClaims(accessToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != tokenTypeDevice {
		return nil, ErrNotDeviceToken
	}

	return claims, nil
}

func ValidateAppAccessToken(accessToken string) (*DstsTokenClaims, error) {
	claims, err := parseAndValidateCommonClaims(accessToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != tokenTypeApp {
		return nil, ErrNotAppToken
	}

	return claims, nil
}

func parseAndValidateCommonClaims(accessToken string) (*DstsTokenClaims, error) {
	var claims DstsTokenClaims

	token, err := jwt.ParseWithClaims(accessToken, &claims, getSigningKey)
	if err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, ErrInvalidToken
	}

	if token.Header["alg"] == nil {
		return nil, ErrInvalidTokenHeaderSigningAlg
	}

	if !strings.HasPrefix(claims.Issuer, dstsIssuerName) {
		return nil, ErrInvalidIssuerClaim
	}
	return &claims, nil
}
