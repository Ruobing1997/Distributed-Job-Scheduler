package constants

import (
	"errors"
	"time"
)

type TaskDB struct {
	ID                string
	Name              string
	Type              int
	Schedule          string
	Payload           string
	CallbackURL       string
	Status            int
	ExecutionTime     time.Time
	NextExecutionTime time.Time
	CreateTime        time.Time
	UpdateTime        time.Time
	Retries           int
	Result            int
}

func (t *TaskDB) Validate() error {
	if t.Name == "" {
		return errors.New("name cannot be empty")
	} else if t.Type < 0 || t.Type > 1 {
		return errors.New("type must be 0 or 1")
	} else if t.Payload == "" {
		return errors.New("payload cannot be empty")
	} else if t.Retries < 0 {
		return errors.New("retries must be greater than or equal to 0")
	}
	return nil
}

type TaskCache struct {
	ID                string
	Index             int
	ExecutionTime     time.Time
	NextExecutionTime time.Time
}

var statusMap = map[int]string{
	0: "Pending",
	1: "Running",
	2: "Completed",
	3: "Failed",
}

var resultMap = map[int]string{
	0: "In Queue",
	1: "Success",
	2: "Failed",
}

var typeMap = map[int]string{
	0: "Alarm",
	1: "Concurrent",
}

const MONGOURI = "mongodb://localhost:27017"

const OFFSET = 10000

const SHARDSAMOUNT = 6
const EIGHTBIT = 8
const SIXTEENBIT = 16
const TWENTYFOURBIT = 24
const REDISPORTOFFSET = 6900
