CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(50) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255) NOT NULL,
    delivery_service VARCHAR(50) NOT NULL,
    shardkey VARCHAR(10) NOT NULL,
    sm_id INTEGER NOT NULL,
    date_created TIMESTAMP WITH TIME ZONE NOT NULL,
    oof_shard VARCHAR(10) NOT NULL
);

CREATE TABLE IF NOT EXISTS deliveries (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255) NOT NULL,
    request_id VARCHAR(255),
    currency VARCHAR(10) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    amount INTEGER NOT NULL,
    payment_dt BIGINT NOT NULL,
    bank VARCHAR(50) NOT NULL,
    delivery_cost INTEGER NOT NULL,
    goods_total INTEGER NOT NULL,
    custom_fee INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    rid VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sale INTEGER NOT NULL,
    size VARCHAR(10) NOT NULL,
    total_price INTEGER NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR(255) NOT NULL,
    status INTEGER NOT NULL
);