CREATE TABLE users (
    id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
    name VARCHAR,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE teams (
     id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
     name VARCHAR
);

CREATE TABLE team_members (
     id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
     team_id VARCHAR,
     user_id VARCHAR,
     is_active BOOLEAN DEFAULT TRUE,
     is_admin BOOLEAN DEFAULT FALSE,
     FOREIGN KEY (team_id) REFERENCES teams(id),
     FOREIGN KEY (user_id) REFERENCES users(id)
);


INSERT INTO users (name) VALUES ('Alice'), ('Bob')