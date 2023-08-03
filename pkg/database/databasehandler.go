package databasehandler

import (
	"context"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"time"
)

type DataRecord interface {
	Validate() error
}

type DatabaseClient interface {
	InsertTask(ctx context.Context, table string, record DataRecord) error
	GetTaskByID(ctx context.Context, table string, id string, args ...interface{}) (DataRecord, error)
	UpdateByID(ctx context.Context, table string, id string, args map[string]interface{}) error
	DeleteByID(ctx context.Context, table string, id string) error
	GetTasksInInterval(startTime time.Time, endTime time.Time, timeTracker time.Time) ([]*constants.TaskDB, error)
	Close() error
	CountRunningTasks(ctx context.Context, idValue string) (int, error)
}
