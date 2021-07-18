DO $$ 
  BEGIN
    BEGIN
      ALTER TABLE rights ADD COLUMN id TEXT PRIMARY KEY;
      ALTER TABLE rights ADD COLUMN user_id TEXT REFERENCES users(id) ON DELETE CASCADE;
      ALTER TABLE rights ADD COLUMN application_id TEXT;
      ALTER TABLE rights ADD COLUMN access_right user_rights NOT NULL;
    EXCEPTION
      WHEN duplicate_column THEN RAISE NOTICE 'column already exists in users.';
    END;
  END;
$$;
