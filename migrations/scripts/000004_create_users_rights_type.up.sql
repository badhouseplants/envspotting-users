DO $$ 
  BEGIN
    CREATE TYPE user_rights AS ENUM ('READ', 'WRITE', 'DELETE');
  EXCEPTION
    WHEN duplicate_object THEN null;
  END 
$$;
