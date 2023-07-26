package mySQL

import (
	"context"
	"database/sql"
	"errors"
	databasehandler "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type Client struct {
	Db *sql.DB
}

// NewMySQLClient initializes the connection to MySQL database.
func NewMySQLClient() *Client {
	db, err := sql.Open(DBDRIVER, DBDATASOURCE)
	if err != nil {
		log.Printf("Error opening database: %s", err.Error())
	}
	return &Client{Db: db}
}

func (c *Client) InsertTask(ctx context.Context, record databasehandler.DataRecord) error {
	taskDB, ok := record.(*constants.TaskDB)
	if !ok {
		return errors.New("invalid data record")
	}
	query := `INSERT INTO task_db 
    (id, name, type, schedule, payload, callback_url, status, execution_time, 
     next_execution_time, create_time, update_time, retries, result) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := c.Db.ExecContext(ctx, query, taskDB.ID, taskDB.Name, taskDB.JobType,
		taskDB.CronExpr, taskDB.Payload, taskDB.CallbackURL,
		taskDB.Status, taskDB.ExecutionTime,
		taskDB.CreateTime, taskDB.UpdateTime, taskDB.Retries, taskDB.Result)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetTaskByID(ctx context.Context, id string, args ...interface{}) (databasehandler.DataRecord, error) {
	query := `SELECT * FROM task_db WHERE id = ?`
	taskDB := &constants.TaskDB{}
	err := c.Db.QueryRowContext(ctx, query, id).Scan(&taskDB.ID, &taskDB.Name, &taskDB.JobType, &taskDB.CronExpr,
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
	return c.Db.Close()
}

func GetDB() *sql.DB {
	db, err := sql.Open(DBDRIVER, DBDATASOURCE)
	if err != nil {
		log.Printf("Error opening database: %s", err.Error())
	}
	return db
}
