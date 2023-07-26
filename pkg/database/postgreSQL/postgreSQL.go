package postgreSQL

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	databasehandler "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
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

func (c *Client) InsertTask(ctx context.Context, table string, record databasehandler.DataRecord) error {
	var err error
	switch table {
	case constants.TASKS_FULL_RECORD:
		taskDB, ok := record.(*constants.TaskDB)
		if !ok {
			return errors.New("invalid data record")
		}
		query := `INSERT INTO job_full_info 
        (id, job_name, job_type, cron_expr, execute_format, execute_script, callback_url, status, execution_time, 
         create_time, update_time, retries) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
		_, err = c.db.ExecContext(ctx, query, taskDB.ID, taskDB.JobName, taskDB.JobType,
			taskDB.CronExpr, taskDB.Payload.Format, taskDB.Payload.Script, taskDB.CallbackURL,
			taskDB.Status, taskDB.ExecutionTime,
			taskDB.CreateTime, taskDB.UpdateTime, taskDB.Retries)
	case constants.RUNNING_JOBS_RECORD:
		runTimeTask, ok := record.(*constants.RunTimeTask)
		if !ok {
			return errors.New("invalid data record")
		}
		query := `INSERT INTO running_tasks_record (id, execution_time, job_type, job_status,
                                  execute_format, execute_script, retries_left, cron_expression) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err = c.db.ExecContext(ctx, query,
			runTimeTask.ID,
			runTimeTask.ExecutionTime,
			runTimeTask.JobType,
			runTimeTask.JobStatus,
			runTimeTask.Payload.Format,
			runTimeTask.Payload.Script,
			runTimeTask.RetriesLeft,
			runTimeTask.CronExpr)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetTaskByID(ctx context.Context, table string, id string, args ...interface{}) (databasehandler.DataRecord, error) {
	switch table {
	case constants.TASKS_FULL_RECORD:
		query := `SELECT * FROM job_full_info WHERE id = $1`
		taskDB := constants.TaskDB{}
		err := c.db.QueryRowContext(ctx, query, id).Scan(&taskDB.ID, &taskDB.JobName, &taskDB.JobType, &taskDB.CronExpr,
			&taskDB.Payload, &taskDB.CallbackURL, &taskDB.Status, &taskDB.ExecutionTime,
			&taskDB.CreateTime, &taskDB.UpdateTime, &taskDB.Retries)
		if err != nil {
			return nil, err
		}
		return &taskDB, nil
	case constants.RUNNING_JOBS_RECORD:
		query := `SELECT * FROM job_full_info WHERE id = $1`
		var runtimeJobInfo constants.RunTimeTask
		err := c.db.QueryRowContext(ctx, query, id).Scan(&runtimeJobInfo.ID, &runtimeJobInfo.JobType,
			&runtimeJobInfo.JobStatus, &runtimeJobInfo.RetriesLeft)
		if err != nil {
			return nil, err
		}
		return &runtimeJobInfo, nil
	}
	return nil, errors.New("invalid table name")
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) GetTasksInInterval(startTime time.Time, endTime time.Time, timeTracker time.Time) ([]*constants.TaskDB, error) {
	query := `SELECT * FROM job_full_info WHERE execution_time BETWEEN $1 AND $2 AND create_time >= $3`
	rows, err := c.db.Query(query, startTime, endTime, timeTracker)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*constants.TaskDB
	for rows.Next() {
		var task constants.TaskDB
		var payload constants.Payload
		err := rows.Scan(&task.ID, &task.JobName, &task.JobType, &task.CronExpr,
			&payload.Format, &payload.Script, &task.CallbackURL, &task.Status,
			&task.ExecutionTime, &task.CreateTime, &task.UpdateTime, &task.Retries)
		if err != nil {
			return nil, err
		}
		task.Payload = &payload
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *Client) DeleteByID(ctx context.Context, table string, id string) error {
	query := `DELETE ` + table + ` WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting execution record: %s", err.Error())
	}
	return nil
}

func (c *Client) UpdateByID(ctx context.Context, table string, id string, args map[string]interface{}) error {
	var setValues []string
	var values []interface{}
	i := 1
	for k, v := range args {
		setValues = append(setValues, fmt.Sprintf("%s = $%d", k, i))
		values = append(values, v)
		i++
	}

	query := `Update ` + table + ` SET ` + strings.Join(setValues, ", ") + ` WHERE id = $` + strconv.Itoa(i)
	values = append(values, id)
	_, err := c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}
	return nil
}
