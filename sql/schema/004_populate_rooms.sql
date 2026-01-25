-- +goose Up
INSERT INTO rooms (name)
VALUES ('big'), ('small');

-- +goose Down
DELETE FROM rooms WHERE name IN ('big', 'small');