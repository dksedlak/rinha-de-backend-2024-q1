CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS balances(
    client_id INTEGER NOT NULL,
    client_limit BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    last_transactions JSON NOT NULL,
    last_commit UUID NOT NULL,
    CONSTRAINT pk_balance PRIMARY KEY(client_id)
);

INSERT INTO balances(client_id, client_limit, amount, last_transactions, last_commit) VALUES(1,   100000, 0, '[]', uuid_generate_v4());
INSERT INTO balances(client_id, client_limit, amount, last_transactions, last_commit) VALUES(2,    80000, 0, '[]', uuid_generate_v4());
INSERT INTO balances(client_id, client_limit, amount, last_transactions, last_commit) VALUES(3,  1000000, 0, '[]', uuid_generate_v4());
INSERT INTO balances(client_id, client_limit, amount, last_transactions, last_commit) VALUES(4, 10000000, 0, '[]', uuid_generate_v4());
INSERT INTO balances(client_id, client_limit, amount, last_transactions, last_commit) VALUES(5,   500000, 0, '[]', uuid_generate_v4());