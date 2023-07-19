CREATE INDEX IF NOT EXISTS task_db_execution_time_brin_idx ON task_db USING BRIN (execution_time);

CREATE TABLE IF NOT EXISTS task_db (
   id VARCHAR(255) PRIMARY KEY,
   name VARCHAR(255) NOT NULL,
   type INTEGER NOT NULL,
   schedule VARCHAR(255),
   payload TEXT,
   callback_url VARCHAR(255),
   status INTEGER NOT NULL,
   execution_time TIMESTAMP,
   next_execution_time TIMESTAMP,
   create_time TIMESTAMP NOT NULL,
   update_time TIMESTAMP NOT NULL,
   retries INTEGER NOT NULL,
   result INTEGER NOT NULL
);