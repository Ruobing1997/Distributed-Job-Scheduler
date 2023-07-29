
CREATE TABLE IF NOT EXISTS job_full_info (
    id VARCHAR(255) PRIMARY KEY,
    job_name VARCHAR(255) NOT NULL,
    job_type INTEGER NOT NULL,
    cron_expr VARCHAR(255),
    execute_format INTEGER NOT NULL,
    execute_script VARCHAR(255),
    callback_url VARCHAR(255),
    status INTEGER NOT NULL,
    execution_time TIMESTAMP,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP NOT NULL,
    retries INTEGER NOT NULL
    );

CREATE INDEX IF NOT EXISTS task_db_execution_time_brin_idx ON job_full_info USING BRIN (execution_time);
CREATE INDEX IF NOT EXISTS task_db_create_time_brin_idx ON job_full_info USING BRIN (create_time);

CREATE TABLE IF NOT EXISTS running_tasks_record
(
    id           VARCHAR(255) PRIMARY KEY,
    execution_time TIMESTAMP NOT NULL,
    job_type     int NOT NULL,
    job_status   int NOT NULL,
    execute_format INTEGER NOT NULL,
    execute_script VARCHAR(255),
    retries_left int NOT NULL,
    cron_expression VARCHAR(255)
    );

CREATE INDEX IF NOT EXISTS execution_record_job_status_idx ON running_tasks_record (job_status);