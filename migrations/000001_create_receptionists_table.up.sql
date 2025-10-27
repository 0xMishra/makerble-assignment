CREATE TABLE IF NOT EXISTS receptionists (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  email citext UNIQUE NOT NULL,
  password_hash bytea NOT NULL,
  version integer NOT NULL DEFAULT 1,
  shift_start time NOT NULL,
  shift_end time NOT NULL,
);
