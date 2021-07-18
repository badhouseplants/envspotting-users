DO $$ 
  BEGIN
    BEGIN
      ALTER TABLE users ADD COLUMN id TEXT PRIMARY KEY;
      ALTER TABLE users ADD COLUMN username TEXT UNIQUE;
      ALTER TABLE users ADD COLUMN password TEXT;
      ALTER TABLE users ADD COLUMN applications TEXT[];
      ALTER TABLE users ADD COLUMN gitlab_token TEXT;
    EXCEPTION
      WHEN duplicate_column THEN RAISE NOTICE 'column already exists in users.';
    END;
  END;
$$;
