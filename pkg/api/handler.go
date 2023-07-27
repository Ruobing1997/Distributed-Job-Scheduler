package api

import (
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/gin-gonic/gin"
	"net/http"
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
	c.JSON(http.StatusOK, gin.H{"status": "Successfully deleted update"})
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

	updateVarsMap := map[string]interface{}{
		"jobName":  jobName,
		"jobType":  jobType,
		"cronExpr": cronExpr,
		"format":   format,
		"script":   script,
		"retries":  retries}
	err := task_manager.HandleUpdateTasks(
		taskID, updateVarsMap)

	if err != nil {
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
	c.JSON(http.StatusOK, task)
}

func RegisterUserHandler(c *gin.Context) {
	var userRequest UserRequest
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := userRequest.Username
	password := userRequest.Password

	var currentUser *constants.UserInfo
	currentUser.Username = username
	currentUser.Password = password
	if currentUser.Username == "admin" &&
		currentUser.Password == "admin" {
		currentUser.Role = 1
	} else {
		currentUser.Role = 0
	}
	err := task_manager.RegisterUser(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Successfully registered user"})
}

func LoginUserHandler(c *gin.Context) {
	var userRequest UserRequest
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := userRequest.Username
	password := userRequest.Password
	isValid, err := task_manager.LoginUser(username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isValid {
		c.JSON(http.StatusOK, gin.H{"status": "Successfully logged in"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid username or password"})
	}
}
