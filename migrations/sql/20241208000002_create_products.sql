-- +goose Up
CREATE TABLE app_products (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES app_users(id),
    product_name VARCHAR(255) NOT NULL,
    product_description TEXT,
    product_price DECIMAL(10,2) NOT NULL,
    product_images TEXT[],
    compressed_product_images TEXT[],
    processing_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_user_id ON app_products(user_id);
CREATE INDEX idx_products_price ON app_products(product_price);
CREATE INDEX idx_products_name ON app_products(product_name);

-- +goose Down
DROP TABLE IF EXISTS app_products;