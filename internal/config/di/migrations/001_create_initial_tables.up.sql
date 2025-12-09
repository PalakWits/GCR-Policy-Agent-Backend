CREATE TABLE
    IF NOT EXISTS sellers (
        seller_id TEXT,
        domain TEXT,
        status TEXT,
        type TEXT,
        subscriber_url TEXT,
        country TEXT,
        city TEXT,
        valid_from TIMESTAMPTZ,
        valid_until TIMESTAMPTZ,
        active BOOLEAN,
        registry_raw JSONB,
        last_seen_in_reg TIMESTAMPTZ,
        created_at TIMESTAMPTZ,
        updated_at TIMESTAMPTZ,
        PRIMARY KEY (seller_id, domain)
    );

CREATE TABLE
    IF NOT EXISTS seller_catalog_state (
        seller_id TEXT,
        domain TEXT,
        status TEXT,
        last_pull_at TIMESTAMPTZ,
        last_success_at TIMESTAMPTZ,
        last_error TEXT,
        sync_version BIGINT,
        updated_at TIMESTAMPTZ,
        PRIMARY KEY (seller_id, domain)
    );

-- +migrate Down
DROP TABLE IF EXISTS sellers;

DROP TABLE IF EXISTS seller_catalog_state;
