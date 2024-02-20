CREATE TABLE IF NOT EXISTS balances(
    client_id INTEGER NOT NULL,
    client_limit BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    last_transactions JSON NOT NULL,
    CONSTRAINT pk_balance PRIMARY KEY(client_id)
);

INSERT INTO balances(client_id, client_limit, amount, last_transactions) VALUES(1,   100000, 0, '[]');
INSERT INTO balances(client_id, client_limit, amount, last_transactions) VALUES(2,    80000, 0, '[]');
INSERT INTO balances(client_id, client_limit, amount, last_transactions) VALUES(3,  1000000, 0, '[]');
INSERT INTO balances(client_id, client_limit, amount, last_transactions) VALUES(4, 10000000, 0, '[]');
INSERT INTO balances(client_id, client_limit, amount, last_transactions) VALUES(5,   500000, 0, '[]');