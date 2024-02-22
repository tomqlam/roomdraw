-- - Join Suite/Room Request Table
--Columns:
--	RequestUID, Type of request (join Lock-pull, join pull, join room), request recipient USER_ID, SuiteUID, RoomUID, Request Creator, ApprovalCount (only applicable to lockpull), ApprovalGroupID (only applicable to lockpull), JoinRequestReply (accept, deny, pending), DateTime created

CREATE TABLE JoinRequests (
    request_uuid uuid NOT NULL,
    request_type varchar NOT NULL,
    request_recipient varchar NOT NULL,
    suite_uuid uuid,
    room_uuid uuid NOT NULL,
    request_creator_id varchar NOT NULL,
    approval_count int,
    approval_group_uuid uuid,
    join_request_reply varchar,
    created_at timestamp NOT NULL DEFAULT NOW(),
    PRIMARY KEY (request_uuid),
    FOREIGN KEY (request_recipient) REFERENCES Users(id),
    FOREIGN KEY (request_creator) REFERENCES Users(id),
    FOREIGN KEY (suite_uuid) REFERENCES Suites(suite_uuid),
    FOREIGN KEY (room_uuid) REFERENCES Rooms(room_uuid),
    FOREIGN KEY (approval_group_uuid) REFERENCES Groups(group_uuid)
);