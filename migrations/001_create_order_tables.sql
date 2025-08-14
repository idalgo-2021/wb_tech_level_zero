CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Заказы
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_uid TEXT UNIQUE NOT NULL,
    track_number TEXT NOT NULL,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INTEGER,
    date_created TIMESTAMPTZ,
    oof_shard TEXT
);

-- Доставка
CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

-- Оплата
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount NUMERIC(12,2),
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost NUMERIC(12,2),
    goods_total NUMERIC(12,2),
    custom_fee NUMERIC(12,2)
);

-- Номенклатура заказа
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number TEXT,
    price NUMERIC(12,2),
    rid TEXT,
    name TEXT,
    sale INTEGER,
    size TEXT,
    total_price NUMERIC(12,2),
    nm_id BIGINT,
    brand TEXT,
    status INTEGER
);


CREATE INDEX idx_orders_order_uid ON orders(order_uid);
CREATE INDEX idx_items_order_id ON items(order_id);
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_deliveries_order_id ON deliveries(order_id);
