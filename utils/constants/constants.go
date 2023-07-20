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
	ID            string
	Payload       string
	ExecutionTime time.Time
	Retries       int
}

type PayloadJson struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
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

var payloadTypeMap = map[int]string{
	0: "shell",
	1: "python",
	2: "email",
}

const SHELL = 0
const PYTHON = 1
const EMAIL = 2

const MONGOURI = "mongodb://localhost:27017"

const GRPC_TIMEOUT = time.Second
const LEASE_DURATION = time.Second
const LEASE_RENEW_INTERVAL = time.Second // renew lease every 1 second
const OFFSET = 10000

const SHARDSAMOUNT = 6
const EIGHTBIT = 8
const SIXTEENBIT = 16
const TWENTYFOURBIT = 24
const REDISPORTOFFSET = 6900
