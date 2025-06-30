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