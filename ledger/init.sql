-- Database user
CREATE USER 'ledger_user'@'%' IDENTIFIED BY 'Auth123';

CREATE DATABASE ledger_db;

GRANT ALL PRIVILEGES ON ledger_db.* TO 'ledger_user'@'%';

USE ledger_db;

CREATE TABLE ledger (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    operation VARCHAR(255) NOT NULL,
    transaction_time VARCHAR(255) NOT NULL,

    INDEX(order_id)
);
