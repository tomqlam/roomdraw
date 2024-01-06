-- Random SuiteUID generate postgres uuid, Numeric Dorm, Dorm Name, Room count, Room UID Array, preplaced rooms.
--POstgres SQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE Suites (
    suite_uuid uuid DEFAULT UUID_GENERATE_V4(),
    dorm int NOT NULL,
    dorm_name varchar NOT NULL,
    floor int NOT NULL,
    room_count int NOT NULL,
    rooms uuid array,
    alternative_pull bool NOT NULL,
    suite_design varchar NOT NULL DEFAULT '',
    PRIMARY KEY (suite_uuid)
);