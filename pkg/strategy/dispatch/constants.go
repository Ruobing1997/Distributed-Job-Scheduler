package dispatch

const WORKER_SERVICE_URL = "http://worker-service.default.svc.cluster.local"
const INVERTED_JSON_K8S_SERVICE_URL = "http://inverted-json-service.default.svc.cluster.local"
const TASK_CHANNEL = "/task-channel"
const APPLICATION_JSON = "application/json"
const POST = "POST"

const (
	JobFailed        = 0
	JobDispatched    = 1
	JobSucceed       = 2
	WorkerNoResponse = 3
)
