CREATE TABLE products (
    product_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price BIGINT NOT NULL,
    quantity INT NOT NULL
);

CREATE INDEX idx_product_name ON products (name);

CREATE TABLE transactions (
    transaction_id VARCHAR(36) PRIMARY KEY,
    total_price BIGINT NOT NULL
);

CREATE TABLE transaction_details (
    transaction_detail_id VARCHAR(36) PRIMARY KEY,
    sub_total_price BIGINT NOT NULL,
    price BIGINT NOT NULL,
    quantity INT NOT NULL,
    transaction_id VARCHAR(36),
    product_id VARCHAR(36),
    CONSTRAINT fk_product_transaction_details FOREIGN KEY (product_id) REFERENCES products (product_id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_transaction_transaction_details FOREIGN KEY (transaction_id) REFERENCES transactions (transaction_id) ON DELETE CASCADE ON UPDATE CASCADE
);
