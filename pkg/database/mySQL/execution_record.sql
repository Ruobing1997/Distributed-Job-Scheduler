USE super_nova_test_1;

CREATE TABLE IF NOT Exists execution_record (
    job_id varchar(255) NOT NULL,
    execution_time datetime NOT NULL,
    worker_ip varchar(255) NOT NULL,
    job_status varchar(255) NOT NULL
)