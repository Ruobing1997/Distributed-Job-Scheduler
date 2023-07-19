package postgreSQL

import (
	"context"
	"database/sql"
	"errors"
	databasehandler "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Client struct {
	db *sql.DB
}

// NewMySQLClient initializes the connection to MySQL database.
func NewpostgreSQLClient() *Client {
	db, err := sql.Open(DBDRIVER, connStr)
	if err != nil {
		log.Printf("Error opening database: %s", err.Error())
	}
	return &Client{db: db}
}

func (c *Client) InsertTask(ctx context.Context, record databasehandler.DataRecord) error {
	taskDB, ok := record.(*constants.TaskDB)
	if !ok {
		return errors.New("invalid data record")
	}
	query := `INSERT INTO task_db 
        (id, name, type, schedule, payload, callback_url, status, execution_time, 
         next_execution_time, create_time, update_time, retries, result) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := c.db.ExecContext(ctx, query, taskDB.ID, taskDB.Name, taskDB.Type,
		taskDB.Schedule, taskDB.Payload, taskDB.CallbackURL,
		taskDB.Status, taskDB.ExecutionTime,
		taskDB.CreateTime, taskDB.UpdateTime, taskDB.Retries, taskDB.Result)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetTaskByID(ctx context.Context, id string, args ...interface{}) (databasehandler.DataRecord, error) {
	query := `SELECT * FROM task_db WHERE id = $1`
	taskDB := &constants.TaskDB{}
	err := c.db.QueryRowContext(ctx, query, id).Scan(&taskDB.ID, &taskDB.Name, &taskDB.Type, &taskDB.Schedule,
		&taskDB.Payload, &taskDB.CallbackURL, &taskDB.Status, &taskDB.ExecutionTime, &taskDB.NextExecutionTime,
		&taskDB.CreateTime, &taskDB.UpdateTime, &taskDB.Retries, &taskDB.Result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return taskDB, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func GetTasksInInterval(db *sql.DB, startTime time.Time, endTime time.Time) ([]*constants.TaskDB, error) {
	query := `SELECT * FROM task_db WHERE execution_time BETWEEN $1 AND $2`
	rows, err := db.Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*constants.TaskDB
	for rows.Next() {
		var task constants.TaskDB
		err := rows.Scan(&task.ID, &task.Name, &task.Type, &task.Schedule, &task.Payload, &task.CallbackURL, &task.Status,
			&task.ExecutionTime, &task.NextExecutionTime, &task.CreateTime, &task.UpdateTime, &task.Retries, &task.Result)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func GetDB() *sql.DB {
	db, err := sql.Open(DBDRIVER, connStr)
	if err != nil {
		log.Printf("Error opening database: %s", err.Error())
	}
	return db
}
