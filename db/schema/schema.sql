CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    owner varchar NOT NULL,
    balance bigint NOT NULL,
    currency varchar NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now())
);


CREATE TABLE entries (
    id BIGSERIAL PRIMARY KEY,
    account_id bigint NOT NULL,
    amount bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    form_account_id bigint NOT NULL,
    to_account_id bigint NOT NULL,
    amount bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE users (
    username varchar PRIMARY KEY,
    hashed_password varchar NOT NULL,
    full_name varchar NOT NULL,
    email varchar NOT NULL,
    password_changed_at timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
    created_at timestamptz NOT NULL DEFAULT (now())
);