CREATE TABLE IF NOT EXISTS patients (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  gender text NOT NULL,
  age float(10) NOT NULL,
  contact integer NOT NULL,
  address text NOT NULL,
  medical_history text NOT NULL,
  insurance_info text NOT NULL,
  last_visit time NOT NULL,
  version integer NOT NULL DEFAULT 1,
  doctor_id bigint NOT NULL
);
