package db

import (
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/scylladb/gocqlx/v2"
	"go.uber.org/zap"
)

var (
	schedLogger   *zap.Logger
	clusterCfg    *gocql.ClusterConfig
	gSessionMutex sync.RWMutex
	gSession      gocqlx.Session
	dbConfig      *config.DatabaseConfig
)

const (
	cassandraMigrationConnStr  = "cassandra://%v:%v/%v?username=%v&password=%v"
	cassandraConnectionTimeout = (1 * time.Minute)

	// Database providers.
	providerCassandra    = "cassandra"
	providerAwsKeyspaces = "aws_keyspaces"
)

// Init - initialize the connection to the database.
func Init(logger *zap.Logger, cfgMgr *config.ConfigMgr) error {
	var err error

	schedLogger = logger
	dbConfig = cfgMgr.GetDatabaseConfig()

	// Initialize the connection to the database, depending on the keyspace provider
	// requested in the database configuration.
	switch dbConfig.DatabaseType {
	case providerCassandra:
		err = connectCassandraInstance()
		if err != nil {
			schedLogger.Error("Failed to initialize connection to Cassandra instance!",
				zap.Error(err),
			)
			return err
		}

	case providerAwsKeyspaces:
		err = connectAwsKeyspacesInstance(cfgMgr.GetAwsSettings())
		if err != nil {
			schedLogger.Error("Failed to initialize connection to AWS Keyspaces instance!",
				zap.Error(err),
			)
			return err
		}

	default:
		schedLogger.Error("Invalid keyspace provider specified in configuration!",
			zap.String("Provider specified:", dbConfig.DatabaseType),
		)
		return ErrInvalidDatabaseType
	}

	// Perform database schema migrations.
	err = migrateDatabaseSchema()
	if err != nil {
		schedLogger.Error("Failed to migrate the schema for the database!",
			zap.Error(err),
		)
		return err
	}

	// Initialize pre-created query statements for the database tables.
	createTaskStatements()
	createScheduledRunStatements()
	createConsignmentStatements()
	createRegisteredServiceStatements()

	// Initialize the service dispatch lookup table which maintains a mapping
	// between MQTT topic and corresponding service queue topic for each
	// registered service.
	err = initServiceDispatchLookupTable(cfgMgr.GetServiceRegistrations())
	if err != nil {
		schedLogger.Error("Failed to initialize the service dispatch lookup table!",
			zap.Error(err),
		)
		return err
	}

	schedLogger.Info("Connected to the scheduler database!",
		zap.Strings("Database host: ", dbConfig.DatabaseHosts),
		zap.Int("Client port: ", dbConfig.ClientPort),
	)
	return nil
}

// Initialize the connection to the cluster and retrieve a session object used
// to interact with the database cluster.
func connectCassandraInstance() error {
	var err error

	clusterCfg = gocql.NewCluster(dbConfig.DatabaseHosts...)
	clusterCfg.Port = dbConfig.ClientPort
	clusterCfg.Keyspace = dbConfig.Keyspace
	clusterCfg.Consistency = gocql.Quorum
	clusterCfg.ProtoVersion = 0
	clusterCfg.Timeout = cassandraConnectionTimeout
	clusterCfg.Authenticator = gocql.PasswordAuthenticator{
		Username: dbConfig.Username,
		Password: dbConfig.Password,
	}

	gSessionMutex.Lock()
	gSession, err = gocqlx.WrapSession(clusterCfg.CreateSession())
	gSessionMutex.Unlock()
	if err != nil {
		schedLogger.Error("Failed to create a session to the scheduler database!",
			zap.Error(err),
		)
		return err
	}

	return nil
}

func Shutdown() {
	gSessionMutex.Lock()
	gSession.Close()
	gSessionMutex.Unlock()

	schedLogger.Info("Shut down database session!")
	if dbConfig.DatabaseType == providerAwsKeyspaces {
		refreshSessionStopChannel <- true
	}
}
