package db

import (
	"github.com/scylladb/gocqlx/v2/qb"
	"go.uber.org/zap"
)

func GetRegisteredService(serviceID string) (*RegisteredService, error) {
	if serviceID == "" {
		schedLogger.Error("Invalid service ID was specified!")
		return nil, ErrInvalidRequest
	}

	var foundService []*RegisteredService
	gSessionMutex.RLock()
	err := qb.Select(registeredServicesMetadata.Name).
		Where(qb.Eq("service_id")).
		Query(gSession).
		Bind(serviceID).
		SelectRelease(&foundService)
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to execute query to find service ID!",
			zap.String("Service ID:", serviceID),
			zap.Error(err),
		)
		return nil, err
	}

	if len(foundService) > 0 {
		return foundService[0], nil
	}
	return nil, ErrNotFound
}

func ListRegisteredServices() ([]*RegisteredService, error) {
	var foundServices []*RegisteredService

	gSessionMutex.RLock()
	err := qb.Select(registeredServicesMetadata.Name).
		Query(gSession).
		SelectRelease(&foundServices)
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to list registered services!",
			zap.Error(err),
		)
		return nil, err
	}

	return foundServices, nil
}
