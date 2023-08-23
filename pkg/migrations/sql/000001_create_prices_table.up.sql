CREATE TABLE IF NOT EXISTS prices (
    id VARCHAR(255) PRIMARY KEY,
    price DECIMAL(20, 10),
    expiration_date DATETIME
);
