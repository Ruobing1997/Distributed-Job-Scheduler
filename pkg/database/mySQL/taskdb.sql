USE super_nova_test_1;
CREATE TABLE task_db (
     id            VARCHAR(255) PRIMARY KEY,
     name          VARCHAR(255) NOT NULL,
     type          INT NOT NULL,
     schedule      VARCHAR(255),
     payload       TEXT,
     callback_url  VARCHAR(255),
     status        INT NOT NULL,
     execution_time DATETIME,
     next_execution_time DATETIME,
     create_time   DATETIME NOT NULL,
     update_time   DATETIME NOT NULL,
     retries       INT NOT NULL,
     result        TEXT
);