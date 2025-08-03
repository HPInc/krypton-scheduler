package mqtt

import (
	"context"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/hpinc/krypton-scheduler/service/dstsclient"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"go.uber.org/zap"
)

// OnConnectionUp callback is invoked when a connection is made, including
// on reconnections.
func connectionUpCallback(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	if mqttCtx.Err() != nil {
		return
	}

	schedLogger.Info("MQTT connection is up! Attempting to subscribe for topics",
		zap.Bool("Session present", connAck.SessionPresent),
		zap.Int("Reason code", int(connAck.ReasonCode)),
	)

	// MQTT connection is up - subscribe to the required topics.
	connMutex.RLock()
	ctx, cancel := context.WithTimeout(mqttCtx, mqttConnectionTimeout)
	defer cancel()
	ack, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: mqttSubscriptionsMap,
	})
	connMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to subscribe to MQTT topics!",
			zap.String("MQTT client ID", mqttConfig.ClientId),
			zap.String("Reason string", ack.Properties.ReasonString),
			zap.ByteString("Reasons", ack.Reasons),
			zap.Error(err),
		)
		return
	}

	schedLogger.Info("Subscribed to MQTT topics!",
		zap.String("MQTT client ID", mqttConfig.ClientId),
		zap.String("Reason string", ack.Properties.ReasonString),
	)
}

// OnConnectError callback is invoked when a connection attempt fails.
func connectionErrorCallback(err error) {
	// If the app access token has expired, refresh it.
	if dstsclient.IsAppTokenExpired() {
		err = updateMqttUsername()
		if err != nil {
			// All retry attempts to acquire an app token from DSTS have failed.
			schedLogger.Error("All attempts to acquire DSTS app token have failed!",
				zap.Error(err),
			)
			_ = Shutdown()
			panic(err)
		}
	}

	metrics.MetricMqttConnectionErrorCount.Inc()
	schedLogger.Error("MQTT client encountered a connection error!",
		zap.String("MQTT client ID", mqttConfig.ClientId),
		zap.Error(err),
	)
	reconnect()
}

// OnClientError callback is for example called on net.Error
func clientErrorCallback(err error) {
	metrics.MetricMqttErrorCount.Inc()
	schedLogger.Error("MQTT client encountered an error!",
		zap.String("MQTT client ID", mqttConfig.ClientId),
		zap.Error(err),
	)
}

// OnServerDisconnect callback is called only when a packets.DISCONNECT
// is received from server
func serverDisconnectCallback(disconnect *paho.Disconnect) {
	schedLogger.Info("MQTT broker requested a disconnect!",
		zap.String("MQTT client ID", mqttConfig.ClientId),
		zap.Int("Reason code", int(disconnect.ReasonCode)),
	)

	// 135 - client is unauthorized. This happens when the app access token of
	// the scheduler has expired. Renew the access token and create a fresh
	// connection.
	if disconnect.ReasonCode == 135 {
		err := updateMqttUsername()
		if err != nil {
			// All retry attempts to acquire an app token from DSTS have failed.
			schedLogger.Error("All attempts to acquire DSTS app token have failed!",
				zap.Error(err),
			)
			_ = Shutdown()
			panic(err)
		}
		reconnect()
		return
	}

	if disconnect.Properties != nil {
		schedLogger.Error("MQTT broker requested a disconnect!",
			zap.String("Reason:", disconnect.Properties.ReasonString),
		)
	}
}
