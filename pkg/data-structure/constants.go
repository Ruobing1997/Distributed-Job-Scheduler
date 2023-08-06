package data_structure_redis

import "time"

const REDISPQ_Addr = "localhost:6379"
const REDISPQ_Password = ""
const REDISPQ_DB = 0
const REDIS_PQ_KEY = "task_queue"
const REDIS_RETRY_KEY = "retry_task_set"
const REDIS_MAP_KEY = "task_map"
const REDIS_CHANNEL = "near_execution_tasks"
const RETRY_AVAILABLE = "retry_available"

const REDIS_LEASE_MAP = "task_lease_map"
const REDIS_LEASE_CHANNEL = "task_lease_channel"
const REDIS_LEASE_MAP_VALUE_PROCESSING = "1"

// TODO: add ahead time after testing
const PROXIMITY_THRESHOLD = 2 * time.Second
const DISPATCHBUFFER = 2 * time.Second
