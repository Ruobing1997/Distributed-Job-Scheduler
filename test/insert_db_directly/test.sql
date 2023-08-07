DO $$
DECLARE
i INTEGER := 2000;
BEGIN
    WHILE i < 4000 LOOP
        INSERT INTO job_full_info(
            id,
            job_name,
            job_type,
            cron_expression,
            execute_format,
            execute_script,
            callback_url,
            job_status,
            execution_time,
            previous_execution_time,
            create_time,
            update_time,
            retries_left
        ) VALUES (
            'job_' || i::TEXT,              -- 使用i作为id的后缀
            'Job Name ' || i::TEXT,         -- 使用i作为job_name的后缀
            1,                              -- 假设的job_type
            '* * * * * *',                  -- 每秒执行一次的cron表达式
            0,                              -- execute_format为0
            'echo 1',                       -- 脚本内容
            '',                           -- callback_url，如果有可以填写
            1,                              -- 假设的job_status
            NOW() + interval '1 minute',    -- 当前时间1分钟后作为执行时间
            NOW(),                          -- 假设的上次执行时间为现在
            NOW() + interval '1 minute',    -- 创建时间
            NOW(),                          -- 更新时间
            3                               -- 假设的重试次数
        );
        i := i + 1;
END LOOP;
END $$;
