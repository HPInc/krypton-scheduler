package db

import (
	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

var (
	// Metadata describing the registered services table in the scheduler database.
	registeredServicesMetadata table.Metadata

	registeredServicesTable *table.Table

	// Pre-created CQL query statements to interact with the registered services
	//  table.
	registeredServicesStatements *statements
)

type RegisteredService struct {
	// The service ID to which the entry belongs.
	ServiceID string `db:"service_id" json:"service_id"`

	// The name of the service
	Name string `db:"name" json:"name"`

	// The AWS account the SQS queues belong to.
	OwnerAwsAccount string `db:"owner_aws_account" json:"owner_aws_account"`

	// Mapping of MQTT topics and their corresponding SQS queues.
	Topics map[string]string `db:"topics" json:"topics"`
}

func NewRegisteredService(serviceInfo *config.ServiceRegistration) *RegisteredService {
	return &RegisteredService{
		ServiceID:       serviceInfo.ServiceId,
		Name:            serviceInfo.Name,
		OwnerAwsAccount: serviceInfo.OwnerAwsAccount,
		Topics:          serviceInfo.Topics,
	}
}

func createRegisteredServiceStatements() {
	registeredServicesMetadata = table.Metadata{
		Name: "registered_services",
		Columns: []string{
			"service_id",
			"name",
			"owner_aws_account",
			"topics",
		},
		PartKey: []string{
			"service_id",
		},
	}

	registeredServicesTable = table.New(registeredServicesMetadata)

	// Store pre-created CQL query statements to interact with the
	// registered services table.
	deleteStatement, deleteNames := registeredServicesTable.Delete()
	insertStatement, insertNames := registeredServicesTable.Insert()
	getStatement, getNames := qb.Select(registeredServicesMetadata.Name).
		Columns(registeredServicesMetadata.Columns...).ToCql()

	registeredServicesStatements = &statements{
		delete: query{
			statement: deleteStatement,
			names:     deleteNames,
		},
		insert: query{
			statement: insertStatement,
			names:     insertNames,
		},
		get: query{
			statement: getStatement,
			names:     getNames,
		},
	}
}
