-- Database user
DROP USER IF EXISTS 'ledger_user'@'localhost';
CREATE USER 'ledger_user'@'localhost' IDENTIFIED BY 'Auth123';

DROP DATABASE IF EXISTS ledger_db;
CREATE DATABASE ledger_db;

GRANT ALL PRIVILEGES ON ledger_db.* TO 'ledger_user'@'localhost';

USE ledger_db;

CREATE TABLE ledger (
    order_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    operation VARCHAR(255) NOT NULL,
    transaction_time VARCHAR(255) NOT NULL,

    INDEX(order_id)
);
