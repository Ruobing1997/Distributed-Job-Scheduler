package constants

import (
	"time"
)

type Task struct {
	ID          string
	Name        string
	Type        int
	Schedule    string
	Importance  float64
	Payload     string
	CallbackURL string
	Status      int
	CreateTime  time.Time
	UpdateTime  time.Time
	Retries     int
	Result      string
}

var statusMap = map[int]string{
	0: "pending",
	1: "running",
	2: "completed",
	3: "failed",
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
