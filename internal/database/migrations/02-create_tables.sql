CREATE TABLE users (
    id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
    name VARCHAR,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE teams (
     id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
     name VARCHAR UNIQUE
);

CREATE TABLE team_members (
     id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
     team_id VARCHAR,
     user_id VARCHAR,
     is_admin BOOLEAN DEFAULT FALSE,
     FOREIGN KEY (team_id) REFERENCES teams(id),
     FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_team_members_team ON team_members(team_id);
CREATE INDEX idx_team_members_user ON team_members(user_id);

INSERT INTO users (name) VALUES ('Alice'), ('Bob');

CREATE TABLE pull_request_statuses(
    id VARCHAR PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL
);

INSERT INTO pull_request_statuses(id, name) VALUES ('1', 'OPEN'), ('2', 'MERGED');

CREATE TABLE pull_requests (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    author_id VARCHAR NOT NULL,
    status VARCHAR NOT NULL DEFAULT '1',
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    merged_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id),
    FOREIGN KEY (status) REFERENCES pull_request_statuses(id)
);

CREATE INDEX idx_pull_requests_author ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);
CREATE INDEX idx_pull_requests_created ON pull_requests(created_at);
CREATE INDEX idx_pull_requests_author_status ON pull_requests(author_id, status);

CREATE TABLE pull_request_assigned_reviewers (
    id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
    pull_request_id VARCHAR NOT NULL,
    user_id VARCHAR NOT NULL,
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_reviewers_pull_request ON pull_request_assigned_reviewers(pull_request_id);
CREATE INDEX idx_reviewers_user ON pull_request_assigned_reviewers(user_id);
CREATE UNIQUE INDEX idx_unique_reviewer_assignment
    ON pull_request_assigned_reviewers(pull_request_id, user_id);