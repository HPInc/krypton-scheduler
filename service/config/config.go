package config

import (
	"crypto/rsa"
	"time"
)

const (
	ServiceName = "HP Scheduler Service"

	// Path to the registered services configuration YAML file.
	defaultRegisteredServicesConfigFilePath = "registered_services.yaml"
)

type ServerConfig struct {
	// Hostname of the DSTS service.
	Host string `yaml:"host"`

	// Port on which the REST server is available.
	RestPort int `yaml:"rest_port"`

	// Location of the registered applications configuration file.
	RegisteredServiceConfigFile string `yaml:"registered_service_config"`

	// Specifies whether to log all incoming REST requests to the debug log.
	DebugLogRestRequests bool `yaml:"log_rest_requests"`

	// Specifies whether app access tokens are required to invoke the
	// scheduler REST API.
	AuthenticateRestApiRequests bool `yaml:"api_authn_enabled"`
}

type DatabaseConfig struct {
	// The Cassandra keyspace within which to store scheduled task information.
	Keyspace string `yaml:"keyspace_name"`

	// Valid values are:
	// cassandra - use a local cassandra database instance.
	// aws_keyspaces - use an AWS Keyspaces managed Cassandra instance.
	DatabaseType string `yaml:"type"`

	// Hosts on which the database cluster is located
	DatabaseHosts []string `yaml:"db_hosts"`

	// Ports exposed by Cassandra:
	// - TCP 7000	Cassandra inter-node cluster communication.
	// - TCP 7001	Cassandra SSL inter-node cluster communication.
	// - TCP 7199	Cassandra JMX monitoring port.
	// - TCP 9042	Cassandra client port.
	// - TCP 9160	Cassandra Thrift client port.
	// - TCP 9404	Prometheus plugin port.
	ClientPort int `yaml:"client_port"`

	// The path to the schema migration scripts for the identity database.
	SchemaMigrationScripts string `yaml:"schema"`

	// Whether to perform schema migration.
	SchemaMigrationEnabled bool `yaml:"migrate"`

	// Specifies whether database calls should be debug logged.
	DebugLoggingEnabled bool `yaml:"debug"`

	// The username to use when connecting to the datastore.
	Username string `yaml:"user"`

	// Database password. For security reasons, this may not be specified
	// using the configuration YAML file.
	Password string
}

// Queue manager configuration settings.
type QueueMgrConfig struct {
	Endpoint string `yaml:"endpoint"`

	// The delay with which to watch for new messages in the queue.
	WatchDelay int32 `yaml:"watch_delay"`

	// Name of the input queue on which the scheduler listens for requests
	// to schedule new tasks.
	InputQueueName string `yaml:"input_queue"`

	// Name of the dispatch queue on which the scheduler listens for tasks to
	// be dispatched to the MQTT broker for delivery.
	DispatchQueueName string `yaml:"dispatch_queue"`

	// Name of the input queue on which the DCM service listens for device
	// configuration events.
	DcmInputQueueName string `yaml:"dcm_queue"`
}

// MQTT configuration settings.
type MqttConfig struct {
	// Hosts on which the MQTT broker is located
	MqttBrokerHosts []string `yaml:"mqtt_hosts"`

	// Valid values are:
	// aws_iot - use an AWS IoT core endpoint as MQTT broker.
	// local   - use a local MQTT broker instance such as HiveMQ.
	BrokerType string `yaml:"broker_type"`

	// The path to the TLS Root CA certificate used to connect to the broker.
	TlsRootCertPath string `yaml:"tls_cert_path"`

	// Seconds between keep alive packets
	KeepAlive uint16 `yaml:"keep_alive"`

	// Period between connection attempts
	ConnectRetryDelay time.Duration `yaml:"connect_retry_delay"`

	// The quality of service (QoS) to send messages with.
	QoS int `yaml:"qos"`

	// MQTT client ID of this scheduler node. Generated from the scheduler
	// App ID (SchedulerAppId) registered with DSTS.
	ClientId string

	// Specifies whether MQTT messages should be debug logged.
	DebugLoggingEnabled bool `yaml:"debug"`
}

// DSTS configuration settings.
type DstsConfig struct {
	// Hostname of the DSTS service.
	Host string `yaml:"host"`

	// Port on which the DSTS RPC server is available.
	RpcPort int `yaml:"rpc_port"`

	// The App ID for the scheduler that is registered with the DSTS as a
	// registered app. This App ID will be used to request app tokens from the
	// DSTS.
	SchedulerAppId string `yaml:"scheduler_app_id"`

	// The name of the environment variable that contains the private key for
	// the scheduler app. This private key is used to sign assertions and request
	// app tokens from the DSTS.
	PrivateKeyEnv string `yaml:"private_key_env"`

	// The private key for the scheduler app.
	PrivateKey *rsa.PrivateKey
}

// AWS configuration settings.
type AwsSettings struct {
	// The path to the AWS_WEB_IDENTITY_TOKEN_FILE which contains the role's
	// credentials (IRSA) used to connect to AWS Keyspaces.
	AwsWebIdentityTokenFile string

	// The AWS role under which the scheduler service runs. The credentials of
	// this role are used to access the AWS Keyspace instance. This value is not
	// used if the database typs is cassandra.
	AwsRoleArn string

	// The AWS region in which the scheduler service is deployed.
	Region string
}

type Config struct {
	ConfigFilePath string

	// Configuration settings for the gRPC server.
	ServerConfig `yaml:"server"`

	// Database configuration settings.
	DatabaseConfig `yaml:"database"`

	// Queue manager configuration settings.
	QueueMgrConfig `yaml:"queuemgr"`

	// MQTT configuration settings.
	MqttConfig `yaml:"mqtt"`

	// DSTS configuration settings.
	DstsConfig `yaml:"dsts"`

	// AWS configuration settings. These are derived from environment variables
	// and not exposed in the configuration file to avoid exposure risk for
	// credentials and secrets.
	AwsSettings

	// Whether the service is running in test mode.
	TestMode bool `yaml:"test_mode"`
}
