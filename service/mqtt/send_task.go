package mqtt

import (
	"context"

	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

func SendTaskToBroker(topicName string, payload *[]byte) error {
	if mqttCtx.Err() != nil {
		schedLogger.Info("Stopping the MQTT publisher goroutine! Context has been canceled")
		return mqttCtx.Err()
	}

	// AwaitConnection will return immediately if connection is up; adding
	// this call stops publication whilst the connection is unavailable.
	connMutex.RLock()
	ctx, cancel := context.WithTimeout(mqttCtx, mqttConnectionTimeout)
	defer cancel()

	err := connMgr.AwaitConnection(ctx)
	if err != nil {
		connMutex.RUnlock()
		// Should only happen when context is cancelled
		schedLogger.Error("Publisher done - AwaitConnection returned an error",
			zap.Error(err),
		)
		return err
	}

	result, err := connMgr.Publish(ctx, &paho.Publish{
		QoS:     byte(mqttConfig.QoS),
		Topic:   topicName,
		Payload: *payload,
	})
	connMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to publish message to the requested MQTT topic!",
			zap.String("Topic name", topicName),
			zap.Error(err),
		)
		return err
	} else {
		if result != nil {
			if result.Properties != nil {
				schedLogger.Info("Publish response.",
					zap.String("Topic name", topicName),
					zap.String("Reason string", result.Properties.ReasonString),
				)
			} else {
				if result.ReasonCode != 0 && result.ReasonCode != 16 {
					schedLogger.Info("Publish response.",
						zap.String("Topic name", topicName),
						zap.Int("Reason code", int(result.ReasonCode)),
					)
				}
			}
		}
	}

	return nil
}
