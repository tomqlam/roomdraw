-- Columns: USER_ID(unique) claremontid, Year, DrawNumber, Preplaced(bool), In-Dorm (numbers for which dorm 0: None, 1:East etc.), groupID
-- POSTGRES SQL
CREATE TABLE Users (
    id serial,
    year varchar NOT NULL,
    first_name varchar NOT NULL,
    last_name varchar NOT NULL,
    email varchar,
    draw_number decimal NOT NULL,
    preplaced boolean NOT NULL,
    in_dorm int NOT NULL,
    sgroup_uuid uuid,
    participated boolean NOT NULL DEFAULT false,
    participation_time timestamp,
    room_uuid uuid,
    reslife_role varchar NOT NULL DEFAULT 'none',
    notifications_enabled boolean NOT NULL DEFAULT false,
    notification_created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    notification_updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    gender_preferences varchar[] NOT NULL DEFAULT '{}',
    
    PRIMARY KEY (id),
    FOREIGN KEY (sgroup_uuid) REFERENCES SuiteGroups(sgroup_uuid)
);