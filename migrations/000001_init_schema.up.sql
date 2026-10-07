CREATE TABLE IF NOT EXISTS fleets (
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    current_wasm_hash BYTEA,
    member_count BIGINT NOT NULL DEFAULT 0,
    first_seen_ledger BIGINT NOT NULL,
    last_seen_ledger BIGINT NOT NULL,
    last_indexed_ledger BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (owner_address, tag)
);

CREATE TABLE IF NOT EXISTS fleet_members (
    contract_id TEXT PRIMARY KEY,
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    wasm_hash BYTEA NOT NULL,
    first_seen_ledger BIGINT NOT NULL,
    last_seen_ledger BIGINT NOT NULL,
    last_verified_ledger BIGINT,
    active BOOLEAN NOT NULL DEFAULT TRUE,

    FOREIGN KEY (owner_address, tag)
        REFERENCES fleets(owner_address, tag)
);

CREATE TABLE IF NOT EXISTS fleet_releases (
    id BIGSERIAL PRIMARY KEY,
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    old_wasm_hash BYTEA,
    new_wasm_hash BYTEA NOT NULL,
    ledger BIGINT NOT NULL,
    tx_hash TEXT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS fleet_verifications (
    id BIGSERIAL PRIMARY KEY,
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    expected_wasm_hash BYTEA NOT NULL,
    total_members BIGINT NOT NULL,
    matching_members BIGINT NOT NULL,
    mismatching_members BIGINT NOT NULL,
    missing_members BIGINT NOT NULL,
    status TEXT NOT NULL,
    indexed_through BIGINT NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS indexer_checkpoints (
    stream TEXT PRIMARY KEY,
    ledger BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fleet_members_fleet
    ON fleet_members(owner_address, tag);

CREATE INDEX IF NOT EXISTS idx_fleet_members_wasm
    ON fleet_members(wasm_hash);

CREATE INDEX IF NOT EXISTS idx_fleet_releases_fleet_ledger
    ON fleet_releases(owner_address, tag, ledger);

CREATE INDEX IF NOT EXISTS idx_verifications_fleet_time
    ON fleet_verifications(owner_address, tag, verified_at);
