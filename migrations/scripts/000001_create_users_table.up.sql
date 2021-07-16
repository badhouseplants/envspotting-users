CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE,
		password TEXT,
    -- List of application in the sidebar
    applications TEXT[]
    -- Tokens
    gitlab_token TEXT
);