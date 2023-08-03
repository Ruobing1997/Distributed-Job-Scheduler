package task_manager

import "time"

const DURATION = time.Minute * 10
const REDIS_EXPIRY_CHANNEL = "__keyevent@0__:expired"
const WORKER_SERVICE = "worker-service"
