package db

import (
	"go.uber.org/zap"
)

// Add the specified registered service to the scheduler database.
func (s *RegisteredService) CreateRegisteredService() error {
	if s.ServiceID == "" {
		schedLogger.Error("Invalid service ID")
		return ErrInvalidRequest
	}

	// Create a new service registration entry in the scheduler database.
	gSessionMutex.RLock()
	err := gSession.Query(registeredServicesStatements.insert.statement,
		registeredServicesStatements.insert.names).
		BindStruct(s).
		ExecRelease()
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to create a registered service entry in the scheduler database!",
			zap.String("Service ID:", s.ServiceID),
			zap.Error(err),
		)
		return err
	}

	return nil
}
