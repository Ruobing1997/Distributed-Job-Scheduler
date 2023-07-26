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
