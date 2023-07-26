
CREATE TABLE IF NOT EXISTS execution_record
(
    id           VARCHAR(255) PRIMARY KEY,
    job_type     int NOT NULL,
    job_status   int NOT NULL,
    retries_left int NOT NULL
);

CREATE INDEX IF NOT EXISTS execution_record_job_status_idx ON execution_record (job_status);
