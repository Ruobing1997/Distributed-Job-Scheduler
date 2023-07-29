package postgreSQL

import (
	"context"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInsertTask(t *testing.T) {
	// 创建模拟数据库连接和 mock 对象
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// 使用模拟数据库创建客户端
	client := &Client{db: db}

	t.Run("insert into job_full_info table", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO job_full_info").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		taskDB := &constants.TaskDB{
			ID:            "testID",
			JobName:       "testJob",
			JobType:       1,
			CronExpr:      "*/5 * * * *",
			Payload:       &constants.Payload{Format: 1, Script: "testScript"},
			CallbackURL:   "testCallback",
			Status:        0,
			ExecutionTime: time.Now(),
			CreateTime:    time.Now(),
			UpdateTime:    time.Now(),
			Retries:       3,
		}
		err := client.InsertTask(context.Background(), constants.TASKS_FULL_RECORD, taskDB)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert into running_tasks_record table", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO running_tasks_record").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		runTimeTask := &constants.RunTimeTask{
			ID:            "testID",
			ExecutionTime: time.Now(),
			JobType:       1,
			JobStatus:     0,
			Payload:       &constants.Payload{Format: 1, Script: "testScript"},
			RetriesLeft:   3,
			CronExpr:      "*/5 * * * *",
		}
		err := client.InsertTask(context.Background(), constants.RUNNING_JOBS_RECORD, runTimeTask)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid data record for job_full_info table", func(t *testing.T) {
		err := client.InsertTask(context.Background(), constants.TASKS_FULL_RECORD, &constants.RunTimeTask{})
		assert.Error(t, err)
		assert.Equal(t, "invalid data record", err.Error())
	})

	t.Run("invalid data record for running_tasks_record table", func(t *testing.T) {
		err := client.InsertTask(context.Background(), constants.RUNNING_JOBS_RECORD, &constants.TaskDB{})
		assert.Error(t, err)
		assert.Equal(t, "invalid data record", err.Error())
	})

}

func TestGetTaskByIDFORJOBFULLINFO(t *testing.T) {
	// 创建模拟数据库和模拟对象
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	client := &Client{
		db: db,
	}

	// 模拟数据库查询的返回值
	executionTime, _ := time.Parse("2006-01-02 15:04:05", "2022-01-01 12:00:00")
	createTime, _ := time.Parse("2006-01-02 15:04:05", "2022-01-01 11:00:00")
	updateTime, _ := time.Parse("2006-01-02 15:04:05", "2022-01-01 13:00:00")
	rows := sqlmock.NewRows([]string{"id", "job_name", "job_type", "cron_expr", "execute_format", "execute_script", "callback_url", "status", "execution_time", "create_time", "update_time", "retries"}).
		AddRow("1", "testJob", 0, "* * * * *", 0, "print('Hello world')", "http://localhost/callback", 1, executionTime, createTime, updateTime, 3)

	mock.ExpectQuery(`SELECT \* FROM job_full_info WHERE id = \$1`).WithArgs("1").WillReturnRows(rows)

	// 调用 GetTaskByID 函数
	record, err := client.GetTaskByID(context.Background(), constants.TASKS_FULL_RECORD, "1")
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	taskDB, ok := record.(*constants.TaskDB)
	if !ok {
		t.Fatalf("Expected TaskDB type, but got different type")
	}

	// 使用 testify 断言库检查返回的值
	assert.Equal(t, "testJob", taskDB.JobName)
	assert.Equal(t, 0, taskDB.JobType)

	// 确保没有意外的数据库调用
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTaskByIDFORRUNNINGTASKSRECORD(t *testing.T) {
	// 创建模拟数据库和模拟对象
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	defer db.Close()

	client := &Client{
		db: db,
	}

	// 模拟数据库查询的返回值
	executionTime, _ := time.Parse("2006-01-02 15:04:05", "2022-01-01 12:00:00")
	rows := sqlmock.NewRows([]string{"id", "execution_time", "job_type", "job_status", "execute_format", "execute_script", "retries_left", "cron_expression"}).
		AddRow("1", executionTime, 1, 0, 1, "print('Hello world')", 3, "* * * * *")

	mock.ExpectQuery(`SELECT \* FROM running_tasks_record WHERE id = \$1`).WithArgs("1").WillReturnRows(rows)

	// 调用 GetTaskByID 函数
	record, err := client.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, "1")
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	runTimeTask, ok := record.(*constants.RunTimeTask)
	if !ok {
		t.Fatalf("Expected TaskDB type, but got different type")
	}

	// 使用 testify 断言库检查返回的值
	assert.Equal(t, "1", runTimeTask.ID)
	assert.Equal(t, 1, runTimeTask.JobType)

	// 确保没有意外的数据库调用
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
