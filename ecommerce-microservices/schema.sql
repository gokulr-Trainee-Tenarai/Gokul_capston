-- Products Table
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO products (name, price) VALUES 
('Mechanical Keyboard', 89.99),
('Wireless Gaming Mouse', 49.50),
('4K Ultra HD Monitor', 329.00)
ON CONFLICT DO NOTHING;

-- Orders Table
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id INT NOT NULL DEFAULT 101,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    total NUMERIC(10, 2) NOT NULL CHECK (total >= 0),
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO orders (customer_id, product_id, quantity, total, status) VALUES 
(101, 1, 2, 179.98, 'CREATED'),
(102, 2, 1, 49.50, 'CONFIRMED')
ON CONFLICT DO NOTHING;
