CREATE TYPE role_enum AS ENUM ('doctor', 'receptionist');

CREATE TABLE IF NOT EXISTS tokens (
  hash text PRIMARY KEY,
  email citext NOT NULL,
  role role_enum NOT NULL,
  expiry timestamp(0) with time zone NOT NULL,
  scope text NOT NULL
);
