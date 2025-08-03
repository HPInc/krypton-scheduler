package db

import (
	"github.com/scylladb/gocqlx/v2/qb"
	"go.uber.org/zap"
)

// Get tasks queued within the scheduler database for the specified consignment
// ID.
func GetTasksForConsignment(tenantID string, consignmentID string,
	startPage []byte, pageSize int) ([]*Consignment, []byte, error) {
	if consignmentID == "" {
		schedLogger.Error("Invalid consignment ID specified!")
		return nil, nil, ErrInvalidRequest
	}

	var foundConsignments []*Consignment
	gSessionMutex.RLock()
	query := qb.Select(consignmentsMetadata.Name).
		Where(qb.Eq("tenant_id"), qb.Eq("consignment_id")).
		Query(gSession).
		BindMap(qb.M{"tenant_id": tenantID, "consignment_id": consignmentID})
	defer func() {
		query.Release()
		gSessionMutex.RUnlock()
	}()

	query.PageState(startPage)
	if pageSize == 0 {
		pageSize = itemsPerPage
	}
	query.PageSize(itemsPerPage)

	iter := query.Iter()
	err := iter.Select(&foundConsignments)
	if err != nil {
		schedLogger.Error("Failed to query for tasks in consignment!",
			zap.String("Consignment ID:", consignmentID),
			zap.Error(err),
		)
		return nil, nil, err
	}

	return foundConsignments, iter.PageState(), nil
}
