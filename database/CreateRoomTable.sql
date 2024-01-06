-- Random RoomUID, Numeric Dorm, Dorm Name, Room ID (can contain letters), IsInSuite, SuiteID, Max Occupancy, Current Occupancy, Occupant Array (USER_ID), isLockPulled
-- Postgres SQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE Rooms (
    room_uuid uuid DEFAULT UUID_GENERATE_V4(),
    dorm int NOT NULL,
    dorm_name varchar NOT NULL,
    room_id varchar NOT NULL,
    suite_uuid uuid NOT NULL,
    max_occupancy int NOT NULL,
    current_occupancy int NOT NULL,
    occupants int array,
    pull_priority jsonb NOT NULL DEFAULT '{
        "valid": false,
        "isPreplaced": false,
        "hasInDorm": false,
        "drawNumber": 0,
        "year": 0,
        "pullType": 0,
        "inherited": {
            "hasInDorm": false,
            "drawNumber": 0,
            "year": 0
        }
    }'::jsonb,
    PRIMARY KEY (room_uuid),
    FOREIGN KEY (suite_uuid) REFERENCES Suites(suite_uuid)
);
