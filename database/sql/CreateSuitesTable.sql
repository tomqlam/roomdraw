-- Random SuiteUID generate postgres uuid, Numeric Dorm, Dorm Name, Room count, Room UID Array, preplaced rooms.
--POstgres SQL

CREATE TABLE Suites (
    suite_uuid uuid NOT NULL,
    dorm int NOT NULL,
    dorm_name varchar NOT NULL,
    floor int NOT NULL,
    room_count int NOT NULL,
    rooms uuid array,
    alternative_pull bool NOT NULL,
    suite_design varchar NOT NULL DEFAULT '',
    can_lock_pull bool NOT NULL DEFAULT false,
    lock_pulled_room uuid,
    reslife_room uuid,
    gender_preferences varchar[] NOT NULL DEFAULT '{}',
    can_be_gender_preferenced bool NOT NULL DEFAULT false,
    PRIMARY KEY (suite_uuid)
);