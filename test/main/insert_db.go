package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=9.135.75.241 port=5432 user=postgres_admin password=JjKQt@8254KpkY dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Prepare the SQL statement
	stmt, err := db.Prepare(`INSERT INTO job_full_info 
		(id, job_name, job_type, cron_expression, execute_format, execute_script, 
		callback_url, job_status, execution_time, previous_execution_time, 
		create_time, update_time, retries_left) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Insert 5000 tasks
	for i := 1; i <= 2; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		jobName := fmt.Sprintf("Job Number %d", i)
		jobType := 0            // Example job type
		cronExpr := "* * * * *" // Every second
		executeFormat := 1      // Example format
		executeScript := "print('Hello World!')"
		callbackURL := "http://your.callback.url"
		jobStatus := 1 // Example status
		executionTime := time.Now().Add(2 * time.Minute)
		previousExecutionTime := time.Now().Add(-1 * time.Second)
		createTime := time.Now()
		updateTime := time.Now()
		retriesLeft := 3 // Example retries

		_, err := stmt.Exec(jobID, jobName, jobType, cronExpr, executeFormat, executeScript,
			callbackURL, jobStatus, executionTime, previousExecutionTime,
			createTime, updateTime, retriesLeft)
		if err != nil {
			log.Printf("Failed to insert job %d: %v", i, err)
		}
	}

	fmt.Println("Inserted 5000 tasks successfully!")
}
