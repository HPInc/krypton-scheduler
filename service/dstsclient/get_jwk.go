package dstsclient

import (
	"context"
	"crypto/rsa"

	pb "github.com/hpinc/krypton-scheduler/protos/dstsprotos"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// Retrieve the token signing keys in JWKS format from the DSTS JWKS endpoint.
func getJWKSSigningKey() error {

	// Fetch the keys from the DSTS JWKS endpoint.
	ctx, cancelFunc := context.WithTimeout(gCtx, dstsRequestTimeout)
	defer cancelFunc()

	response, err := gClient.GetSigningKey(ctx, &pb.GetSigningKeyRequest{})
	if err != nil {
		schedLogger.Error("Failed to get the JWKS signing key from DSTS!",
			zap.Error(err),
		)
		return err
	}

	if response.Header.Status != uint32(codes.OK) {
		schedLogger.Error("Failed to get the JWKS signing key from DSTS. RPC failed!",
			zap.Uint32("Status code:", response.Header.Status),
			zap.Error(err),
		)
		return err
	}

	signingKeyTable = make(map[string]*rsa.PublicKey, len(response.SigningKey))
	for _, key := range response.SigningKey {
		switch keyType := key.Kty; keyType {
		case ktyRSA:
			publicKey, err := parseRSASigningKey(key)
			if err != nil {
				schedLogger.Error("Error parsing signing key",
					zap.String("type:", key.Kty),
					zap.String("kid:", key.Kid),
					zap.Error(err))
				return err
			}

			signingKeyTable[key.Kid] = publicKey

		default:
			continue
		}
	}

	return nil
}
