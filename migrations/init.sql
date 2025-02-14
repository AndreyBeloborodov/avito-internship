-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(255) UNIQUE NOT NULL,
    balance INT NOT NULL DEFAULT 1000
    );

-- Создаем таблицу мерча
CREATE TABLE IF NOT EXISTS merches (
                                     id SERIAL PRIMARY KEY,
                                     name VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL CHECK (price > 0)
    );

-- Заполняем таблицу мерча начальными данными
INSERT INTO merches (name, price) VALUES
                                    ('t-shirt', 80),
                                    ('cup', 20),
                                    ('book', 50),
                                    ('pen', 10),
                                    ('powerbank', 200),
                                    ('hoody', 300),
                                    ('umbrella', 200),
                                    ('socks', 10),
                                    ('wallet', 50),
                                    ('pink-hoody', 500)
    ON CONFLICT (name) DO NOTHING;

-- Создаем таблицу истории транзакций монет
CREATE TABLE IF NOT EXISTS transactions (
                                            id SERIAL PRIMARY KEY,
                                            sender_id INT REFERENCES users(id),
    receiver_id INT REFERENCES users(id),
    amount INT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Создаем таблицу покупок
CREATE TABLE IF NOT EXISTS purchases (
                                         id SERIAL PRIMARY KEY,
                                         user_id INT REFERENCES users(id),
    merch_id INT REFERENCES merch(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
