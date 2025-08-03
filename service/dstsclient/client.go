package dstsclient

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/google/uuid"
	pb "github.com/hpinc/krypton-scheduler/protos/dstsprotos"
	"github.com/hpinc/krypton-scheduler/service/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	dstsProtocolVersion = "v1"
	dstsConnStr         = "%s:%d"

	dstsRequestTimeout      = time.Second * 3
	dstsOperationRetryCount = 5
	baseRetryDuration       = time.Second * 5
)

var (
	schedLogger        *zap.Logger
	dstsConfig         *config.DstsConfig
	gCtx               context.Context
	gConnection        *grpc.ClientConn
	gClient            pb.DeviceSTSClient
	gAppToken          string
	gAppTokenExpiresAt time.Time
)

func GetAccessToken() (string, error) {
	// If the current app access token has expired, acquire a fresh token from
	// the DSTS.
	if time.Now().After(gAppTokenExpiresAt) {
		tokenOperation := retryWithBackoff(func(ctx context.Context) error {
			return getAppToken(ctx)
		})
		err := tokenOperation(gCtx)
		if err != nil {
			return "", err
		}
	}
	return gAppToken, nil
}

func IsAppTokenExpired() bool {
	return time.Now().After(gAppTokenExpiresAt)
}

// Start - initialize a client connection to the device STS using the DSTS
// configuration settings. Use the Ping RPC to ensure we can connect to the DSTS.
func Start(ctx context.Context, logger *zap.Logger,
	cfgMgr *config.ConfigMgr) error {
	var err error
	schedLogger = logger
	dstsConfig = cfgMgr.GetDstsConfig()

	err = ctx.Err()
	if err != nil {
		schedLogger.Error("Context is no longer valid!",
			zap.Error(err),
		)
		return err
	}
	gCtx = ctx

	pkeyStr := os.Getenv(dstsConfig.PrivateKeyEnv)
	if pkeyStr == "" {
		schedLogger.Error("Specified private key for the scheduler app is invalid!")
		return os.ErrInvalid
	}

	block, _ := pem.Decode([]byte(pkeyStr))
	if block == nil {
		schedLogger.Error("Failed to decode specified private key for the scheduler app is invalid!")
		return os.ErrInvalid
	}

	pkey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		schedLogger.Error("Failed to parse the scheduler private key!",
			zap.Error(err),
		)
		return err
	}
	dstsConfig.PrivateKey = pkey.(*rsa.PrivateKey)

	// Create a connection to the DSTS service.
	addr := fmt.Sprintf(dstsConnStr, dstsConfig.Host, dstsConfig.RpcPort)
	err = connectWithRetry(addr)
	if err != nil {
		schedLogger.Error("Failed to connect to the DSTS!",
			zap.Error(err))
		return err
	}

	schedLogger.Debug("Successfully created a connection to the DSTS!",
		zap.String("Address: ", addr))
	gClient = pb.NewDeviceSTSClient(gConnection)

	// Ping the DSTS to ensure connectivity.
	err = pingWithRetry()
	if err != nil {
		return err
	}

	// Get an app access token for the scheduler app. This will be used to
	// connect to the MQTT broker (IoT core).
	_, err = GetAccessToken()
	if err != nil {
		schedLogger.Error("Failed to retrieve an app access token from DSTS for the scheduler!",
			zap.Error(err),
		)
		return err
	}

	schedLogger.Info("Retrieved an app access token from DSTS")
	return nil
}

func connectWithRetry(addr string) error {
	var err error
	connectOperation := retryWithBackoff(func(ctx context.Context) error {
		gConnection, err = grpc.Dial(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		return err
	})

	return connectOperation(gCtx)
}

func pingWithRetry() error {
	pingOperation := retryWithBackoff(func(ctx context.Context) error {
		// Invoke the Ping RPC on the DSTS to ensure we can connect to it.
		_, err := gClient.Ping(ctx, &pb.PingRequest{Message: "Ping from scheduler"})
		return err
	})
	return pingOperation(gCtx)
}

func newDstsProtocolHeader() *pb.DstsRequestHeader {
	return &pb.DstsRequestHeader{
		ProtocolVersion: dstsProtocolVersion,
		RequestId:       uuid.NewString(),
		RequestTime:     timestamppb.Now(),
	}
}

type retryableOperation func(ctx context.Context) error

func retryWithBackoff(operation retryableOperation) retryableOperation {
	return func(context.Context) error {
		var err error
		for i := 0; i < dstsOperationRetryCount; i++ {
			err = gCtx.Err()
			if err != nil {
				schedLogger.Error("Global context has been closed. Aborting DSTS call!",
					zap.Error(err),
				)
				return err
			}

			// Invoke the requested RPC on the DSTS.
			ctx, cancelFunc := context.WithTimeout(gCtx, dstsRequestTimeout)
			err = operation(ctx)
			cancelFunc()
			if err == nil {
				break
			}

			secRetry := time.Duration(math.Pow(2, float64(i))) * baseRetryDuration
			schedLogger.Error("DSTS RPC call failed!",
				zap.Int("Attempt", i+1),
				zap.Duration("Retrying in", secRetry),
				zap.Error(err))
			time.Sleep(secRetry)
		}

		if err != nil {
			schedLogger.Error("All attempts to make the RPC call failed!")
		}
		return err
	}
}
