package config

import (
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type envSetting struct {
	isSecret bool
	value    interface{}
}

// loadEnvironmentVariableOverrides - check values specified for supported
// environment variables. These can be used to override configuration settings
// specified in the config file.
func (c *ConfigMgr) loadEnvironmentVariableOverrides() {
	m := map[string]envSetting{
		// Scheduler service configuration settings
		"SCHEDULER_SERVER":                         {value: &c.config.ServerConfig.Host},
		"SCHEDULER_PORT":                           {value: &c.config.ServerConfig.RestPort},
		"SCHEDULER_REGISTERED_SERVICE_CONFIG_FILE": {value: &c.config.ServerConfig.RegisteredServiceConfigFile},
		"SCHEDULER_REST_DEBUG_LOGGING":             {value: &c.config.ServerConfig.DebugLogRestRequests},
		"SCHEDULER_REST_API_AUTH_ENABLED":          {value: &c.config.ServerConfig.AuthenticateRestApiRequests},

		// Database configuration settings
		"SCHEDULER_DB_TYPE":            {value: &c.config.DatabaseConfig.DatabaseType},
		"SCHEDULER_DB_HOSTS":           {value: &c.config.DatabaseConfig.DatabaseHosts},
		"SCHEDULER_DB_PORT":            {value: &c.config.DatabaseConfig.ClientPort},
		"SCHEDULER_DB_USER":            {value: &c.config.DatabaseConfig.Username},
		"SCHEDULER_DB_PASSWORD":        {isSecret: true, value: &c.config.DatabaseConfig.Password},
		"SCHEDULER_DB_SCHEMA_LOCATION": {value: &c.config.DatabaseConfig.SchemaMigrationScripts},

		// Notification configuration settings
		"SCHEDULER_QUEUE_ENDPOINT":       {value: &c.config.QueueMgrConfig.Endpoint},
		"SCHEDULER_INPUT_QUEUE_NAME":     {value: &c.config.QueueMgrConfig.InputQueueName},
		"SCHEDULER_DISPATCH_QUEUE_NAME":  {value: &c.config.QueueMgrConfig.DispatchQueueName},
		"SCHEDULER_DCM_INPUT_QUEUE_NAME": {value: &c.config.QueueMgrConfig.DcmInputQueueName},
		"SCHEDULER_QUEUE_WATCH_DELAY":    {value: &c.config.QueueMgrConfig.WatchDelay},

		// MQTT configuration settings
		"SCHEDULER_MQTT_BROKER_HOSTS":        {value: &c.config.MqttConfig.MqttBrokerHosts},
		"SCHEDULER_MQTT_BROKER_TYPE":         {value: &c.config.MqttConfig.BrokerType},
		"SCHEDULER_MQTT_TLS_CERT_PATH":       {value: &c.config.MqttConfig.TlsRootCertPath},
		"SCHEDULER_MQTT_KEEP_ALIVE":          {value: &c.config.MqttConfig.KeepAlive},
		"SCHEDULER_MQTT_CONNECT_RETRY_DELAY": {value: &c.config.MqttConfig.ConnectRetryDelay},
		"SCHEDULER_MQTT_DEBUG_LOGGING":       {value: &c.config.MqttConfig.DebugLoggingEnabled},

		// DSTS configuration settings
		"SCHEDULER_DSTS_HOST":     {value: &c.config.DstsConfig.Host},
		"SCHEDULER_DSTS_RPC_PORT": {value: &c.config.DstsConfig.RpcPort},

		// AWS configuration settings
		"AWS_REGION":                  {value: &c.config.AwsSettings.Region},
		"AWS_WEB_IDENTITY_TOKEN_FILE": {value: &c.config.AwsSettings.AwsWebIdentityTokenFile},
		"AWS_ROLE_ARN":                {value: &c.config.AwsSettings.AwsRoleArn},
	}
	for k, v := range m {
		e := os.Getenv(k)
		if e != "" {
			schedLogger.Info("Overriding configuration from environment variable.",
				zap.String("variable: ", k),
				zap.String("value: ", getLoggableValue(v.isSecret, e)))
			v := v
			replaceConfigValue(os.Getenv(k), &v)
		}
	}
}

// envValue will be non empty as this function is private to file
func replaceConfigValue(envValue string, t *envSetting) {
	switch t.value.(type) {
	case *string:
		*t.value.(*string) = envValue
	case *[]string:
		valSlice := strings.Split(envValue, ",")
		for i := range valSlice {
			valSlice[i] = strings.TrimSpace(valSlice[i])
		}
		*t.value.(*[]string) = valSlice
	case *bool:
		b, err := strconv.ParseBool(envValue)
		if err != nil {
			schedLogger.Error("Bad bool value in env")
		} else {
			*t.value.(*bool) = b
		}
	case *int:
		i, err := strconv.Atoi(envValue)
		if err != nil {
			schedLogger.Error("Bad integer value in env",
				zap.Error(err))
		} else {
			*t.value.(*int) = i
		}
	default:
		schedLogger.Error("There was a bad type map in env override",
			zap.String("value", envValue))
	}
}

func getLoggableValue(isSecret bool, value string) string {
	if isSecret {
		return "***"
	}
	return value
}
