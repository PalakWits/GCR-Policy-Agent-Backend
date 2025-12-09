-- +migrate Up
ALTER TABLE baps ADD COLUMN first_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- +migrate Down
ALTER TABLE baps DROP COLUMN first_seen_at;
