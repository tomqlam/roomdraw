-- Table to track clear room usage rate limits and user blocklisting
CREATE TABLE user_rate_limits (
    email varchar PRIMARY KEY,
    clear_room_count int NOT NULL DEFAULT 0,
    clear_room_date date DEFAULT CURRENT_DATE, -- Store the date for the current count
    is_blocklisted boolean NOT NULL DEFAULT false,
    blocklisted_at timestamp WITH TIME ZONE,
    blocklisted_reason varchar
);

-- Index for faster lookups
CREATE INDEX idx_user_rate_limits_email ON user_rate_limits(email);
