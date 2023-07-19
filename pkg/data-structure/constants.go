package data_structure_redis

import "time"

const REDISPQ_Addr = "localhost:6379"
const REDISPQ_Password = ""
const REDISPQ_DB = 0
const REDIS_PQ_KEY = "task_queue"
const REDIS_MAP_KEY = "task_map"
const REDIS_CHANNEL = "near_execution_tasks"
const TASK_AVAILABLE = "task_available"

// TODO: add ahead time after testing
const PROXIMITY_THRESHOLD = time.Second * 10
const DISPATCHBUFFER = 0 // 100ms ahead to dispatch tasks
