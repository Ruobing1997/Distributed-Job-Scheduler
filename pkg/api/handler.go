package api

import (
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type TaskRequest struct {
	JobName  string `json:"jobName"`
	JobType  int    `json:"jobType"`
	CronExpr string `json:"cronExpr"`
	Format   int    `json:"format"`
	Script   string `json:"script"`
	Retries  int    `json:"retries"`
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GenerateTaskHandler(c *gin.Context) (taskID string) {
	var taskRequest TaskRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return "nil"
	}

	jobName := taskRequest.JobName
	jobType := taskRequest.JobType
	cronExpr := taskRequest.CronExpr
	format := taskRequest.Format
	script := taskRequest.Script
	retries := taskRequest.Retries

	taskID, err := task_manager.HandleNewTasks(jobName, jobType, cronExpr, format, script, "", retries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return "nil"
	}
	c.JSON(http.StatusOK, gin.H{"status": "Successfully generated update"})
	return taskID
}

func DeleteTaskHandler(c *gin.Context) {
	taskID := c.Param("id")
	err := task_manager.HandleDeleteTasks(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusSeeOther, "/api/dashboard")
}

func UpdateTaskHandler(c *gin.Context) {
	taskID := c.Param("id")
	var taskRequest TaskRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobName := taskRequest.JobName
	jobType := taskRequest.JobType
	cronExpr := taskRequest.CronExpr
	format := taskRequest.Format
	script := taskRequest.Script
	retries := taskRequest.Retries

	log.Printf("jobName: %s, jobType: %d, cronExpr: %s, format: %d, script: %s, retries: %d",
		jobName, jobType, cronExpr, format, script, retries)

	updateVarsMap := map[string]interface{}{
		"job_name":        jobName,
		"job_type":        jobType,
		"cron_expression": cronExpr,
		"execute_format":  format,
		"execute_script":  script,
		"update_time":     time.Now(),
		"create_time":     time.Now(),
		"retries_left":    retries}
	err := task_manager.HandleUpdateTasks(
		taskID, updateVarsMap)

	if err != nil {
		log.Printf("Error updating task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Successfully updated update"})
}

func GetTaskHandler(c *gin.Context) {
	taskID := c.Param("id")
	task, err := task_manager.HandleGetTasks(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "detail_index.html", task)
}

func DashboardHandler(c *gin.Context) {
	tasks, err := task_manager.HandleGetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "dashboard_index.html", tasks)
}

func RunningTasksHandler(c *gin.Context) {
	tasks, err := task_manager.HandleGetRunningTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "running_tasks_dashboard_index.html", tasks)
}

func TaskHistoryHandler(c *gin.Context) {
	taskID := c.Param("id")
	tasks, err := task_manager.HandleGetTaskHistory(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "history_index.html", tasks)
}
