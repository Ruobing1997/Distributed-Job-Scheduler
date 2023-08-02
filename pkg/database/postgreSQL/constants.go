package postgreSQL

const connStr = "host=localhost port=5432 user=postgres password=970409 dbname=postgres sslmode=require"

const DBDRIVER = "postgres"

const (
	InsertOrUpdateTaskFullInfo = `
		INSERT INTO job_full_info 
			(id, job_name, job_type, cron_expr, execute_format, execute_script, callback_url, status,
			 execution_time, previous_execution_time, 
			 create_time, update_time, retries) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) 
		DO UPDATE SET 
			job_name = EXCLUDED.job_name,
			job_type = EXCLUDED.job_type,
			cron_expr = EXCLUDED.cron_expr,
			execute_format = EXCLUDED.execute_format,
			execute_script = EXCLUDED.execute_script,
			callback_url = EXCLUDED.callback_url,
			status = EXCLUDED.status,
			execution_time = EXCLUDED.execution_time,
			previous_execution_time = EXCLUDED.previous_execution_time,
			create_time = EXCLUDED.create_time,
			update_time = EXCLUDED.update_time,
			retries = EXCLUDED.retries
	`
	InsertOrUpdateRunningTask = `
		INSERT INTO running_tasks_record 
			(id, execution_time, job_type, job_status, execute_format, execute_script, retries_left, cron_expression, worker_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id)
		DO UPDATE SET 
			execution_time = EXCLUDED.execution_time,
			job_type = EXCLUDED.job_type,
			job_status = EXCLUDED.job_status,
			execute_format = EXCLUDED.execute_format,
			execute_script = EXCLUDED.execute_script,
			retries_left = EXCLUDED.retries_left,
			cron_expression = EXCLUDED.cron_expression,
			worker_id = EXCLUDED.worker_id
	`
)
