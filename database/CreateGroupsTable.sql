-- - Group table
-- 	GroupUID, PeopleInGroup, USER_ID array
-- If group leader leaves a room voluntarily, new groupleader will be derived from the array. However, if the group leader leaving leaves the group to become 1 person, the group is disbanded (user table looks for group id that is disbanded and removes it from all relevant records)
-- if group leader is bumped, group is disbanded and everyone is bumped.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE SuiteGroups (
    sgroup_uuid uuid DEFAULT UUID_GENERATE_V4(),
    sgroup_size int NOT NULL,
    sgroup_name varchar NOT NULL,
    sgroup_suite uuid NOT NULL,
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
    disbanded boolean NOT NULL DEFAULT false,
    rgroups uuid array,
    PRIMARY KEY (sgroup_uuid),
    FOREIGN KEY (sgroup_suite) REFERENCES Suites(suite_uuid)
);
