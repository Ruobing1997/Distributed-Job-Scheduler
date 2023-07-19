// Package dispatch
// /* Use the invertedjson method as the strategy for dispatching the request.
package dispatch

import (
	"encoding/json"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
)

func ConvertCacheToJson(task *constants.TaskCache) ([]byte, error) {
	result, err := json.Marshal(task)
	return result, err
}
