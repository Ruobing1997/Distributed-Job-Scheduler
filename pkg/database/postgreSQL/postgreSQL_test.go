package postgreSQL

import (
	"testing"
	"time"
)

func TestGetTasksInInterval(t *testing.T) {
	NewpostgreSQLClient()
	var db = GetDB()
	defer db.Close()
	// Insert some test data
	_, err := db.Exec(`INSERT INTO task_db (id, name, type, schedule, payload, callback_url, status, execution_time, next_execution_time, create_time, update_time, retries, result) VALUES
		('1', 'Task 1', 1, 'daily', 'payload1', 'https://callback1.com', 1, '2022-01-01 12:00:00', '2022-01-02 12:00:00', '2022-01-01 00:00:00', '2022-01-01 00:00:00', 0, 0),
		('2', 'Task 2', 2, 'weekly', 'payload2', 'https://callback2.com', 1, '2022-01-01 13:00:00', '2022-01-08 13:00:00', '2022-01-01 00:00:00', '2022-01-01 00:00:00', 0, 0)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	// Delete the test data when the test is finished
	defer func() {
		_, err := db.Exec("DELETE FROM task_db WHERE id IN ('1', '2')")
		if err != nil {
			t.Fatalf("Failed to delete test data: %v", err)
		}
	}()

	// Call the GetTasksInInterval function
	startTime := time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2022, 1, 1, 14, 0, 0, 0, time.UTC)
	timeTracker := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	tasks, err := GetTasksInInterval(db, startTime, endTime, timeTracker)

	if err != nil {
		t.Fatalf("Failed to get tasks in interval: %v", err)
	}

	// Check the number of tasks returned
	expectedTasks := 2
	if len(tasks) != expectedTasks {
		t.Errorf("Expected %d tasks, but got %d", expectedTasks, len(tasks))
	}

	// Check the task details
	if tasks[0].ID != "1" || tasks[0].Name != "Task 1" || tasks[0].Type != 1 {
		t.Errorf("Task 1 details are incorrect")
	}
	if tasks[1].ID != "2" || tasks[1].Name != "Task 2" || tasks[1].Type != 2 {
		t.Errorf("Task 2 details are incorrect")
	}
}
