-- Table to store transaction logs for database modifications
CREATE TABLE transaction_logs (
    log_id SERIAL PRIMARY KEY,
    operation_type VARCHAR(50) NOT NULL,       -- e.g., "UPDATE_ROOM_OCCUPANTS", "CLEAR_ROOM", "PREPLACE_OCCUPANTS"
    endpoint VARCHAR(255) NOT NULL,            -- API endpoint that was called
    user_email VARCHAR(255) NOT NULL,          -- Email of the user who performed the action
    user_name VARCHAR(255),                    -- Name of the user (if available)
    entity_type VARCHAR(50) NOT NULL,          -- e.g., "ROOM", "USER", "SUITE"
    entity_id VARCHAR(100) NOT NULL,           -- ID of the affected entity (room_uuid, user_id, suite_uuid etc.) - Use VARCHAR for flexibility (UUIDs, IDs)
    previous_state JSONB,                      -- State before the change (can be null for creates)
    new_state JSONB,                           -- State after the change (can be null for deletes)
    details JSONB,                             -- Additional details about the operation
    ip_address VARCHAR(45),                    -- IP address of the client
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    request_id UUID                            -- Optional: to group related operations in a single request
);

-- Create indexes for efficient querying
CREATE INDEX idx_transaction_logs_operation_type ON transaction_logs(operation_type);
CREATE INDEX idx_transaction_logs_user_email ON transaction_logs(user_email);
CREATE INDEX idx_transaction_logs_entity_type ON transaction_logs(entity_type);
CREATE INDEX idx_transaction_logs_entity_id ON transaction_logs(entity_id);
CREATE INDEX idx_transaction_logs_created_at ON transaction_logs(created_at);
CREATE INDEX idx_transaction_logs_request_id ON transaction_logs(request_id); -- Index for request_id if used frequently