-- Create blockchains table first since it's referenced by other tables
CREATE TABLE IF NOT EXISTS blockchains (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    network VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT false,
    logo_url TEXT,
    explorer_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create tokens table with foreign key to blockchains
CREATE TABLE IF NOT EXISTS tokens (
    id UUID PRIMARY KEY,
    token_id VARCHAR(255),
    type VARCHAR(50),
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    blockchain_id UUID NOT NULL REFERENCES blockchains(id),
    logo_url TEXT,
    yahoo_symbol VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create wallets_meta table with foreign key to blockchains
CREATE TABLE IF NOT EXISTS wallet_meta (
    id UUID PRIMARY KEY,
    main_wallet VARCHAR(255) NOT NULL,
    blockchain_id UUID NOT NULL REFERENCES blockchains(id),
    last_index INTEGER
);

-- Create wallets table with foreign key to wallets_meta
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    meta_id UUID NOT NULL REFERENCES wallet_meta(id),
    index INTEGER NOT NULL,
    derivation_path VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    tx_id VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    token_id UUID NOT NULL REFERENCES tokens(id),
    to_address VARCHAR(255) NOT NULL,
    amount DECIMAL(38, 18) NOT NULL,
    fee DECIMAL(38, 18) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blockchain_parse_errors (
    blockchain_id UUID NOT NULL PRIMARY KEY REFERENCES blockchains(id),
    block_number BIGINT,
    retries INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blockchain_parses (
    blockchain_id UUID NOT NULL PRIMARY KEY REFERENCES blockchains(id),
    last_block_number BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sweeping_sessions (
    id UUID PRIMARY KEY,
    wallet_meta_id UUID NOT NULL REFERENCES wallet_meta(id),
    token_id UUID NOT NULL REFERENCES tokens(id),
    min_amount_threshold DECIMAL(38, 18) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    meta JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sweeps (
    transaction_id UUID NOT NULL PRIMARY KEY REFERENCES transactions(id),
    sweeping_session_id UUID NOT NULL REFERENCES sweeping_sessions(id),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS utxos (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    vout INTEGER NOT NULL,
    scriptpubkeybytes BYTEA NOT NULL,
    address VARCHAR(255) NOT NULL,
    amount DECIMAL(38, 18) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tokens_blockchain_id ON tokens(blockchain_id);
CREATE INDEX idx_tokens_symbol ON tokens(symbol);
CREATE INDEX idx_tokens_token_id ON tokens(token_id);
CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_wallets_meta_blockchain_main_wallet ON wallet_meta(blockchain_id, main_wallet);