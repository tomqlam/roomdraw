-- - Bump table
-- 	BumpUID, GroupUID bumped, UserID bumped, USER_ID of bumper, bumpTime DateTime, isBumpWarning (if bump has not happened yet but will happen if person accepts) 
CREATE TABLE Bumps (
    bump_uuid uuid DEFAULT UUID_GENERATE_V4(),
    group_uuid uuid NOT NULL,
    bumped_id varchar NOT NULL,
    bumper_id varchar NOT NULL,
    bump_time timestamp,
    is_bump_warning boolean NOT NULL DEFAULT false,
    PRIMARY KEY (bump_uuid),
    FOREIGN KEY (group_uuid) REFERENCES Groups(group_uuid),
    FOREIGN KEY (bumped_id) REFERENCES Users(id),
    FOREIGN KEY (bumper_id) REFERENCES Users(id)
);
