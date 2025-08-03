package config

import (
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type RegisteredServiceConfig struct {
	Services []ServiceRegistration `yaml:"registered_services"`
}

// Registration information about services that interact with the scheduler.
type ServiceRegistration struct {
	// Name of the service being registered.
	Name string `yaml:"name"`

	// Unique identifier assigned to the service.
	ServiceId string `yaml:"service_id"`

	// The AWS account that owns the SQS topic referenced by this entry.
	OwnerAwsAccount string `yaml:"owner_aws_account"`

	// Map of MQTT topics that the service is interested in and corresponding
	// SQS input topics on which it would like to receive these MQTT messages.
	Topics map[string]string `yaml:"topics"`
}

// Load configuration information for registered services from the YAML configuration
// file.
func (c *ConfigMgr) LoadRegisteredServiceConfig(configFilePath string) bool {
	var filename string = defaultRegisteredServicesConfigFilePath

	// Check if the default configuration file has been overridden using the
	// environment variable.
	if configFilePath != "" {
		schedLogger.Info("Using configuration file specified by environment variable.",
			zap.String("Service configuration file:", configFilePath),
		)
		filename = configFilePath
	}

	// Open the configuration file for parsing.
	fh, err := os.Open(filename)
	if err != nil {
		schedLogger.Error("Failed to load registered service configuration file!",
			zap.String("Service configuration file:", filename),
			zap.Error(err),
		)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	decoder := yaml.NewDecoder(fh)
	err = decoder.Decode(&c.serviceConfig)
	if err != nil {
		schedLogger.Error("Failed to parse registered service configuration file!",
			zap.String("Service configuration file:", filename),
			zap.Error(err),
		)
		_ = fh.Close()
		return false
	}

	_ = fh.Close()
	schedLogger.Info("Parsed registered service configuration from the configuration file!",
		zap.String("Service configuration file:", filename),
	)

	c.Display()
	return true
}
