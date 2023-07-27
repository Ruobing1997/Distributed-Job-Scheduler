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
		var format int
		var script string
		err := c.db.QueryRowContext(ctx, query, id).Scan(&taskDB.ID, &taskDB.JobName, &taskDB.JobType, &taskDB.CronExpr,
			&format, &script, &taskDB.CallbackURL, &taskDB.Status, &taskDB.ExecutionTime,
			&taskDB.CreateTime, &taskDB.UpdateTime, &taskDB.Retries)
		if err != nil {
			return nil, err
		}
		taskDB.Payload = &constants.Payload{
			Format: format,
			Script: script,
		}
		return &taskDB, nil
	case constants.RUNNING_JOBS_RECORD:
		query := `SELECT * FROM running_tasks_record WHERE id = $1`
		var runtimeJobInfo constants.RunTimeTask
		var format int
		var script string
		err := c.db.QueryRowContext(ctx, query, id).Scan(&runtimeJobInfo.ID, &runtimeJobInfo.ExecutionTime, &runtimeJobInfo.JobType,
			&runtimeJobInfo.JobStatus, &format, &script, &runtimeJobInfo.RetriesLeft, &runtimeJobInfo.CronExpr)
		if err != nil {
			return nil, err
		}
		runtimeJobInfo.Payload = &constants.Payload{
			Format: format,
			Script: script,
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
		var taskDB constants.TaskDB
		var format int
		var script string
		err := rows.Scan(&taskDB.ID, &taskDB.JobName, &taskDB.JobType, &taskDB.CronExpr,
			&format, &script, &taskDB.CallbackURL, &taskDB.Status, &taskDB.ExecutionTime,
			&taskDB.CreateTime, &taskDB.UpdateTime, &taskDB.Retries)
		if err != nil {
			return nil, err
		}
		taskDB.Payload = &constants.Payload{
			Format: format,
			Script: script,
		}
		tasks = append(tasks, &taskDB)
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

	query := `update ` + table + ` SET ` + strings.Join(setValues, ", ") + ` WHERE id = $` + strconv.Itoa(i)
	values = append(values, id)
	_, err := c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) InsertUser(ctx context.Context, record databasehandler.DataRecord) error {
	user, ok := record.(*constants.UserInfo)
	if !ok {
		return errors.New("invalid data record")
	}
	query := `INSERT INTO users (username, password, role) VALUES ($1, $2, $3)`
	_, err := c.db.ExecContext(ctx, query, user.Username, user.Password, user.Role)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) IsValidCredential(ctx context.Context, username string, password string) (bool, error) {
	query := `SELECT * FROM users WHERE username = $1 AND password = $2`
	var user constants.UserInfo
	err := c.db.QueryRowContext(ctx, query, username, password).Scan(&user.Username, &user.Password, &user.Role)
	if err != nil {
		return false, err
	}
	return true, nil
}
