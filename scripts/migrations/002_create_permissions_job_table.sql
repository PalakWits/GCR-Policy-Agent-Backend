-- +migrate Up
CREATE TABLE
    IF NOT EXISTS permissions_jobs (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        bap_id VARCHAR(255) NOT NULL,
        status VARCHAR(255) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW (),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW ()
    );

-- +migrate Down
DROP TABLE IF EXISTS permissions_jobs;
