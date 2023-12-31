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
	Status                int // pending, completed, failed
	ExecutionTime         time.Time
	PreviousExecutionTime time.Time
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
	ExecutionID   string
}

type Payload struct {
	Format int    `json:"format"`
	Script string `json:"script"`
}

type RunTimeTask struct {
	ID            string
	ExecutionTime time.Time
	JobType       int
	JobStatus     int // dispatched, running
	Payload       *Payload
	RetriesLeft   int
	CronExpr      string
	WorkerID      string
	ExecutionID   string
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int    `json:"role"`
}

type TaskIDExecIDMap struct {
	TaskID      string
	ExecutionID string
}

func (t *TaskIDExecIDMap) Validate() error {
	if t.TaskID == "" {
		return errors.New("task id cannot be empty")
	} else if t.ExecutionID == "" {
		return errors.New("execution id cannot be empty")
	}
	return nil
}

func (t *UserInfo) Validate() error {
	if t.Username == "" {
		return errors.New("username cannot be empty")
	} else if t.Password == "" {
		return errors.New("password cannot be empty")
	}
	return nil
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

var StatusMap = map[int]string{
	0: "Pending",
	1: "Running",
	2: "Completed",
	3: "Failed",
	4: "ReEntered",
	5: "Retrying",
	6: "Dispatched",
}

var resultMap = map[int]string{
	0: "In Queue",
	1: "Success",
	2: "Failed",
}

var TypeMap = map[int]string{
	0: "OneTime",
	1: "Recurring",
}

var PayloadTypeMap = map[int]string{
	0: "shell",
	1: "python",
	2: "email",
}

const DOMAIN = "localhost"

const JOBRUNNING = 1
const JOBSUCCEED = 2
const JOBFAILED = 3
const JOBREENTERED = 4
const JOBRETRYING = 5
const JOBDISPATCHED = 6

const SHELL = 0
const PYTHON = 1
const EMAIL = 2

const OneTime = 0
const Recurring = 1

const MONGOURI = "mongodb://localhost:27017"

const EXECUTE_TASK_GRPC_TIMEOUT = time.Minute
const RENEW_LEASE_GRPC_TIMEOUT = time.Second

const LEASE_DURATION = 2 * time.Second
const LEASE_RENEW_INTERVAL = 5 * time.Second // renew lease every 1 second
const BATCH_SIZE = 5000
const TIMEOUT = 10 * time.Second
const RUNNING_JOBS_RECORD = "running_tasks_record"
const TASKS_FULL_RECORD = "job_full_info"
const TASKID_TO_EXECID = "taskid_execid_mapping"
const ONE_TIME_JOB_RETRY_TIME = 2 * time.Second
const RECURRING_JOB_RETRY_TIME = 2 * time.Second
