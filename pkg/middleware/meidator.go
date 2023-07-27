package middleware

import (
	"fmt"
	"time"
)

var renewLeaseFunc func(taskID string, newLeaseTime time.Time) (bool, error)

func SetRenewLeaseFunction(f func(taskID string, newLeaseTime time.Time) (bool, error)) {
	renewLeaseFunc = f
}

func RenewLeaseThroughMediator(taskID string, newLeaseTime time.Time) (bool, error) {
	if renewLeaseFunc != nil {
		return renewLeaseFunc(taskID, newLeaseTime)
	}
	return false, fmt.Errorf("renew lease function not set")
}
