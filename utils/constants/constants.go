package constants

import (
	"errors"
	"time"
)

type TaskDB struct {
	ID                    string
	Name                  string
	JobType               int
	CronExpr              string
	Payload               PayloadJson
	CallbackURL           string
	Status                int
	ExecutionTime         time.Time
	NextExecutionTime     time.Time
	PreviousExecutionTime time.Time // 主要解决jiffy老师提到的问题：Leader故障切换期间到期护法的时间都被视为执行过
	CreateTime            time.Time
	UpdateTime            time.Time
	Retries               int // TODO: User应该不需要看到这个，可以删掉
	Result                int
}

func (t *TaskDB) Validate() error {
	if t.Name == "" {
		return errors.New("name cannot be empty")
	} else if t.JobType < 0 || t.JobType > 1 {
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
	ExecutionTime time.Time
	JobType       int
	Payload       PayloadJson
	RetriesLeft   int
	CronExpr      string
}

type PayloadJson struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

type RunTimeTask struct {
	ID            string
	ExecutionTime time.Time
	JobType       int
	JobStatus     int
	Payload       PayloadJson
	RetriesLeft   int
	CronExpr      string
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
	0: "OneTime",
	1: "Recurring",
}

var payloadTypeMap = map[int]string{
	0: "shell",
	1: "python",
	2: "email",
}

const JOBSUCCEED = 2
const JOBFAILED = 3

const SHELL = 0
const PYTHON = 1
const EMAIL = 2

const OneTime = 0
const Recurring = 1

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
const RUNNING_JOBS_RECORD = "execution_record"
const TASKS_FULL_RECORD = "task_db"
