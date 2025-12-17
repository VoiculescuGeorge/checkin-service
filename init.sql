CREATE TABLE working_times (
    id BIGSERIAL PRIMARY KEY,
    employee_id VARCHAR(50) NOT NULL,
    clock_in_time TIMESTAMP NOT NULL,
    clock_out_time TIMESTAMP,
    hours_worked NUMERIC(5, 2),
    labor_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    email_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    labor_retry_count INT NOT NULL DEFAULT 0,
    email_retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_labor_pending ON working_times(labor_status) WHERE labor_status = 'PENDING';
CREATE INDEX idx_email_pending ON working_times(email_status) WHERE email_status = 'PENDING';