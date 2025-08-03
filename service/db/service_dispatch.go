package db

import (
	"fmt"

	"github.com/hpinc/krypton-scheduler/service/config"
	"go.uber.org/zap"
)

const (
	keyFormat = "%s::%s"
)

// The dispatch table stores a mapping between the MQTT topic and its
// corresponding queue topic for each registered service. When the scheduler
// receives an MQTT message, it uses the service ID and the MQTT topic name
// to determine the service queue topic to which the message is dispatched.
var serviceDispatchTable map[string]string
var serviceConfigTable map[string]*RegisteredService

// Initalize a dispatch lookup table that stores a mapping between MQTT topic
// and corresponding queue topic on a per service basis. This lookup table is
// used to determine where to dispatch messages received by the scheduler from
// the MQTT broker for consumption by the target service.
func initServiceDispatchLookupTable(
	serviceRegistrations *[]config.ServiceRegistration) error {
	var routeCount = 0

	// Parse the configuration to determine how to route messages received on
	// various MQTT topics to the appropriate service.
	for _, service := range *serviceRegistrations {
		_, err := GetRegisteredService(service.ServiceId)
		if err == ErrNotFound {
			// #nosec G601
			err = NewRegisteredService(&service).CreateRegisteredService()
			if err != nil {
				schedLogger.Error("Failed to create a registration for the requested service!",
					zap.String("Service ID:", service.ServiceId),
					zap.Error(err),
				)
				return err
			}
		}
	}

	services, err := ListRegisteredServices()
	if err != nil {
		schedLogger.Error("Failed to list registered services!",
			zap.Error(err),
		)
		return err
	}

	serviceConfigTable = make(map[string]*RegisteredService, len(services))
	for _, service := range services {
		routeCount += len(service.Topics)
		serviceConfigTable[service.ServiceID] = service
	}

	serviceDispatchTable = make(map[string]string, routeCount)
	for _, service := range services {
		for key, value := range service.Topics {
			serviceDispatchTable[fmt.Sprintf(keyFormat,
				service.ServiceID, key)] = value
		}
	}

	return nil
}

func GetServiceQueueTopic(serviceID string, mqttTopic string) string {
	queueTopic, ok := serviceDispatchTable[fmt.Sprintf(keyFormat,
		serviceID, mqttTopic)]
	if !ok {
		return ""
	}

	return queueTopic
}

func IsValidServiceId(serviceID string) bool {
	_, ok := serviceConfigTable[serviceID]
	return ok
}

func GetServiceConfig(serviceID string) *RegisteredService {
	cfg, ok := serviceConfigTable[serviceID]
	if !ok {
		return nil
	}
	return cfg
}
