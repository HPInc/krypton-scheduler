package config

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	// Path to the configuration YAML file.
	defaultConfigFilePath = "config.yaml"
)

var (
	schedLogger *zap.Logger
)

type ConfigMgr struct {
	config        Config
	serviceConfig RegisteredServiceConfig
	serviceName   string
}

// NewConfigMgr - initalize a new configuration manager instance.
func NewConfigMgr(logger *zap.Logger, serviceName string) *ConfigMgr {
	schedLogger = logger
	return &ConfigMgr{
		serviceName: serviceName,
	}
}

// Load configuration information from the YAML configuration file.
func (c *ConfigMgr) Load(testModeEnabled bool) bool {
	var filename string = defaultConfigFilePath

	// Check if the default configuration file has been overridden using the
	// environment variable.
	c.config.ConfigFilePath = os.Getenv("SCHEDULER_CONFIG_LOCATION")
	if c.config.ConfigFilePath != "" {
		schedLogger.Info("Using configuration file specified by command line switch.",
			zap.String("Configuration file:", c.config.ConfigFilePath),
		)
		filename = c.config.ConfigFilePath
	}

	// Open the configuration file for parsing.
	fh, err := os.Open(filename)
	if err != nil {
		schedLogger.Error("Failed to load configuration file!",
			zap.String("Configuration file:", filename),
			zap.Error(err),
		)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	decoder := yaml.NewDecoder(fh)
	err = decoder.Decode(&c.config)
	if err != nil {
		schedLogger.Error("Failed to parse configuration file!",
			zap.String("Configuration file:", filename),
			zap.Error(err),
		)
		_ = fh.Close()
		return false
	}

	_ = fh.Close()
	schedLogger.Info("Parsed configuration from the configuration file!",
		zap.String("Configuration file:", filename),
	)

	// Load any configuration overrides specified using environment variables.
	c.loadEnvironmentVariableOverrides()

	testModeEnvVar := os.Getenv("TEST_MODE")
	if (testModeEnvVar == "enabled") || (testModeEnabled) {
		c.config.TestMode = true
		fmt.Println("Scheduler service is running in test mode with test hooks enabled.")
	}

	return c.LoadRegisteredServiceConfig(c.config.RegisteredServiceConfigFile)
}

// Return the server configuration settings.
func (c *ConfigMgr) GetServerConfig() *ServerConfig {
	return &c.config.ServerConfig
}

// Return the database configuration settings.
func (c *ConfigMgr) GetDatabaseConfig() *DatabaseConfig {
	return &c.config.DatabaseConfig
}

// Return the queue manager configuration settings.
func (c *ConfigMgr) GetQueueMgrConfig() *QueueMgrConfig {
	return &c.config.QueueMgrConfig
}

// Return the MQTT configuration settings.
func (c *ConfigMgr) GetMqttConfig() *MqttConfig {
	return &c.config.MqttConfig
}

func (c *ConfigMgr) GetAwsSettings() *AwsSettings {
	return &c.config.AwsSettings
}

func (c *ConfigMgr) GetDstsConfig() *DstsConfig {
	return &c.config.DstsConfig
}

func (c *ConfigMgr) GetSchedulerAppID() string {
	return c.config.DstsConfig.SchedulerAppId
}

func (c *ConfigMgr) GetServiceRegistrations() *[]ServiceRegistration {
	return &c.serviceConfig.Services
}

// Update the database password retrieved from secrets manager or the
// environment variable.
func (c *ConfigMgr) SetDatabasePassword(dbPassword string) {
	c.config.DatabaseConfig.Password = dbPassword
}

// Check if the service is running in test mode.
func (c *ConfigMgr) IsTestModeEnabled() bool {
	return c.config.TestMode
}

func (c *ConfigMgr) IsDebugLoggingRestRequestsEnabled() bool {
	return c.config.ServerConfig.DebugLogRestRequests
}

// Display the configuration information parsed from the configuration file in
// the structured log.
func (c *ConfigMgr) Display() {
	schedLogger.Info("Current configuration",
		zap.String(" - Service name", c.serviceName),
		zap.Bool(" - Test mode enabled", c.config.TestMode),
	)
	schedLogger.Info("Server settings",
		zap.String(" - Hostname", c.config.ServerConfig.Host),
		zap.Int(" - Rest Port", c.config.ServerConfig.RestPort),
		zap.Bool(" - Debug logging enabled", c.config.DebugLogRestRequests),
		zap.Bool(" - Authentication enabled", c.config.AuthenticateRestApiRequests),
	)
	schedLogger.Info("Database settings",
		zap.String(" - Keyspace name", c.config.DatabaseConfig.Keyspace),
		zap.String(" - Keyspace provider", c.config.DatabaseConfig.DatabaseType),
		zap.Strings(" - Hosts", c.config.DatabaseConfig.DatabaseHosts),
	)
	schedLogger.Info("MQTT settings",
		zap.Strings(" - Broker hosts", c.config.MqttConfig.MqttBrokerHosts),
		zap.Uint16(" - Keep alive", c.config.MqttConfig.KeepAlive),
		zap.Duration(" - Connect retry delay", c.config.MqttConfig.ConnectRetryDelay),
		zap.Int("- Qos", c.config.MqttConfig.QoS),
		zap.Bool(" - Debug logging enabled", c.config.MqttConfig.DebugLoggingEnabled),
	)
	schedLogger.Info("DSTS settings",
		zap.String(" - Hostname", c.config.DstsConfig.Host),
		zap.Int(" - RPC Port", c.config.DstsConfig.RpcPort),
		zap.String(" - Scheduler App ID", c.config.DstsConfig.SchedulerAppId),
		zap.String(" - Private key env variable", c.config.DstsConfig.PrivateKeyEnv),
	)
	schedLogger.Info("Registered service settings",
		zap.Any("Services:", c.GetServiceRegistrations()),
	)
}
