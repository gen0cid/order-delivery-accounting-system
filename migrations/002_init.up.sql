CREATE table reviews(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL UNIQUE,
    courier_id UUID NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

-- Установка внешних ключей
    CONSTRAINT fk_reviews_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_reviews_courier FOREIGN KEY (courier_id) REFERENCES couriers(id) ON DELETE CASCADE
);

ALTER TABLE couriers
ADD COLUMN rating DECIMAL(3, 2) DEFAULT 0.00,
ADD COLUMN total_reviews INTEGER DEFAULT 0;