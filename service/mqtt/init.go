package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/dstsclient"
	"go.uber.org/zap"
)

var (
	// Structured logging support.
	schedLogger      *zap.Logger
	sugarSchedLogger *zap.SugaredLogger

	// Global context and cancellation support.
	mqttCtx    context.Context
	cancelFunc context.CancelFunc

	// MQTT client configuration settings.
	clientConfig autopaho.ClientConfig
	mqttConfig   *config.MqttConfig
	connMgr      *autopaho.ConnectionManager
	connMutex    sync.RWMutex

	// Handler functions
	messageHandlers *MessageHandlers
)

const (
	mqttConnectionTimeout = 5 * (time.Second)

	// Types of MQTT brokers supported by the service.
	brokerTypeAwsIoT = "aws_iot"
	brokerTypeLocal  = "local"

	userNameFormat = "username?x-amz-customauthorizer-name=KryptonIoTAuthorizer&device_token=%s"
)

// Handlers (callback functions) used to process messages received from the
// MQTT broker.
type MessageHandlers struct {
	// Process messages sent from devices to their device management service.
	// Process task response events for tasks dispatched by the MQTT broker
	// on behalf of the device management service.
	OnDeviceMessage func(mqttTopic string, message *pb.DeviceMessage) error

	OnTaskResponse func(mqttTopic string, message *pb.DeviceMessage) error
}

func updateMqttUsername() error {
	// If connecting to AWS IoT core, specify the app access token using the
	// username field and request the Krypton custom IoT authorizer.
	if mqttConfig.BrokerType == brokerTypeAwsIoT {
		clientConfig.ResetUsernamePassword()
		token, err := dstsclient.GetAccessToken()
		if err != nil {
			return err
		}
		clientConfig.SetUsernamePassword(fmt.Sprintf(userNameFormat, token), nil)
	}
	return nil
}

func reconnect() {
	if mqttCtx.Err() != nil {
		schedLogger.Info("The MQTT publish context has been canceled",
			zap.Error(mqttCtx.Err()),
		)
		return
	}

	connMutex.Lock()
	defer connMutex.Unlock()

	ctx, cancel := context.WithTimeout(mqttCtx, mqttConnectionTimeout)
	defer cancel()

	// Disconnect the existing connection.
	err := connMgr.Disconnect(ctx)
	if err != nil {
		schedLogger.Error("Failed to disconnect connection manager!",
			zap.Error(err),
		)
	}

	// Connect to the broker - this will return immediately after initiating
	// the connection process
	connMgr, err = autopaho.NewConnection(mqttCtx, clientConfig)
	if err != nil {
		schedLogger.Error("Failed to establish a connection to the broker!",
			zap.String("MQTT client ID", mqttConfig.ClientId),
			zap.Error(err),
		)
	}
}

func Init(logger *zap.Logger, cfgMgr *config.ConfigMgr,
	handlers *MessageHandlers) error {
	var (
		err error
	)

	schedLogger = logger
	sugarSchedLogger = schedLogger.Sugar()
	mqttConfig = cfgMgr.GetMqttConfig()
	messageHandlers = handlers

	mqttCtx, cancelFunc = context.WithCancel(context.Background())

	// Connect to the device security token service (DSTS)
	err = dstsclient.Start(mqttCtx, logger, cfgMgr)
	if err != nil {
		schedLogger.Error("Failed to connect to the DSTS!",
			zap.Error(err),
		)
		cancelFunc()
		return err
	}

	mqttConfig.ClientId = generateClientId(cfgMgr.GetSchedulerAppID())

	// Parse the MQTT broker URLs specified in configuration.
	brokerUrl := make([]*url.URL, len(mqttConfig.MqttBrokerHosts))
	for i := 0; i < len(mqttConfig.MqttBrokerHosts); i++ {
		parsedUrl, err := url.Parse(mqttConfig.MqttBrokerHosts[i])
		if err != nil {
			schedLogger.Error("Failed to parse the provided broker URL",
				zap.String("URL provided", mqttConfig.MqttBrokerHosts[0]),
				zap.Error(err),
			)
			return err
		}
		brokerUrl[i] = parsedUrl
	}

	// Initialize the MQTT client configuration based on settings parsed from
	// the configuration file and environment overrides.
	clientConfig = autopaho.ClientConfig{
		BrokerUrls:        brokerUrl,
		TlsCfg:            nil,
		KeepAlive:         mqttConfig.KeepAlive,
		ConnectRetryDelay: mqttConfig.ConnectRetryDelay * time.Second,
		OnConnectionUp:    connectionUpCallback,
		OnConnectError:    connectionErrorCallback,
		PahoErrors:        mqttLogger{},
		ClientConfig: paho.ClientConfig{
			ClientID:           mqttConfig.ClientId,
			OnClientError:      clientErrorCallback,
			OnServerDisconnect: serverDisconnectCallback,
			Router:             initMqttSubscriberRouter(),
		},
	}

	if mqttConfig.DebugLoggingEnabled {
		clientConfig.Debug = mqttLogger{}
		clientConfig.PahoDebug = mqttLogger{}
	}

	// If a TLS root cert path was specified, initialize the TLS configuration.
	if mqttConfig.TlsRootCertPath != "" {
		certs, err := loadTlsCert(mqttConfig.TlsRootCertPath)
		if err != nil {
			return err
		}
		clientConfig.TlsCfg = &tls.Config{
			RootCAs:    certs,
			NextProtos: []string{"mqtt"},
			MinVersion: tls.VersionTLS12,
		}
	}

	// If connecting to AWS IoT core, specify the app access token using the
	// username field and request the Krypton custom IoT authorizer.
	err = updateMqttUsername()
	if err != nil {
		schedLogger.Error("Failed to acquire an app access token from DSTS!",
			zap.Error(err),
		)
		cancelFunc()
		return err
	}

	// Connect to the broker - this will return immediately after initiating
	// the connection process
	connMutex.Lock()

	connMgr, err = autopaho.NewConnection(mqttCtx, clientConfig)
	connMutex.Unlock()
	if err != nil {
		schedLogger.Error("Failed to establish a connection to the broker!",
			zap.String("MQTT client ID", mqttConfig.ClientId),
			zap.Error(err),
		)
		cancelFunc()
		return err
	}

	// Parse the configuration to determine how to route messages received on
	// various MQTT topics to the appropriate service.
	for _, service := range *cfgMgr.GetServiceRegistrations() {
		_, err := db.GetRegisteredService(service.ServiceId)
		if err == db.ErrNotFound {
			// #nosec G601
			err = db.NewRegisteredService(&service).CreateRegisteredService()
			if err != nil {
				schedLogger.Error("Failed to create a registration for the requested service!",
					zap.String("Service ID", service.ServiceId),
					zap.Error(err),
				)
				cancelFunc()
				return err
			}
		}
	}

	return nil
}

func loadTlsCert(rootCertPath string) (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(filepath.Clean(rootCertPath))
	if err != nil {
		schedLogger.Error("Failed to read the root CA certificate file!",
			zap.String("CA certificate path", rootCertPath),
			zap.Error(err),
		)
		return nil, err
	}
	if !certs.AppendCertsFromPEM(pemData) {
		schedLogger.Error("Failed to read the root CA certificate file!",
			zap.String("CA certificate path", rootCertPath),
			zap.Error(err),
		)
		return nil, errors.New("failed to append root ca cert")
	}

	return certs, nil
}

func Shutdown() error {
	cancelFunc()
	return sugarSchedLogger.Sync()
}
