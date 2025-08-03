package dstsclient

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	pb "github.com/hpinc/krypton-scheduler/protos/dstsprotos"
	"go.uber.org/zap"
)

const (
	// ktyRSA is the key type (kty) in the JWT header for RSA.
	ktyRSA       = "RSA"
	rsaPublicKey = "RSA PUBLIC KEY"
)

var signingKeyTable map[string]*rsa.PublicKey

func getSigningKey(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		schedLogger.Error("Invalid kid in the token header!")
		return nil, ErrInvalidTokenHeaderKid
	}

	// Check if a signing key corresponding to the kid was found in the
	// signing key table.
	pubKey, ok := signingKeyTable[kid]
	if !ok {
		// Key with this kid was not found - fetch the JWKS keys from the
		// DSTS to check if this is a new signing key.
		err := getJWKSSigningKey()
		if err != nil {
			schedLogger.Error("Failed to get JWKS signing keys from DSTS!",
				zap.String("Token signed by:", kid),
				zap.Error(err),
			)
			return nil, err
		}

		pubKey, ok = signingKeyTable[kid]
		if !ok {
			return nil, fmt.Errorf("no public key to validate kid: %s", kid)
		}
	}

	return pubKey, nil
}

// parseRSASigningKey parses a JWK and turns it into an RSA public key.
func parseRSASigningKey(j *pb.JSONWebKey) (publicKey *rsa.PublicKey, err error) {
	if j.E == "" || j.N == "" {
		return nil, ErrMissingAssets
	}

	// Decode the exponent from Base64.
	// According to RFC 7518, this is a Base64 URL unsigned integer.
	// https://tools.ietf.org/html/rfc7518#section-6.3
	exponent, err := base64urlTrailingPadding(j.E)
	if err != nil {
		return nil, err
	}

	modulus, err := base64urlTrailingPadding(j.N)
	if err != nil {
		return nil, err
	}

	exp := big.NewInt(0).SetBytes(exponent).Uint64()
	if exp > math.MaxInt {
		return nil, ErrOverflow
	}

	// Turn the exponent into an integer.
	// According to RFC 7517, these numbers are in big-endian format.
	// https://tools.ietf.org/html/rfc7517#appendix-A.1
	return &rsa.PublicKey{
		E: int(exp),
		N: big.NewInt(0).SetBytes(modulus),
	}, nil
}

// base64urlTrailingPadding removes trailing padding before decoding a string from base64url. Some non-RFC compliant
// JWKS contain padding at the end values for base64url encoded public keys.
//
// Trailing padding is required to be removed from base64url encoded keys.
// RFC 7517 defines base64url the same as RFC 7515 Section 2:
// https://datatracker.ietf.org/doc/html/rfc7517#section-1.1
// https://datatracker.ietf.org/doc/html/rfc7515#section-2
func base64urlTrailingPadding(s string) ([]byte, error) {
	s = strings.TrimRight(s, "=")
	return base64.RawURLEncoding.DecodeString(s)
}
