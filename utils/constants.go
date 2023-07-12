package constants

import (
	"time"
)

type Task struct {
	ID          string
	Name        string
	Type        string
	Schedule    string
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

const SHARDSAMOUNT = 6
const EIGHTBIT = 8
const SIXTEENBIT = 16
const TWENTYFOURBIT = 24
