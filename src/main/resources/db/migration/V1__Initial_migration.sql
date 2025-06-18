-- USERS
CREATE TABLE users (
                       id BIGINT PRIMARY KEY AUTO_INCREMENT,
                       email VARCHAR(100) NOT NULL UNIQUE,
                       name VARCHAR(100),
                       role VARCHAR(20) NOT NULL
);

-- ROOMS
CREATE TABLE rooms (
                       id BIGINT PRIMARY KEY AUTO_INCREMENT,
                       name VARCHAR(30) NOT NULL UNIQUE
);

-- RESERVATIONS
CREATE TABLE reservations (
                              id BIGINT PRIMARY KEY AUTO_INCREMENT,
                              user_id BIGINT NOT NULL,
                              room_id BIGINT NOT NULL,
                              start_time DATETIME NOT NULL,
                              end_time DATETIME NOT NULL,
                              status VARCHAR(30) NOT NULL,

                              CONSTRAINT fk_reservation_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                              CONSTRAINT fk_reservation_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,

                              INDEX idx_room_time (room_id, start_time, end_time)
);
