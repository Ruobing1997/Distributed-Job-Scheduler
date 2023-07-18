package databasehandler

import "context"

type DataRecord interface {
	Validate() error
}

type DatabaseClient interface {
	InsertTask(ctx context.Context, record DataRecord) error
	GetTaskByID(ctx context.Context, id string, args ...interface{}) (DataRecord, error)
	Close() error
}
