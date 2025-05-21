CREATE TABLE users (
                       id BIGINT PRIMARY KEY AUTO_INCREMENT,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password VARCHAR(255) NOT NULL
);

-- Books table
CREATE TABLE books (
                       id BIGINT PRIMARY KEY AUTO_INCREMENT,
                       user_id BIGINT NOT NULL,
                       title VARCHAR(255),
                       author VARCHAR(255),
                       total_pages INT,
                       status ENUM('reading', 'completed') DEFAULT 'reading',
                       FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Progresses table
CREATE TABLE progresses (
                            id BIGINT PRIMARY KEY AUTO_INCREMENT,
                            book_id BIGINT NOT NULL,
                            from_page INT NOT NULL,
                            until_page INT NOT NULL,
                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                            FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
);

-- Goals table
CREATE TABLE goals (
                       id BIGINT PRIMARY KEY AUTO_INCREMENT,
                       book_id BIGINT NOT NULL,
                       name VARCHAR(255),
                       target_page INT,
                       finished BOOLEAN DEFAULT FALSE,
                       expired_at DATETIME,
                       FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
);
