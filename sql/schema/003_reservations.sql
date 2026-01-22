-- +goose Up
CREATE TABLE reservations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    room_id BIGINT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(30) NOT NULL,
    gcal_event_id VARCHAR(150),

    CONSTRAINT fk_reservation_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_reservation_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT unique_room_time UNIQUE (room_id, start_time, end_time)
);

CREATE INDEX idx_room_time ON reservations (room_id, start_time, end_time);

-- +goose Down
DROP TABLE IF EXISTS reservations;