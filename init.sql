CREATE TABLE IF NOT EXISTS working_times (
    id SERIAL PRIMARY KEY,
    employee_id VARCHAR(255) NOT NULL,
    clock_in_time TIMESTAMP WITH TIME ZONE NOT NULL,
    clock_out_time TIMESTAMP WITH TIME ZONE,
    hours_worked NUMERIC,
    status VARCHAR(50) NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_working_times_employee_id ON working_times (employee_id);
