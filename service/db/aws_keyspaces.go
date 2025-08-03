package db

import (
	"context"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin/sigv4"
	"github.com/gocql/gocql"
	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/scylladb/gocqlx/v2"
	"go.uber.org/zap"
)

const (
	awsOperationTimeout = time.Second * 5

	// Root cert for AWS Keyspaces.
	awsKeyspacesRootCertPath = "./sf-class2-root.crt"

	// Determines how frequently the AWS Keyspaces session should be refreshed.
	irsaRefreshInterval = 3 // hours
)

var (
	auth        sigv4.AwsAuthenticator
	awsSettings *config.AwsSettings

	// stop channel for AWS Keyspaces session refresher
	refreshSessionStopChannel chan bool
)

// Initialize the connection to the AWS Keyspaces instance and retrieve a
// session object used to interact with the database cluster.
func connectAwsKeyspacesInstance(settings *config.AwsSettings) error {
	var err error

	if settings.AwsWebIdentityTokenFile == "" {
		return ErrTokenFileNotFound
	}

	awsSettings = settings
	refreshSessionStopChannel = make(chan bool)

	clusterCfg = gocql.NewCluster(dbConfig.DatabaseHosts[0])
	clusterCfg.Port = dbConfig.ClientPort
	clusterCfg.Keyspace = dbConfig.Keyspace
	clusterCfg.Consistency = gocql.LocalQuorum
	clusterCfg.ProtoVersion = 0
	clusterCfg.Timeout = cassandraConnectionTimeout
	clusterCfg.SslOpts = &gocql.SslOptions{
		CaPath:                 awsKeyspacesRootCertPath,
		EnableHostVerification: false,
	}
	clusterCfg.DisableInitialHostLookup = false

	auth = sigv4.NewAwsAuthenticator()

	// Refresh the AWS Keyspaces session.
	err = refreshAwsKeyspacesSession()
	if err != nil {
		schedLogger.Error("Failed to retrieve role credentials to connect to AWS Keyspaces!",
			zap.Error(err),
		)
	}

	go awsKeyspacesSessionRefresher()
	return nil
}

// Retrieve fresh credentials of the AWS role under which the scheduler pods run.
// Establish a fresh session with the AWS Keyspaces instance.
func refreshAwsKeyspacesSession() error {

	// Initialize an AWS session and retrieve the AWS role creds for the
	// scheduler pod. These are used to connect to the AWS Keyspaces instance.
	ctx, cancelFunc := context.WithTimeout(context.Background(), awsOperationTimeout)
	defer cancelFunc()

	// Load the default AWS configuration - eg. AWS_REGION.
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		schedLogger.Error("Failed to load default AWS configuration!",
			zap.Error(err),
		)
		return err
	}

	// Configure a web identity role provider to fetch the IAM role credentials
	// for the scheduler service.
	credProvider := stscreds.NewWebIdentityRoleProvider(
		sts.NewFromConfig(cfg),
		awsSettings.AwsRoleArn,
		stscreds.IdentityTokenFile(awsSettings.AwsWebIdentityTokenFile))

	creds, err := credProvider.Retrieve(ctx)
	if err != nil {
		schedLogger.Error("Failed to retrieve IAM role credentials from identity token file!",
			zap.Error(err),
		)
		return err
	}

	// Update the creds for the SigV4 authenticator used by gocql for its session.
	gSessionMutex.Lock()
	defer gSessionMutex.Unlock()

	// Close the existing session and all connections to AWS Keyspaces.
	if gSession.Session != nil {
		gSession.Close()
	}

	// Create a fresh session with AWS Keyspaces.
	auth.AccessKeyId = creds.AccessKeyID
	auth.SecretAccessKey = creds.SecretAccessKey
	auth.SessionToken = creds.SessionToken
	clusterCfg.Authenticator = auth

	gSession, err = gocqlx.WrapSession(clusterCfg.CreateSession())
	if err != nil {
		schedLogger.Error("Failed to create a session to the AWS Keyspaces instance!",
			zap.String("Token file", awsSettings.AwsWebIdentityTokenFile),
			zap.Error(err),
		)
		return err
	}

	schedLogger.Info("Refreshed the session with AWS Keyspaces!",
		zap.String("Source", creds.Source),
		zap.String("Expires at", creds.Expires.Format(time.RFC1123)))
	return nil
}

func awsKeyspacesSessionRefresher() {
	ticker := time.NewTicker(time.Hour * time.Duration(irsaRefreshInterval))
	schedLogger.Info("Starting the AWS Keyspaces session refresher ...")

	for {
		select {
		case <-refreshSessionStopChannel:
			schedLogger.Info("Shutting down the AWS Keyspaces session refresher!")
			return

		case <-ticker.C:
			_ = refreshAwsKeyspacesSession()
		}
	}
}
