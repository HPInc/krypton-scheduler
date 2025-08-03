package db

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Add the specified consignment to the scheduler database.
func (c *Consignment) CreateConsignment() error {
	// Ensure the tenant ID, consignment ID etc. were specified by the caller.
	_, err := uuid.Parse(c.TenantID)
	if err != nil {
		schedLogger.Error("Invalid tenant ID",
			zap.Error(err),
		)
		return ErrInvalidRequest
	}

	if c.ConsignmentID == "" {
		schedLogger.Error("Invalid consignment ID")
		return ErrInvalidRequest
	}

	// Create a new consignment entry in the scheduler database.
	gSessionMutex.RLock()
	err = gSession.Query(consignmentsStatements.insert.statement,
		consignmentsStatements.insert.names).
		BindStruct(c).
		ExecRelease()
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to add consignment to task mapping to the scheduler database!",
			zap.String("Tenant ID:", c.TenantID),
			zap.String("Consignment ID:", c.ConsignmentID),
			zap.String("Task ID:", c.TaskID.String()),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func UpdateConsignmentTaskStatus(tenantID string, consignmentID string, taskID string,
	status TaskStatus) error {
	gSessionMutex.RLock()
	err := gSession.Session.Query(`UPDATE consignments SET status=? WHERE tenant_id=? AND consignment_id=? AND task_id=?`,
		status.String(), tenantID, consignmentID, taskID).Exec()
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to update the task status!",
			zap.String("Task ID:", taskID),
			zap.String("Tenant ID:", tenantID),
			zap.String("Consignment ID:", consignmentID),
			zap.Error(err),
		)
		return err
	}

	return nil
}
