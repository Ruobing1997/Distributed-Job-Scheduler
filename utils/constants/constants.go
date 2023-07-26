package constants

import (
	"errors"
	"time"
)

type TaskDB struct {
	ID                    string
	JobName               string
	JobType               int
	CronExpr              string
	Payload               *Payload
	CallbackURL           string
	Status                int
	ExecutionTime         time.Time
	PreviousExecutionTime time.Time // 主要解决jiffy老师提到的问题：Leader故障切换期间到期的时间都被视为执行过
	CreateTime            time.Time
	UpdateTime            time.Time
	Retries               int
}

func (t *TaskDB) Validate() error {
	if t.JobName == "" {
		return errors.New("name cannot be empty")
	} else if t.JobType < 0 || t.JobType > 1 {
		return errors.New("type must be 0 or 1")
	} else if t.Payload.Script == "" {
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
	Payload       *Payload
	RetriesLeft   int
	CronExpr      string
}

type Payload struct {
	Format int    `json:"format"`
	Script string `json:"script"`
}

type RunTimeTask struct {
	ID            string
	ExecutionTime time.Time
	JobType       int
	JobStatus     int
	Payload       *Payload
	RetriesLeft   int
	CronExpr      string
}

func (t *RunTimeTask) Validate() error {
	if t.JobType < 0 || t.JobType > 1 {
		return errors.New("type must be 0 or 1")
	} else if t.Payload.Script == "" {
		return errors.New("payload cannot be empty")
	} else if t.RetriesLeft < 0 {
		return errors.New("retries must be greater than or equal to 0")
	} else if t.JobStatus < 0 || t.JobStatus > 4 {
		return errors.New("status must be 0, 1, 2, 3 or 4")
	}
	return nil
}

var statusMap = map[int]string{
	0: "Pending",
	1: "Running",
	2: "Completed",
	3: "Failed",
	4: "ReEntered",
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
const JOBREENTERED = 4

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
const RUNNING_JOBS_RECORD = "running_tasks_record"
const TASKS_FULL_RECORD = "job_full_info"
