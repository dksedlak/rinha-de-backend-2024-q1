CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS clients(
    id INTEGER NOT NULL,
    CONSTRAINT pk_clients PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS balances(
    client_id INTEGER NOT NULL,
    balance_limit BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    last_commit UUID NOT NULL,
    CONSTRAINT pk_balances PRIMARY KEY(client_id),
    CONSTRAINT fk_balances_clients FOREIGN KEY (client_id) 
        REFERENCES clients(id)
);

CREATE TABLE IF NOT EXISTS transactions (
    client_id INTEGER NOT NULL,
    last_transactions JSON,
    last_commit UUID NOT NULL,
    CONSTRAINT pk_transactions PRIMARY KEY(client_id),
    CONSTRAINT fk_transactions_clients FOREIGN KEY (client_id) REFERENCES clients(id)
);

INSERT INTO clients VALUES(1);
INSERT INTO clients VALUES(2);
INSERT INTO clients VALUES(3);
INSERT INTO clients VALUES(4);
INSERT INTO clients VALUES(5);

INSERT INTO balances(client_id, balance_limit, amount, last_commit) VALUES(1,   100000, 0, uuid_generate_v4());
INSERT INTO balances(client_id, balance_limit, amount, last_commit) VALUES(2,    80000, 0, uuid_generate_v4());
INSERT INTO balances(client_id, balance_limit, amount, last_commit) VALUES(3,  1000000, 0, uuid_generate_v4());
INSERT INTO balances(client_id, balance_limit, amount, last_commit) VALUES(4, 10000000, 0, uuid_generate_v4());
INSERT INTO balances(client_id, balance_limit, amount, last_commit) VALUES(5,   500000, 0, uuid_generate_v4());
