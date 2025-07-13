ALTER TABLE reservations
    ADD CONSTRAINT unique_room_time UNIQUE (room_id, start_time, end_time);
