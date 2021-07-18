DO $$
  BEGIN
    BEGIN
      ALTER TABLE rights ADD CONSTRAINT unique_right UNIQUE (user_id, application_id);
    EXCEPTION
      WHEN duplicate_object THEN RAISE NOTICE 'Table constraint unique_right already exists';
    END;
  END 
$$;