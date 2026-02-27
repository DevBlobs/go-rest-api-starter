CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    arrival_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_items_created_at ON items(created_at DESC);
CREATE INDEX idx_items_arrival_date ON items(arrival_date);
