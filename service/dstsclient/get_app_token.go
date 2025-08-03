package dstsclient

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	pb "github.com/hpinc/krypton-scheduler/protos/dstsprotos"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

const (
	clientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

type AssertionClaims struct {
	// Standard JWT claims such as 'aud', 'exp', 'jti', 'iat', 'iss', 'nbf',
	// 'sub'
	jwt.RegisteredClaims

	// Nonce - the challenge returned by the device STS at the challenge
	// endpoint. This nonce is included in the signed assertion to protect
	// against assertion replay attacks.
	Nonce string `json:"nonce"`
}

func getAppToken(ctx context.Context) error {
	// Retrieve an application authentication challenge from the DSTS.
	response, err := gClient.GetAppAuthenticationChallenge(ctx,
		&pb.AppAuthenticationChallengeRequest{
			Header:  newDstsProtocolHeader(),
			Version: dstsProtocolVersion,
			AppId:   dstsConfig.SchedulerAppId,
		})
	if err != nil {
		schedLogger.Error("Failed to get the app authentication challenge!",
			zap.Error(err),
		)
		return err
	}

	if response.Header.Status != uint32(codes.OK) {
		schedLogger.Error("Failed to get the app authentication challenge. RPC failed!",
			zap.Uint32("Status code:", response.Header.Status),
			zap.Error(err),
		)
		return err
	}

	// Construct a JWT assertion and sign it with the app private key.
	claims := AssertionClaims{
		Nonce: response.Challenge,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    dstsConfig.SchedulerAppId,
			Subject:   dstsConfig.SchedulerAppId,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}
	assertionToken := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	assertion, err := assertionToken.SignedString(dstsConfig.PrivateKey)
	if err != nil {
		schedLogger.Error("Failed to generate signed client assertion.",
			zap.Error(err))
		return err
	}

	// Complete app authentication.
	authResponse, err := gClient.AuthenticateApp(gCtx, &pb.AppAuthenticationRequest{
		Header:        newDstsProtocolHeader(),
		Version:       dstsProtocolVersion,
		AppId:         dstsConfig.SchedulerAppId,
		AssertionType: clientAssertionType,
		Assertion:     assertion,
	})
	if err != nil {
		schedLogger.Error("AuthenticateApp RPC failed",
			zap.Error(err))
		return err
	}

	if authResponse.Header.Status != uint32(codes.OK) {
		schedLogger.Error("Failed to get the app access token. RPC failed!",
			zap.Uint32("Status code:", authResponse.Header.Status),
			zap.Error(err),
		)
		return err
	}

	gAppToken = authResponse.AccessToken
	gAppTokenExpiresAt = authResponse.ExpiresAt.AsTime()
	return nil
}
