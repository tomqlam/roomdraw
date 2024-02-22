-- - Approve Lock Pull Table
-- 	ApprovalUID, ApprovalGroupID (doesn’t change between approvers), requester USER_ID (must be person with highest number in n-1 filled suite), candidate USER_ID, approver USER_ID (new record for each approver), SuiteID, DateTime created, status(no response, rejected, approved), consensusSuccess(success, fail, inprogress), replyReceived DateTime
CREATE TABLE ApproveLockPull (
    approval_uuid uuid NOT NULL,
    approval_group_uuid uuid NOT NULL,
    requester_id varchar NOT NULL,
    candidate_id varchar NOT NULL,
    approver_id varchar NOT NULL,
    suite_uuid uuid NOT NULL,
    created_at_datetime timestamp NOT NULL DEFAULT NOW(),
    response varchar NOT NULL,
    consensus_success varchar NOT NULL,
    reply_received_datetime timestamp,
    PRIMARY KEY (approval_uuid),
    FOREIGN KEY (requester_id) REFERENCES Users(id),
    FOREIGN KEY (candidate_id) REFERENCES Users(id),
    FOREIGN KEY (approver_id) REFERENCES Users(id),
    FOREIGN KEY (suite_uuid) REFERENCES Suites(suite_uuid)
);