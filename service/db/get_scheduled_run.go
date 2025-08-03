package db

import (
	"github.com/scylladb/gocqlx/v2/qb"
	"go.uber.org/zap"
)

const (
	// Default number of items to return per page of paginated results.
	itemsPerPage = 100
)

// Get scheduled runs within the specified run partition.
func GetScheduledRuns(runPartition string, startPage []byte) ([]*ScheduledRun,
	[]byte, error) {
	var foundSchedules []*ScheduledRun

	if runPartition == "" {
		return nil, nil, ErrInvalidRequest
	}

	gSessionMutex.RLock()
	query := qb.Select(scheduledRunMetadata.Name).
		Where(qb.Eq("run_partition")).
		Query(gSession).
		Bind(runPartition)
	defer func() {
		query.Release()
		gSessionMutex.RUnlock()
	}()

	query.PageState(startPage)
	query.PageSize(itemsPerPage)

	iter := query.Iter()
	err := iter.Select(&foundSchedules)
	if err != nil {
		schedLogger.Error("Failed to query for scheduled tasks to be run next",
			zap.Error(err),
		)
		return nil, nil, err
	}

	return foundSchedules, iter.PageState(), nil
}
