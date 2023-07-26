package postgreSQL

import (
	"context"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"testing"
	"time"
)

func setUpTestForTaskDB(t *testing.T, client *Client) {
	_, err := client.db.Exec(`INSERT INTO task_db (id, name, type, schedule, payload, callback_url, status, execution_time, next_execution_time, create_time, update_time, retries, result) VALUES
		('1', 'Task 1', 1, 'daily', 'payload1', 'https://callback1.com', 1, '2022-01-01 12:00:00', '2022-01-02 12:00:00', '2022-01-01 00:00:00', '2022-01-01 00:00:00', 0, 0),
		('2', 'Task 2', 2, 'weekly', 'payload2', 'https://callback2.com', 1, '2022-01-01 13:00:00', '2022-01-08 13:00:00', '2022-01-01 00:00:00', '2022-01-01 00:00:00', 0, 0)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
}

func setUpTestForExecutionRecord(t *testing.T, client *Client) {
	_, err := client.db.Exec(`INSERT INTO execution_record (id, job_type, job_status, retries_left) VALUES 
		('1', 1, 1, 0), ('2', 2, 1, 0)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
}

func TestGetTasksInInterval(t *testing.T) {
	client := NewpostgreSQLClient()
	setUpTestForTaskDB(t, client)

	// Call the GetTasksInInterval function
	startTime := time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2022, 1, 1, 14, 0, 0, 0, time.UTC)
	timeTracker := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	tasks, err := GetTasksInInterval(client.db, startTime, endTime, timeTracker)

	if err != nil {
		t.Fatalf("Failed to get tasks in interval: %v", err)
	}

	// Check the number of tasks returned
	expectedTasks := 2
	if len(tasks) != expectedTasks {
		t.Errorf("Expected %d tasks, but got %d", expectedTasks, len(tasks))
	}

	// Check the task details
	if tasks[0].ID != "1" || tasks[0].Name != "Task 1" || tasks[0].JobType != 1 {
		t.Errorf("Task 1 details are incorrect")
	}
	if tasks[1].ID != "2" || tasks[1].Name != "Task 2" || tasks[1].JobType != 2 {
		t.Errorf("Task 2 details are incorrect")
	}

	defer func() {
		_, err = client.db.Exec("DELETE FROM task_db WHERE id IN ('1', '2')")
		if err != nil {
			t.Fatalf("Failed to delete test data: %v", err)
		}
	}()
}

func TestGetRuntimeJobInfoByID(t *testing.T) {
	client := NewpostgreSQLClient()
	setUpTestForExecutionRecord(t, client)
	runtimeTaskInfo1, err := client.GetRuntimeJobInfoByID("1")
	if err != nil {
		t.Fatalf("Failed to get runtime job info by ID: %v", err)
	}
	if runtimeTaskInfo1.ID != "1" || runtimeTaskInfo1.JobType != 1 || runtimeTaskInfo1.JobStatus != 1 || runtimeTaskInfo1.RetriesLeft != 0 {
		t.Errorf("Runtime job info is incorrect")
	}
	runtimeTaskInfo2, err := client.GetRuntimeJobInfoByID("2")
	if err != nil {
		t.Fatalf("Failed to get runtime job info by ID: %v", err)
	}
	if runtimeTaskInfo2.ID != "2" || runtimeTaskInfo2.JobType != 2 || runtimeTaskInfo2.JobStatus != 1 || runtimeTaskInfo2.RetriesLeft != 0 {
		t.Errorf("Runtime job info is incorrect")
	}

	defer func() {
		_, err = client.db.Exec("DELETE FROM execution_record WHERE id IN ('1', '2')")
		if err != nil {
			t.Fatalf("Failed to delete test data: %v", err)
		}
	}()
}

func TestGetTaskByID(t *testing.T) {
	client := NewpostgreSQLClient()
	setUpTestForTaskDB(t, client)
	taskInfo1, err := client.GetTaskByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("Failed to get task info by ID: %v", err)
	}

	task := taskInfo1.(*constants.TaskDB)
	if task.ID != "1" || task.Name != "Task 1" {
		t.Errorf("Task info is incorrect")
	}

	defer func() {
		_, err = client.db.Exec("DELETE FROM task_db WHERE id IN ('1', '2')")
		if err != nil {
			t.Fatalf("Failed to delete test data: %v", err)
		}
	}()
}
