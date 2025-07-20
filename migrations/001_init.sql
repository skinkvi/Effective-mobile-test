-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    price INT NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
);
---- create above / drop below ----
DROP TABLE IF EXISTS subscriptions;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
