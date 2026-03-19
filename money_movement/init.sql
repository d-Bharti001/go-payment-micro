DROP USER IF EXISTS 'money_movement_user'@'localhost';
CREATE USER 'money_movement_user'@'localhost' IDENTIFIED BY 'Auth123';

DROP DATABASE IF EXISTS money_movement;
CREATE DATABASE money_movement;

GRANT ALL PRIVILEGES ON money_movement.* TO 'money_movement_user'@'localhost';

USE money_movement;

CREATE TABLE wallets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    wallet_type VARCHAR(255) NOT NULL
);

-- one wallet can have multiple associated accounts
CREATE TABLE accounts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    paise INT NOT NULL DEFAULT 0,
    account_type VARCHAR(255) NOT NULL,
    wallet_id INT NOT NULL REFERENCES wallets(id)
);

CREATE TABLE transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pid VARCHAR(255) NOT NULL,
    src_user_id VARCHAR(255) NOT NULL,
    dst_user_id VARCHAR(255) NOT NULL,
    src_wallet_id INT NOT NULL,
    dst_wallet_id INT NOT NULL,
    src_account_id INT NOT NULL,
    dst_account_id INT NOT NULL,
    src_account_type VARCHAR(255) NOT NULL,
    dst_account_type VARCHAR(255) NOT NULL,
    final_dst_merchant_wallet_id INT,
    amount INT NOT NULL,

    INDEX(pid)
);

-- merchant and customer wallets
INSERT INTO wallets
    (id, user_id, wallet_type)
VALUES
    (1, 'georgio@email.com', 'CUSTOMER'),
    (2, 'merchant@email.com', 'MERCHANT');

-- customer accounts
INSERT INTO accounts
    (paise, account_type, wallet_id)
VALUES
    (500000, 'DEFAULT', 1),
    (0, 'PAYMENT', 1);

-- merchant account
INSERT INTO accounts
    (paise, account_type, wallet_id)
VALUES
    (0, 'INCOMING', 2);
